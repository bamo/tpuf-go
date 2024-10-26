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

func TestRecall(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		request        *tpuf.RecallRequest
		httpResponse   *http.Response
		httpError      error
		expectedError  string
		expectedMethod string
		expectedURL    string
		expectedBody   string
		expectedResult *tpuf.RecallResponse
	}{
		{
			name:      "successful recall",
			namespace: "test-namespace",
			request: &tpuf.RecallRequest{
				Num:  100,
				TopK: 10,
				Filters: &tpuf.AndFilter{
					Filters: []tpuf.Filter{
						&tpuf.BaseFilter{Attribute: "category", Operator: tpuf.OpEq, Value: "electronics"},
					},
				},
				Queries: [][]float32{{0.1, 0.2, 0.3}, {0.4, 0.5, 0.6}},
			},
			httpResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewBufferString(`{
					"avg_recall": 0.95,
					"avg_exhaustive_count": 1000,
					"avg_ann_count": 100
				}`)),
			},
			expectedMethod: http.MethodPost,
			expectedURL:    "https://api.turbopuffer.com/v1/vectors/test-namespace/_debug/recall",
			expectedBody:   `{"num":100,"top_k":10,"filters":["And",[["category","Eq","electronics"]]],"queries":[[0.1,0.2,0.3],[0.4,0.5,0.6]]}`,
			expectedResult: &tpuf.RecallResponse{
				AvgRecall:          0.95,
				AvgExhaustiveCount: 1000,
				AvgAnnCount:        100,
			},
		},
		{
			name:      "recall error",
			namespace: "test-namespace",
			request: &tpuf.RecallRequest{
				Num: 10,
			},
			httpResponse: &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(bytes.NewBufferString(`{"error":"Invalid request","status":"error"}`)),
			},
			expectedError:  "failed to perform recall: error: Invalid request (HTTP 400)",
			expectedMethod: http.MethodPost,
			expectedURL:    "https://api.turbopuffer.com/v1/vectors/test-namespace/_debug/recall",
			expectedBody:   `{"num":10}`,
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

			result, err := client.Recall(context.Background(), tt.namespace, tt.request)

			if tt.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result, "unexpected recall result")
			} else {
				assert.EqualError(t, err, tt.expectedError)
				assert.Nil(t, result)
			}
		})
	}
}
