// Package tpuf provides a go client for the Turbopuffer API.
// See https://turbopuffer.com/docs for more information.
package tpuf

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/cenkalti/backoff/v4"
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

	// MaxRetries is the maximum number of times to retry a request if a retriable
	// error is encountered.  Defaults to 6.
	// Retry interval is exponential backoff starting out at 2 seconds and maxing at 64.
	MaxRetries int

	// DisableRetry disables retries for all requests.
	DisableRetry bool

	// HttpClient is the HTTP client used for making requests.
	// Defaults to &http.Client{}.
	HttpClient HttpClient

	// Timer is the timer used for exponential backoff.
	Timer backoff.Timer
}

const defaultBaseURL = "https://api.turbopuffer.com"

const (
	GCPUSCentral1BaseURL  = "https://gcp-us-central1.turbopuffer.com"
	GCPUSWest1BaseURL     = "https://gcp-us-west1.turbopuffer.com"
	GCPUSEast4BaseURL     = "https://gcp-us-east4.turbopuffer.com"
	GCPEuropeWest3BaseURL = "https://gcp-europe-west3.turbopuffer.com"
)

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

const defaultMaxRetries = 6

func (c *Client) maxRetries() int {
	if c.DisableRetry {
		return 0
	}
	if c.MaxRetries == 0 {
		return defaultMaxRetries
	}
	return c.MaxRetries
}

func (c *Client) get(ctx context.Context, path string, values url.Values) ([]byte, error) {
	return c.do(ctx, http.MethodGet, path, values, nil)
}

func (c *Client) post(ctx context.Context, path string, body []byte) ([]byte, error) {
	return c.do(ctx, http.MethodPost, path, nil, body)
}

func (c *Client) delete(ctx context.Context, path string) ([]byte, error) {
	return c.do(ctx, http.MethodDelete, path, nil, nil)
}

func (c *Client) do(ctx context.Context, method string, path string, values url.Values, body []byte) ([]byte, error) {
	endpoint, err := url.JoinPath(c.baseURL(), path)
	if err != nil {
		return nil, err
	}
	reqUrl, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	reqUrl.RawQuery = values.Encode()

	return backoff.RetryNotifyWithTimerAndData(
		func() ([]byte, error) {
			var bodyToUse io.Reader
			if len(body) > 0 {
				bodyToUse = bytes.NewReader(body)
			}
			return c.doOnce(ctx, method, reqUrl, bodyToUse)
		},
		backoff.WithMaxRetries(backoff.NewExponentialBackOff(
			backoff.WithInitialInterval(2*time.Second),
			backoff.WithMultiplier(2.0),
			backoff.WithMaxInterval(64*time.Second),
		), uint64(c.maxRetries())),
		nil,
		c.Timer,
	)
}

func (c *Client) doOnce(ctx context.Context, method string, reqUrl *url.URL, body io.Reader) ([]byte, error) {
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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		apiErr := c.toApiError(resp)
		if !isRetriable(resp.StatusCode) {
			return nil, backoff.Permanent(apiErr)
		}
		return nil, apiErr
	}

	return io.ReadAll(resp.Body)
}

func isRetriable(statusCode int) bool {
	return statusCode >= 500 ||
		statusCode == http.StatusRequestTimeout ||
		statusCode == http.StatusTooManyRequests ||
		statusCode == http.StatusAccepted
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
