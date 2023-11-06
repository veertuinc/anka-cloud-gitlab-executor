package ankacloud

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"veertu.com/anka-cloud-gitlab-executor/internal/log"
)

const (
	statusOK = "OK"
)

type response struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Body    interface{} `json:"body,omitempty"`
}

type ClientConfig struct {
	ControllerURL string
	CACertPath    string
	SkipTLSVerify bool
}

type Client struct {
	config     ClientConfig
	httpClient *http.Client
}

func (c *Client) parse(body []byte) (response, error) {
	var r response
	err := json.Unmarshal(body, &r)
	if err != nil {
		return r, fmt.Errorf("unexpected response body structure: %s", string(body))
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

func (c *Client) Post(endpoint string, payload interface{}) ([]byte, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(payload)
	if err != nil {
		return nil, fmt.Errorf("failed parsing POST request body: %w", err)
	}

	r, err := c.httpClient.Post(fmt.Sprintf("%s%s", c.config.ControllerURL, endpoint), "application/json", &buf)
	if err != nil {
		return nil, fmt.Errorf("failed sending POST request: %w", err)
	}
	defer r.Body.Close()

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading response body: %w", err)
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

func (c *Client) Delete(endpoint string, payload interface{}) ([]byte, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(payload)
	if err != nil {
		return nil, fmt.Errorf("failed parsing POST request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s%s", c.config.ControllerURL, endpoint), &buf)
	if err != nil {
		return nil, fmt.Errorf("failed creating DELETE request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	r, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed sending DELETE request to %s with payload %+v: %w", endpoint, payload, err)
	}
	defer r.Body.Close()

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading response body: %w", err)
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

func (c *Client) Get(endpoint string, queryParams map[string]string) ([]byte, error) {
	if len(queryParams) > 0 {
		params := toQueryParams(queryParams)
		endpoint = fmt.Sprintf("%s?%s", endpoint, params.Encode())
	}
	r, err := c.httpClient.Get(fmt.Sprintf("%s%s", c.config.ControllerURL, endpoint))
	if err != nil {
		return nil, fmt.Errorf("failed sending GET request to %s: %w", endpoint, err)
	}
	defer r.Body.Close()

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading response body: %w", err)
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

func appendRootCert(certFilePath string, caCertPool *x509.CertPool) error {
	cert, err := os.ReadFile(certFilePath)
	if err != nil {
		return fmt.Errorf("error reading %q: %w", certFilePath, err)
	}
	ok := caCertPool.AppendCertsFromPEM(cert)
	if !ok {
		return fmt.Errorf("error adding cert from %q to cert pool: %w", certFilePath, err)
	}
	return nil
}

func NewClient(config ClientConfig) (*Client, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConnsPerHost = 50

	if strings.HasPrefix(config.ControllerURL, "https") {
		if err := configureTLS(transport, config); err != nil {
			return nil, fmt.Errorf("failed to configure TLS: %w", err)
		}
	}

	return &Client{
		config: config,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   10 * time.Second,
		},
	}, nil
}

func configureTLS(transport *http.Transport, config ClientConfig) error {
	log.Println("Handling HTTPS configuration")

	tlsConfig := transport.TLSClientConfig
	caCertPool, _ := x509.SystemCertPool()
	if caCertPool == nil {
		caCertPool = x509.NewCertPool()
	}
	tlsConfig.RootCAs = caCertPool

	if config.CACertPath != "" {
		if err := appendRootCert(config.CACertPath, caCertPool); err != nil {
			return err
		}
		log.Printf("Added CA cert from %s\n", config.CACertPath)
	}

	if config.SkipTLSVerify {
		log.Println("Allowing to skip server host verification")
		tlsConfig.InsecureSkipVerify = true
	}

	return nil
}
