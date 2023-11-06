package ankacloud

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

type response struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Body    interface{} `json:"body,omitempty"`
}

type Client struct {
	ControllerURL string
	HttpClient    *http.Client
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

	r, err := c.HttpClient.Post(fmt.Sprintf("%s%s", c.ControllerURL, endpoint), "application/json", &buf)
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

	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s%s", c.ControllerURL, endpoint), &buf)
	if err != nil {
		return nil, fmt.Errorf("failed creating DELETE request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	r, err := c.HttpClient.Do(req)
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
	r, err := c.HttpClient.Get(fmt.Sprintf("%s%s", c.ControllerURL, endpoint))
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
