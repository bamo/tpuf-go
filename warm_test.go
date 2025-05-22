package tpuf_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/bamo/tpuf-go"
	"github.com/stretchr/testify/assert"
)

func TestWarmCache(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		httpResponse   *http.Response
		httpError      error
		expectedError  string
		expectedMethod string
		expectedURL    string
		expectedResult *tpuf.WarmCacheResult
	}{
		{
			name:      "successful warm cache",
			namespace: "test-namespace",
			httpResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewBufferString(`{
					"status": "ACCEPTED",
					"message": "cache starting to warm"
				}`)),
			},
			expectedMethod: http.MethodGet,
			expectedURL:    "https://api.turbopuffer.com/v1/namespaces/test-namespace/hint_cache_warm",
			expectedResult: &tpuf.WarmCacheResult{
				Status:  "ACCEPTED",
				Message: "cache starting to warm",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &tpuf.Client{
				ApiToken: "test-token",
				HttpClient: &fakeHttpClient{
					doFunc: func(req *http.Request) (*http.Response, error) {
						assert.Equal(t, tt.expectedMethod, req.Method, "unexpected request method")
						assert.Equal(t, tt.expectedURL, req.URL.String(), "unexpected request URL")

						return tt.httpResponse, tt.httpError
					},
				},
			}

			result, err := client.WarmCache(context.Background(), tt.namespace)
			if tt.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result, "unexpected warm cache result")
			} else {
				assert.EqualError(t, err, tt.expectedError)
				assert.Nil(t, result)
			}
		})
	}
}
