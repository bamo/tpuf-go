// Package tpuf provides a go client for the Turbopuffer API.
// See https://turbopuffer.com/docs for more information.
package tpuf

import (
	"net/http"
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

type ErrorReponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}
