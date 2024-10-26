// Package tpuf provides a go client for the Turbopuffer API.
// See https://turbopuffer.com/docs for more information.
package tpuf

import (
	"context"
	"encoding/json"
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
	// Defaults to https://api.turbopuffer.com
	BaseURL string

	// HttpClient is the HTTP client used for making requests.
	// Defaults to &http.Client{}.
	HttpClient HttpClient
}

var defaultBaseURL = "https://api.turbopuffer.com"

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

func (c *Client) get(ctx context.Context, path string, values url.Values) (*http.Response, error) {
	return c.do(ctx, http.MethodGet, path, values, nil)
}

func (c *Client) post(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	return c.do(ctx, http.MethodPost, path, nil, body)
}

func (c *Client) delete(ctx context.Context, path string) (*http.Response, error) {
	return c.do(ctx, http.MethodDelete, path, nil, nil)
}

func (c *Client) do(ctx context.Context, method string, path string, values url.Values, body io.Reader) (*http.Response, error) {
	endpoint, err := url.JoinPath(c.baseURL(), path)
	if err != nil {
		return nil, err
	}
	reqUrl, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	reqUrl.RawQuery = values.Encode()

	req, err := http.NewRequestWithContext(ctx, method, reqUrl.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.ApiToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := c.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	// Convert HTTP 4XX and 5XX errors to API errors, and return them that way.
	if resp.StatusCode != http.StatusOK {
		apiErr := c.toApiError(resp)
		resp.Body.Close()
		return nil, apiErr
	}
	return resp, nil
}

func (c *Client) toApiError(resp *http.Response) error {
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	apiErr := ApiError{
		HttpStatus: resp.StatusCode,
	}
	if decodeErr := json.Unmarshal(respBody, &apiErr); decodeErr != nil {
		return fmt.Errorf("failed to decode api error: %w (raw response: %s, status code: %d)", decodeErr, string(respBody), resp.StatusCode)
	}
	if resp.StatusCode == http.StatusOK && apiErr.Status == ApiStatusOK {
		return nil
	}
	return apiErr
}

type ApiError struct {
	Status     string `json:"status"`
	Err        string `json:"error"`
	HttpStatus int    `json:"-"`
}

const ApiStatusOK = "OK"

func (e ApiError) Error() string {
	return fmt.Sprintf("%s: %s (HTTP %d)", e.Status, e.Err, e.HttpStatus)
}
