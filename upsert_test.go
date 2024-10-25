package tpuf_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/bamo/tpuf-go"
	"github.com/stretchr/testify/assert"
)

type fakeHttpClient struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (f *fakeHttpClient) Do(req *http.Request) (*http.Response, error) {
	return f.doFunc(req)
}

func TestUpsert(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		request        *tpuf.UpsertRequest
		httpResponse   *http.Response
		httpError      error
		expectedError  string
		expectedMethod string
		expectedURL    string
		expectedBody   string
	}{
		{
			name:      "successful upsert",
			namespace: "test-namespace",
			request: &tpuf.UpsertRequest{
				DistanceMetric: "cosine_distance",
				Upserts: []*tpuf.Upsert{
					{
						ID:     "1",
						Vector: []float32{0.1, 0.1},
						Attributes: map[string]interface{}{
							"my-string":       "one",
							"my-uint":         12,
							"my-bool":         true,
							"my-string-array": []string{"a", "b"},
						},
					},
					{
						ID:     "2",
						Vector: []float32{0.2, 0.2},
						Attributes: map[string]interface{}{
							"my-string-array": []string{"b", "d"},
						},
					},
				},
			},
			httpResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{"status":"OK"}`)),
			},
			expectedMethod: http.MethodPost,
			expectedURL:    "https://api.turbopuffer.com/v1/vectors/test-namespace",
			expectedBody:   `{"distance_metric":"cosine_distance","upserts":[{"id":"1","vector":[0.1,0.1],"attributes":{"my-bool":true,"my-string":"one","my-string-array":["a","b"],"my-uint":12}},{"id":"2","vector":[0.2,0.2],"attributes":{"my-string-array":["b","d"]}}]}`,
		},
		{
			name:      "unsuccessful upsert",
			namespace: "test-namespace",
			request: &tpuf.UpsertRequest{
				Upserts: []*tpuf.Upsert{{ID: "1", Vector: []float32{0.1, 0.1}}},
			},
			httpResponse: &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(bytes.NewBufferString(`{"error":"💔 invalid filter for key my_attr, only Eq/In/Lt/Lte/Gt/Gte/And/Or filters allowed currently for scanning","status":"error"}`)),
			},
			expectedError:  "upsert failed: error: 💔 invalid filter for key my_attr, only Eq/In/Lt/Lte/Gt/Gte/And/Or filters allowed currently for scanning (HTTP 400)",
			expectedMethod: http.MethodPost,
			expectedURL:    "https://api.turbopuffer.com/v1/vectors/test-namespace",
			expectedBody:   `{"upserts":[{"id":"1","vector":[0.1,0.1]}]}`,
		},
		{
			name:           "http error",
			namespace:      "test-namespace",
			request:        &tpuf.UpsertRequest{},
			httpError:      &url.Error{Op: "Post", URL: "https://api.turbopuffer.com/v1/v1/vectors/test-namespace", Err: io.EOF},
			expectedError:  "http request failed: Post \"https://api.turbopuffer.com/v1/v1/vectors/test-namespace\": EOF",
			expectedMethod: http.MethodPost,
			expectedURL:    "https://api.turbopuffer.com/v1/vectors/test-namespace",
			expectedBody:   `{}`,
		},
		{
			name: "delete via upsert",
			request: &tpuf.UpsertRequest{
				Upserts: []*tpuf.Upsert{{ID: "1"}},
			},
			expectedError: "deletion must be performed using Delete, not Upsert to avoid accidental deletion",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &tpuf.Client{
				ApiToken: "test-token",
				HttpClient: &fakeHttpClient{
					doFunc: func(req *http.Request) (*http.Response, error) {
						assert.Equal(t, tt.expectedMethod, req.Method)
						assert.Equal(t, tt.expectedURL, req.URL.String())

						body, _ := io.ReadAll(req.Body)
						assert.JSONEq(t, tt.expectedBody, string(body))

						assert.Equal(t, "Bearer test-token", req.Header.Get("Authorization"))
						assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
						assert.Equal(t, "application/json", req.Header.Get("Accept"))

						return tt.httpResponse, tt.httpError
					},
				},
			}

			err := client.Upsert(context.Background(), tt.namespace, tt.request)

			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expectedError)
			}
		})
	}
}

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
			expectedError:  "upsert failed: error: Invalid request (HTTP 400)",
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
