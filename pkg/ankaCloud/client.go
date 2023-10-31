package ankaCloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"veertu.com/anka-cloud-gitlab-executor/internal/log"
)

const (
	statusOK = "OK"
)

type baseResponse interface {
	GetStatus() string
	GetMessage() string
}

type response struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Body    interface{} `json:"body,omitempty"`
}

func (r response) GetStatus() string {
	return r.Status
}

func (r response) GetMessage() string {
	return r.Message
}

type ClientConfig struct {
	ControllerURL string
}

type Client struct {
	config     ClientConfig
	httpClient *http.Client
}

func parse(r *http.Response, response baseResponse) error {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed reading response body: %w", err)
	}

	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		return fmt.Errorf("unexpected response body structure: %s", string(bodyBytes))
	}

	if r.StatusCode != 200 {
		return fmt.Errorf("status code: %d, error: %s", r.StatusCode, response.GetMessage())
	}

	if response.GetStatus() != statusOK {
		return fmt.Errorf(response.GetMessage())
	}

	return nil
}

func toQueryParams(params map[string]string) url.Values {
	query := url.Values{}
	for k, v := range params {
		query.Add(k, v)
	}
	return query
}

func (c *Client) Post(endpoint string, payload interface{}, response baseResponse) error {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(payload)
	if err != nil {
		return fmt.Errorf("failed parsing POST request body: %w", err)
	}

	r, err := c.httpClient.Post(fmt.Sprintf("%s%s", c.config.ControllerURL, endpoint), "application/json", &buf)
	if err != nil {
		return fmt.Errorf("failed sending POST request: %w", err)
	}

	return parse(r, response)
}

func (c *Client) Delete(endpoint string, payload interface{}, response baseResponse) error {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(payload)
	if err != nil {
		return fmt.Errorf("failed parsing POST request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s%s", c.config.ControllerURL, endpoint), &buf)
	if err != nil {
		return fmt.Errorf("failed creating DELETE request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	r, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed sending DELETE request to %s with payload %+v: %w", endpoint, payload, err)
	}

	if err := parse(r, response); err != nil {
		return err
	}
	log.Debugf("POST request sent to %s with payload: %+v\nResponse status code: %d, Body: %+v\n", endpoint, payload, r.StatusCode, response)
	return nil
}

func (c *Client) Get(endpoint string, queryParams map[string]string, response baseResponse) error {
	if len(queryParams) > 0 {
		params := toQueryParams(queryParams)
		endpoint = fmt.Sprintf("%s?%s", endpoint, params.Encode())
	}
	r, err := c.httpClient.Get(fmt.Sprintf("%s%s", c.config.ControllerURL, endpoint))
	if err != nil {
		return fmt.Errorf("failed sending GET request to %s: %w", endpoint, err)
	}

	if err := parse(r, response); err != nil {
		return err
	}
	log.Debugf("GET request to %s. Response status code: %d, Body: %+v\n", endpoint, r.StatusCode, response)
	return nil
}

func NewClient(config ClientConfig) *Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConnsPerHost = 50

	return &Client{
		config: config,
		httpClient: &http.Client{
			Transport: transport,
		},
	}
}
