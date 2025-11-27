// Package client provides a REST-based HTTP client for interacting with Confluent Cloud and Platform APIs.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/creiche/confluent-go/pkg/api"
)

// Config holds the configuration for the Confluent REST client.
type Config struct {
	// BaseURL is the base URL for the Confluent API (e.g., https://api.confluent.cloud)
	BaseURL string
	// APIKey is the Confluent Cloud API key
	APIKey string
	// APISecret is the Confluent Cloud API secret
	APISecret string
	// HTTPClient is the HTTP client to use (optional, defaults to http.DefaultClient)
	HTTPClient *http.Client
}

// Client is a REST-based HTTP client for Confluent Cloud and Platform APIs.
type Client struct {
	config     Config
	httpClient *http.Client
}

// NewClient creates a new Confluent REST client with the given configuration.
func NewClient(config Config) (*Client, error) {
	if config.BaseURL == "" {
		return nil, fmt.Errorf("BaseURL is required in config")
	}
	if config.APIKey == "" {
		return nil, fmt.Errorf("APIKey is required in config")
	}
	if config.APISecret == "" {
		return nil, fmt.Errorf("APISecret is required in config")
	}

	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{
		config:     config,
		httpClient: httpClient,
	}, nil
}

// Request represents an HTTP request to the Confluent API.
type Request struct {
	Method  string
	Path    string
	Body    interface{}
	Headers map[string]string
}

// Response represents an HTTP response from the Confluent API.
type Response struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

// Do executes an HTTP request to the Confluent API.
func (c *Client) Do(ctx context.Context, req Request) (*Response, error) {
	url := strings.TrimSuffix(c.config.BaseURL, "/") + "/" + strings.TrimPrefix(req.Path, "/")

	var body io.Reader
	if req.Body != nil {
		jsonBody, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		body = bytes.NewReader(jsonBody)
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set authentication headers
	httpReq.SetBasicAuth(c.config.APIKey, c.config.APISecret)

	// Set default headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	// Set custom headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	resp := &Response{
		StatusCode: httpResp.StatusCode,
		Body:       respBody,
		Headers:    httpResp.Header,
	}

	// Check for API errors
	if httpResp.StatusCode >= 400 {
		return resp, api.NewError(httpResp.StatusCode, respBody, httpResp.Header)
	}

	return resp, nil
}

// DecodeJSON decodes the response body as JSON into the provided value.
func (r *Response) DecodeJSON(v interface{}) error {
	if len(r.Body) == 0 {
		return nil
	}
	return json.Unmarshal(r.Body, v)
}
