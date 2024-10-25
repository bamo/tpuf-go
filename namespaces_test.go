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

func TestNamespaces(t *testing.T) {
	tests := []struct {
		name           string
		request        *tpuf.NamespacesRequest
		httpResponse   *http.Response
		httpError      error
		expectedError  string
		expectedMethod string
		expectedURL    string
		expectedResult *tpuf.NamespacesResponse
	}{
		{
			name: "list namespaces without cursor",
			request: &tpuf.NamespacesRequest{
				Prefix:   "test",
				PageSize: 10,
			},
			httpResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewBufferString(`{
					"namespaces": [
						{"id": "test1"},
						{"id": "test2"}
					],
					"next_cursor": "next_page_cursor"
				}`)),
			},
			expectedMethod: http.MethodGet,
			expectedURL:    "https://api.turbopuffer.com/v1/vectors?page_size=10&prefix=test",
			expectedResult: &tpuf.NamespacesResponse{
				Namespaces: []*tpuf.Namespace{
					{ID: "test1"},
					{ID: "test2"},
				},
				NextCursor: "next_page_cursor",
			},
		},
		{
			name: "list namespaces with cursor",
			request: &tpuf.NamespacesRequest{
				Prefix:   "test",
				PageSize: 5,
				Cursor:   "previous_page_cursor",
			},
			httpResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewBufferString(`{
					"namespaces": [
						{"id": "test3"},
						{"id": "test4"}
					],
					"next_cursor": ""
				}`)),
			},
			expectedMethod: http.MethodGet,
			expectedURL:    "https://api.turbopuffer.com/v1/vectors?cursor=previous_page_cursor&page_size=5&prefix=test",
			expectedResult: &tpuf.NamespacesResponse{
				Namespaces: []*tpuf.Namespace{
					{ID: "test3"},
					{ID: "test4"},
				},
				NextCursor: "",
			},
		},
		{
			name: "list namespaces error",
			request: &tpuf.NamespacesRequest{
				PageSize: 10,
			},
			httpResponse: &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(bytes.NewBufferString(`{"error":"Invalid request","status":"error"}`)),
			},
			expectedError:  "failed to list namespaces: error: Invalid request (HTTP 400)",
			expectedMethod: http.MethodGet,
			expectedURL:    "https://api.turbopuffer.com/v1/vectors?page_size=10",
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

						assert.Equal(t, "Bearer test-token", req.Header.Get("Authorization"), "unexpected Authorization header")
						assert.Equal(t, "application/json", req.Header.Get("Accept"), "unexpected Accept header")

						return tt.httpResponse, tt.httpError
					},
				},
			}

			result, err := client.Namespaces(context.Background(), tt.request)

			if tt.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result, "unexpected namespaces result")
			} else {
				assert.EqualError(t, err, tt.expectedError)
				assert.Nil(t, result)
			}
		})
	}
}

func TestDeleteNamespace(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		httpResponse   *http.Response
		httpError      error
		expectedError  string
		expectedMethod string
		expectedURL    string
	}{
		{
			name:      "successful delete",
			namespace: "test-namespace",
			httpResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{"status":"OK"}`)),
			},
			expectedMethod: http.MethodDelete,
			expectedURL:    "https://api.turbopuffer.com/v1/vectors/test-namespace",
		},
		{
			name:      "delete failure",
			namespace: "non-existent-namespace",
			httpResponse: &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(bytes.NewBufferString(`{"error":"Namespace not found","status":"error"}`)),
			},
			expectedError:  "delete namespace failed: error: Namespace not found (HTTP 404)",
			expectedMethod: http.MethodDelete,
			expectedURL:    "https://api.turbopuffer.com/v1/vectors/non-existent-namespace",
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

						assert.Equal(t, "Bearer test-token", req.Header.Get("Authorization"), "unexpected Authorization header")
						assert.Equal(t, "application/json", req.Header.Get("Accept"), "unexpected Accept header")

						return tt.httpResponse, tt.httpError
					},
				},
			}

			err := client.DeleteNamespace(context.Background(), tt.namespace)

			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expectedError)
			}
		})
	}
}
