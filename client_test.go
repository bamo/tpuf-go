package tpuf

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClientDo(t *testing.T) {
	tests := []struct {
		name          string
		maxRetries    int
		disableRetry  bool
		httpResponses []*http.Response
		httpErrors    []error
		expectedError string
		expectedCalls int
		method        string
		requestBody   string
	}{
		{
			name:       "success on first try",
			maxRetries: 3,
			httpResponses: []*http.Response{
				{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBuffer(nil)),
				},
			},
			expectedCalls: 1,
		},
		{
			name:       "retry on 429 TooManyRequests",
			maxRetries: 3,
			httpResponses: []*http.Response{
				{
					StatusCode: http.StatusTooManyRequests,
					Body:       io.NopCloser(bytes.NewBuffer(nil)),
				},
				{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBuffer(nil)),
				},
			},
			expectedCalls: 2,
		},
		{
			name:       "retry on 500 InternalServerError",
			maxRetries: 3,
			httpResponses: []*http.Response{
				{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(bytes.NewBuffer(nil)),
				},
				{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBuffer(nil)),
				},
			},
			expectedCalls: 2,
		},
		{
			name:       "max retries reached",
			maxRetries: 2,
			httpResponses: []*http.Response{
				{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(bytes.NewBuffer(nil)),
				},
				{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(bytes.NewBuffer(nil)),
				},
				{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(bytes.NewBuffer(nil)),
				},
			},
			expectedError: "failed to decode api error: unexpected end of JSON input (raw response: , status code: 500)",
			expectedCalls: 3,
		},
		{
			name:         "retry disabled",
			maxRetries:   3,
			disableRetry: true,
			httpResponses: []*http.Response{
				{
					StatusCode: http.StatusInternalServerError,
					Body:       io.NopCloser(bytes.NewBuffer(nil)),
				},
			},
			expectedError: "failed to decode api error: unexpected end of JSON input (raw response: , status code: 500)",
			expectedCalls: 1,
		},
		{
			name:       "non-retriable error",
			maxRetries: 3,
			httpResponses: []*http.Response{
				{
					StatusCode: http.StatusBadRequest,
					Body:       io.NopCloser(bytes.NewBuffer(nil)),
				},
			},
			expectedError: "failed to decode api error: unexpected end of JSON input (raw response: , status code: 400)",
			expectedCalls: 1,
		},
		{
			name:       "non-retriable error, unmarshalable",
			maxRetries: 3,
			httpResponses: []*http.Response{
				{
					StatusCode: http.StatusBadRequest,
					Body:       io.NopCloser(bytes.NewBufferString(`{"status": "error", "error": "invalid argument"}`)),
				},
			},
			expectedError: "error: invalid argument (HTTP 400)",
			expectedCalls: 1,
		},
		{
			name:        "POST request with body",
			maxRetries:  3,
			method:      http.MethodPost,
			requestBody: `{"key": "value"}`,
			httpResponses: []*http.Response{
				{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(`{"status": "OK"}`)),
				},
			},
			expectedCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeTimer := &fakeTimer{}
			callCount := 0
			client := &Client{
				ApiToken:     "test-token",
				MaxRetries:   tt.maxRetries,
				DisableRetry: tt.disableRetry,
				HttpClient: &fakeHttpClient{
					doFunc: func(req *http.Request) (*http.Response, error) {
						assert.Equal(t, "Bearer test-token", req.Header.Get("Authorization"), "unexpected Authorization header")
						assert.Equal(t, "application/json", req.Header.Get("Accept"), "unexpected Accept header")

						if tt.method == http.MethodPost {
							assert.Equal(t, "application/json", req.Header.Get("Content-Type"), "unexpected Content-Type header")
							body, err := io.ReadAll(req.Body)
							assert.NoError(t, err, "failed to read request body")
							assert.Equal(t, tt.requestBody, string(body), "unexpected request body")
						}

						response := tt.httpResponses[callCount]
						var err error
						if callCount < len(tt.httpErrors) {
							err = tt.httpErrors[callCount]
						}
						callCount++
						return response, err
					},
				},
				Timer: fakeTimer,
			}

			method := tt.method
			if method == "" {
				method = http.MethodGet
			}

			_, err := client.do(context.Background(), method, "/test", nil, []byte(tt.requestBody))

			assert.Equal(t, tt.expectedCalls, callCount, "unexpected number of calls")

			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expectedError)
			}
		})
	}
}

type fakeHttpClient struct {
	doFunc func(*http.Request) (*http.Response, error)
}

func (f *fakeHttpClient) Do(req *http.Request) (*http.Response, error) {
	return f.doFunc(req)
}

type fakeTimer struct {
	ch chan time.Time
}

func (f *fakeTimer) Start(duration time.Duration) {
	if f.ch == nil {
		f.ch = make(chan time.Time, 1)
	}
	f.ch <- time.Now()
}

func (f *fakeTimer) Stop() {
	if f.ch != nil {
		close(f.ch)
	}
}

func (f *fakeTimer) C() <-chan time.Time {
	return f.ch
}

func TestClientDoWithCompression(t *testing.T) {
	tests := []struct {
		name          string
		method        string
		requestBody   string
		expectedError string
		maxRetries    int
		expectedCalls int
		responses     []struct {
			body   string
			status int
		}
	}{
		{
			name:          "successful compressed POST request and response",
			method:        http.MethodPost,
			requestBody:   `{"key": "value"}`,
			expectedCalls: 1,
			responses: []struct {
				body   string
				status int
			}{
				{body: `{"status": "OK"}`, status: http.StatusOK},
			},
		},
		{
			name:          "compressed error response",
			method:        http.MethodPost,
			requestBody:   `{"key": "value"}`,
			expectedError: "error: 💥 something went wrong!!! (HTTP 400)",
			expectedCalls: 1,
			responses: []struct {
				body   string
				status int
			}{
				{body: `{"status": "error", "error": "💥 something went wrong!!!"}`, status: http.StatusBadRequest},
			},
		},
		{
			name:          "successful compressed GET request and response",
			method:        http.MethodGet,
			expectedCalls: 1,
			responses: []struct {
				body   string
				status int
			}{
				{body: `{"status": "OK"}`, status: http.StatusOK},
			},
		},
		{
			name:          "successful compressed DELETE request and response",
			method:        http.MethodDelete,
			expectedCalls: 1,
			responses: []struct {
				body   string
				status int
			}{
				{body: `{"status": "OK"}`, status: http.StatusOK},
			},
		},
		{
			name:          "retry with gzip encoding",
			method:        http.MethodPost,
			requestBody:   `{"key": "retry"}`,
			maxRetries:    2,
			expectedCalls: 2,
			responses: []struct {
				body   string
				status int
			}{
				{body: `{"status": "error", "error": "Too many requests"}`, status: http.StatusTooManyRequests},
				{body: `{"status": "OK"}`, status: http.StatusOK},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount := 0
			client := &Client{
				ApiToken:        "test-token",
				UseGzipEncoding: true,
				MaxRetries:      tt.maxRetries,
				HttpClient: &fakeHttpClient{
					doFunc: func(req *http.Request) (*http.Response, error) {
						callCount++

						// Check request headers
						assert.Equal(t, "Bearer test-token", req.Header.Get("Authorization"))
						assert.Equal(t, "application/json", req.Header.Get("Accept"))
						assert.Equal(t, "gzip", req.Header.Get("Accept-Encoding"))

						if tt.method == http.MethodPost {
							assert.Equal(t, "gzip", req.Header.Get("Content-Encoding"))

							// Check request body compression for POST requests
							gzipReader, err := gzip.NewReader(req.Body)
							assert.NoError(t, err)
							decompressedBody, err := io.ReadAll(gzipReader)
							assert.NoError(t, err)
							assert.Equal(t, tt.requestBody, string(decompressedBody))
						} else {
							assert.Empty(t, req.Header.Get("Content-Encoding"))
						}

						// Prepare compressed response
						var buf bytes.Buffer
						gzipWriter := gzip.NewWriter(&buf)
						responseIndex := callCount - 1
						if responseIndex >= len(tt.responses) {
							responseIndex = len(tt.responses) - 1
						}
						_, err := gzipWriter.Write([]byte(tt.responses[responseIndex].body))
						assert.NoError(t, err)
						assert.NoError(t, gzipWriter.Close())

						return &http.Response{
							StatusCode: tt.responses[responseIndex].status,
							Body:       io.NopCloser(&buf),
							Header: http.Header{
								"Content-Encoding": []string{"gzip"},
							},
						}, nil
					},
				},
				Timer: &fakeTimer{},
			}

			var resp *http.Response
			var err error

			if tt.method == http.MethodPost {
				resp, err = client.do(context.Background(), tt.method, "/test", nil, []byte(tt.requestBody))
			} else {
				resp, err = client.do(context.Background(), tt.method, "/test", nil, nil)
			}

			if tt.expectedError == "" {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				if resp != nil {
					body, readErr := io.ReadAll(resp.Body)
					assert.NoError(t, readErr)
					assert.Equal(t, tt.responses[len(tt.responses)-1].body, string(body))
				}
			} else {
				assert.EqualError(t, err, tt.expectedError)
				assert.Nil(t, resp)
			}

			assert.Equal(t, tt.expectedCalls, callCount)
		})
	}
}
