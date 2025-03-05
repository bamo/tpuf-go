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
			expectedURL:    "https://api.turbopuffer.com/v1/namespaces/test-namespace",
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
			expectedURL:    "https://api.turbopuffer.com/v1/namespaces/test-namespace",
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

func TestDeleteByFilter(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		request        *tpuf.DeleteByFilterRequest
		httpResponse   *http.Response
		httpError      error
		expectedError  string
		expectedMethod string
		expectedURL    string
		expectedBody   string
	}{
		{
			name:      "successful delete by filter",
			namespace: "test-namespace",
			request: &tpuf.DeleteByFilterRequest{
				Filter: &tpuf.BaseFilter{
					Attribute: "category",
					Operator:  tpuf.OpEq,
					Value:     "electronics",
				},
			},
			httpResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{"status":"OK"}`)),
			},
			expectedMethod: http.MethodPost,
			expectedURL:    "https://api.turbopuffer.com/v1/namespaces/test-namespace",
			expectedBody:   `{"delete_by_filter":["category","Eq","electronics"]}`,
		},
		{
			name:      "delete by complex filter",
			namespace: "test-namespace",
			request: &tpuf.DeleteByFilterRequest{
				Filter: &tpuf.AndFilter{
					Filters: []tpuf.Filter{
						&tpuf.BaseFilter{
							Attribute: "category",
							Operator:  tpuf.OpEq,
							Value:     "electronics",
						},
						&tpuf.BaseFilter{
							Attribute: "price",
							Operator:  tpuf.OpGt,
							Value:     100,
						},
					},
				},
			},
			httpResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{"status":"OK"}`)),
			},
			expectedMethod: http.MethodPost,
			expectedURL:    "https://api.turbopuffer.com/v1/namespaces/test-namespace",
			expectedBody:   `{"delete_by_filter":["And",[["category","Eq","electronics"],["price","Gt",100]]]}`,
		},
		{
			name:      "delete by filter error",
			namespace: "test-namespace",
			request: &tpuf.DeleteByFilterRequest{
				Filter: &tpuf.BaseFilter{
					Attribute: "invalid_field",
					Operator:  tpuf.OpEq,
					Value:     "value",
				},
			},
			httpResponse: &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(bytes.NewBufferString(`{"error":"Invalid filter","status":"error"}`)),
			},
			expectedError:  "failed to delete by filter: error: Invalid filter (HTTP 400)",
			expectedMethod: http.MethodPost,
			expectedURL:    "https://api.turbopuffer.com/v1/namespaces/test-namespace",
			expectedBody:   `{"delete_by_filter":["invalid_field","Eq","value"]}`,
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

			err := client.DeleteByFilter(context.Background(), tt.namespace, tt.request)

			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expectedError)
			}
		})
	}
}
