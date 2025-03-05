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

func TestDelete(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		ids            []string
		httpResponse   *http.Response
		httpError      error
		expectedError  string
		expectedMethod string
		expectedURL    string
		expectedBody   string
	}{
		{
			name:      "successful delete",
			namespace: "test-namespace",
			ids:       []string{"1", "2", "3"},
			httpResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{"status":"OK"}`)),
			},
			expectedMethod: http.MethodPost,
			expectedURL:    "https://api.turbopuffer.com/v1/vectors/test-namespace",
			expectedBody:   `{"upserts":[{"id":"1"},{"id":"2"},{"id":"3"}]}`,
		},
		{
			name:      "delete error",
			namespace: "test-namespace",
			ids:       []string{"4", "5"},
			httpResponse: &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(bytes.NewBufferString(`{"error":"Invalid request","status":"error"}`)),
			},
			expectedError:  "failed to upsert documents: error: Invalid request (HTTP 400)",
			expectedMethod: http.MethodPost,
			expectedURL:    "https://api.turbopuffer.com/v1/vectors/test-namespace",
			expectedBody:   `{"upserts":[{"id":"4"},{"id":"5"}]}`,
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

						body, _ := io.ReadAll(req.Body)
						assert.JSONEq(t, tt.expectedBody, string(body), "unexpected request body")

						assert.Equal(t, "Bearer test-token", req.Header.Get("Authorization"), "unexpected Authorization header")
						assert.Equal(t, "application/json", req.Header.Get("Content-Type"), "unexpected Content-Type header")
						assert.Equal(t, "application/json", req.Header.Get("Accept"), "unexpected Accept header")

						return tt.httpResponse, tt.httpError
					},
				},
			}

			err := client.Delete(context.Background(), tt.namespace, tt.ids)

			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expectedError)
			}
		})
	}
}
