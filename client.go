// Package tpuf provides a go client for the Turbopuffer API.
// See https://turbopuffer.com/docs for more information.
package tpuf

import (
	"bytes"
	"compress/gzip"
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

	// UseGzipEncoding enables gzip encoding for requests and responses.
	UseGzipEncoding bool

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

	var bodyBytes []byte
	if body != nil {
		bodyBytes, err = io.ReadAll(body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}
	}

	return backoff.RetryNotifyWithTimerAndData(
		func() (*http.Response, error) {
			var bodyToUse io.Reader
			if bodyBytes != nil {
				bodyToUse = bytes.NewReader(bodyBytes)
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

func (c *Client) doOnce(ctx context.Context, method string, reqUrl *url.URL, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, reqUrl.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.ApiToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if c.UseGzipEncoding {
		if err := c.maybeEncodeGzip(req); err != nil {
			return nil, err
		}
	}

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return nil, err
	}

	if c.UseGzipEncoding {
		if err := c.maybeDecodeGzip(resp); err != nil {
			resp.Body.Close()
			return nil, err
		}
	}

	if resp.StatusCode != http.StatusOK {
		apiErr := c.toApiError(resp)
		resp.Body.Close()
		if !isRetriable(resp.StatusCode) {
			return nil, backoff.Permanent(apiErr)
		}
		return nil, apiErr
	}

	return resp, nil
}

func (c *Client) maybeEncodeGzip(req *http.Request) error {
	req.Header.Set("Accept-Encoding", "gzip")
	if req.Body == nil {
		return nil
	}

	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}
	if _, err := gzipWriter.Write(bodyBytes); err != nil {
		return fmt.Errorf("failed to compress request body: %w", err)
	}
	if err := gzipWriter.Close(); err != nil {
		return fmt.Errorf("failed to close gzip writer: %w", err)
	}
	req.Body = io.NopCloser(&buf)
	req.Header.Set("Content-Encoding", "gzip")
	return nil
}

func (c *Client) maybeDecodeGzip(resp *http.Response) error {
	if resp.Header.Get("Content-Encoding") != "gzip" {
		return nil
	}
	gzipReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	resp.Body = gzipReader
	return nil
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
