// Package tpuf provides a go client for the Turbopuffer API.
// See https://turbopuffer.com/docs for more information.
package tpuf

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client represents the main client for interacting with the API.
type Client struct {
	// ApiToken is the turbopuffer API token used to authenticate all requests.  Required.
	ApiToken string

	// BaseURL is the base URL for all API endpoints.
	// Defaults to https://api.turbopuffer.com/v1
	BaseURL string

	// HttpClient is the HTTP client used for making requests.
	// Defaults to defaultHttpClient.
	HttpClient HttpClient
}

var defaultBaseURL = "https://api.turbopuffer.com/v1"

func (c *Client) baseURL() string {
	if c.BaseURL == "" {
		return defaultBaseURL
	}
	return c.BaseURL
}

var defaultHttpClient = &http.Client{}

func (c *Client) httpClient() HttpClient {
	if c.HttpClient == nil {
		return defaultHttpClient
	}
	return c.HttpClient
}

func (c *Client) post(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	return c.do(ctx, http.MethodPost, path, body)
}

func (c *Client) do(ctx context.Context, method string, path string, body io.Reader) (*http.Response, error) {
	endpoint, err := url.JoinPath(c.baseURL(), path)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.ApiToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	return c.httpClient().Do(req)
}

type ApiResponse struct {
	Status string `json:"status"`
	Err    string `json:"error"`
}

const ApiStatusOK = "OK"

func (r ApiResponse) Error() string {
	if r.Err == "" {
		return ""
	}
	return fmt.Sprintf("%s: %s", r.Status, r.Err)
}
