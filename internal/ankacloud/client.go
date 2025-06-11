package ankacloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/veertuinc/anka-cloud-gitlab-executor/internal/gitlab"
	"github.com/veertuinc/anka-cloud-gitlab-executor/internal/log"
)

const (
	defaultMaxIdleConnsPerHost = 20
	defaultRequestTimeout      = 10 * time.Second
)

type APIClient struct {
	ControllerURL     string
	HttpClient        *http.Client
	CustomHttpHeaders map[string]string
}

func (c *APIClient) parse(body []byte) (response, error) {
	var r response
	err := json.Unmarshal(body, &r)
	if err != nil {
		return r, fmt.Errorf("failed to decode response body %+v: %w", string(body), err)
	}

	if r.Status != statusOK {
		return r, fmt.Errorf("%s", r.Message)
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

	for k, v := range c.CustomHttpHeaders {
		log.Debugf("Setting custom header %s: %s\n", k, v)
		req.Header.Set(k, v)
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

	for k, v := range c.CustomHttpHeaders {
		log.Debugf("Setting custom header %s: %s\n", k, v)
		req.Header.Set(k, v)
	}

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

	for k, v := range c.CustomHttpHeaders {
		log.Debugf("Setting custom header %s: %s\n", k, v)
		req.Header.Set(k, v)
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
	CustomHttpHeaders   map[string]string
}

func (c *APIClientConfig) certAuthEnabled() bool {
	return c.ClientCertKeyPath != "" && c.ClientCertPath != ""
}

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
		ControllerURL:     config.BaseURL,
		CustomHttpHeaders: config.CustomHttpHeaders,
		HttpClient:        httpClient,
	}, nil
}
