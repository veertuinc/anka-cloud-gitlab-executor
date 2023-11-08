package ankacloud

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"veertu.com/anka-cloud-gitlab-executor/internal/gitlab"
	"veertu.com/anka-cloud-gitlab-executor/internal/log"
)

type APIClient struct {
	ControllerURL string
	HttpClient    *http.Client
}

func (c *APIClient) parse(body []byte) (response, error) {
	var r response
	err := json.Unmarshal(body, &r)
	if err != nil {
		return r, fmt.Errorf("failed to decode response body %+v: %w", string(body), err)
	}

	if r.Status != statusOK {
		return r, fmt.Errorf(r.Message)
	}

	return r, nil
}

func toQueryParams(params map[string]string) url.Values {
	query := url.Values{}
	for k, v := range params {
		query.Add(k, v)
	}
	return query
}

func (c *APIClient) Post(ctx context.Context, endpoint string, payload interface{}) ([]byte, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to parse POST request body %+v: %w", payload, err)
	}

	endpointUrl := fmt.Sprintf("%s%s", c.ControllerURL, endpoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpointUrl, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create POST request to %q with body %+v: %w", endpointUrl, payload, err)
	}

	r, err := c.HttpClient.Do(req)
	if err != nil {
		if e, ok := err.(*url.Error); ok && e.Timeout() {
			return nil, gitlab.TransientError(fmt.Errorf("failed to send POST request to %s with payload %+v: %w", endpointUrl, payload, e))
		}
		return nil, fmt.Errorf("failed to send POST request to %s with body %+v: %w", endpointUrl, payload, err)
	}
	defer r.Body.Close()

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	baseResponse, err := c.parse(bodyBytes)
	if err != nil {
		return nil, fmt.Errorf("status code: %d, error: %w", r.StatusCode, err)
	}

	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d, error: %s", r.StatusCode, baseResponse.Message)
	}

	log.Debugf("POST request sent to %s\nRaw payload: %+v\nResponse status code: %d\nRaw body: %+v\n", endpoint, payload, r.StatusCode, string(bodyBytes))
	return bodyBytes, nil
}

func (c *APIClient) Delete(ctx context.Context, endpoint string, payload interface{}) ([]byte, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DELETE request body %+v: %w", payload, err)
	}

	endpointUrl := fmt.Sprintf("%s%s", c.ControllerURL, endpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpointUrl, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create DELETE request to %q with payload %+v: %w", endpointUrl, payload, err)
	}
	req.Header.Set("Content-Type", "application/json")
	r, err := c.HttpClient.Do(req)
	if err != nil {
		if e, ok := err.(*url.Error); ok && e.Timeout() {
			return nil, gitlab.TransientError(fmt.Errorf("failed to send DELETE request to %s with payload %+v: %w", endpointUrl, payload, e))
		}
		return nil, fmt.Errorf("failed to send DELETE request to %s with payload %+v: %w", endpoint, payload, err)
	}
	defer r.Body.Close()

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	baseResponse, err := c.parse(bodyBytes)
	if err != nil {
		return nil, fmt.Errorf("status code: %d, error: %w", r.StatusCode, err)
	}

	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d, error: %s", r.StatusCode, baseResponse.Message)
	}

	log.Debugf("DELETE request sent to %s\n Raw payload: %+v\nResponse status code: %d\nRaw body: %+v\n", endpoint, payload, r.StatusCode, string(bodyBytes))
	return bodyBytes, nil
}

func (c *APIClient) Get(ctx context.Context, endpoint string, queryParams map[string]string) ([]byte, error) {
	if len(queryParams) > 0 {
		params := toQueryParams(queryParams)
		endpoint = fmt.Sprintf("%s?%s", endpoint, params.Encode())
	}
	endpointUrl := fmt.Sprintf("%s%s", c.ControllerURL, endpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpointUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create GET request to %q: %w", endpointUrl, err)
	}

	r, err := c.HttpClient.Do(req)
	if err != nil {
		if e, ok := err.(*url.Error); ok && e.Timeout() {
			return nil, gitlab.TransientError(fmt.Errorf("failed to send GET request to %s: %w", endpointUrl, e))
		}
		return nil, fmt.Errorf("failed to send GET request to %s: %w", endpointUrl, err)
	}
	defer r.Body.Close()

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	baseResponse, err := c.parse(bodyBytes)
	if err != nil {
		return nil, fmt.Errorf("status code: %d, error: %w", r.StatusCode, err)
	}

	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d, error: %s", r.StatusCode, baseResponse.Message)
	}

	log.Debugf("GET request to %s\nResponse status code: %d\nRaw body: %+v\n", endpoint, r.StatusCode, string(bodyBytes))

	return bodyBytes, nil
}

type APIClientConfig struct {
	BaseURL             string
	IsTLS               bool
	CaCertPath          string
	ClientCertPath      string
	ClientCertKeyPath   string
	SkipTLSVerify       bool
	MaxIdleConnsPerHost int
	RequestTimeout      time.Duration
}

func (c *APIClientConfig) certAuthEnabled() bool {
	return c.ClientCertKeyPath != "" && c.ClientCertPath != ""
}

const (
	defaultMaxIdleConnsPerHost = 20
	defaultRequestTimeout      = 10 * time.Second
)

func NewAPIClient(config APIClientConfig) (*APIClient, error) {
	httpClient := &http.Client{
		Timeout: defaultRequestTimeout,
	}

	if config.RequestTimeout > 0 {
		httpClient.Timeout = config.RequestTimeout
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConnsPerHost = defaultMaxIdleConnsPerHost

	if config.MaxIdleConnsPerHost > 0 {
		transport.MaxIdleConnsPerHost = config.MaxIdleConnsPerHost
	}

	if config.IsTLS {
		tlsConfig, err := configureTLS(config)
		if err != nil {
			return nil, fmt.Errorf("failed to configure TLS %+v: %w", config, err)
		}
		transport.TLSClientConfig = tlsConfig
	}

	httpClient.Transport = transport

	return &APIClient{
		ControllerURL: config.BaseURL,
		HttpClient:    httpClient,
	}, nil
}

func configureTLS(config APIClientConfig) (*tls.Config, error) {
	log.Println("Handling TLS configuration")

	tlsConfig := &tls.Config{}
	caCertPool, _ := x509.SystemCertPool()
	if caCertPool == nil {
		caCertPool = x509.NewCertPool()
	}
	tlsConfig.RootCAs = caCertPool

	if config.CaCertPath != "" {
		if err := appendRootCert(config.CaCertPath, caCertPool); err != nil {
			return nil, fmt.Errorf("failed to add CA cert from %q to pool: %w", config.CaCertPath, err)
		}
		log.Printf("Added CA cert at %q\n", config.CaCertPath)
	}

	if config.SkipTLSVerify {
		log.Println("Allowing to skip server host verification")
		tlsConfig.InsecureSkipVerify = true
	}

	if config.certAuthEnabled() {
		cert, err := tls.LoadX509KeyPair(config.ClientCertPath, config.ClientCertKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to process key pair (cert at %q, key at %q): %w", config.ClientCertPath, config.ClientCertKeyPath, err)
		}

		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return tlsConfig, nil
}

func appendRootCert(certFilePath string, caCertPool *x509.CertPool) error {
	cert, err := os.ReadFile(certFilePath)
	if err != nil {
		return fmt.Errorf("failed to read file at %q: %w", certFilePath, err)
	}
	ok := caCertPool.AppendCertsFromPEM(cert)
	if !ok {
		return fmt.Errorf("failed to add cert at %q to cert pool: %w", certFilePath, err)
	}
	return nil
}
