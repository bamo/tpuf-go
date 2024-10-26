package tpuf

import (
	"bytes"
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

			_, err := client.do(context.Background(), http.MethodGet, "/test", nil, nil)

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
