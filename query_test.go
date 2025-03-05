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

func TestQuery(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		request        *tpuf.QueryRequest
		httpResponse   *http.Response
		httpError      error
		expectedError  string
		expectedMethod string
		expectedURL    string
		expectedBody   string
		expectedResult []*tpuf.QueryResult
	}{
		{
			name:      "vector search",
			namespace: "test-namespace",
			request: &tpuf.QueryRequest{
				Vector:         []float32{0.1, 0.2, 0.3},
				DistanceMetric: tpuf.DistanceMetricCosine,
				TopK:           5,
				IncludeVectors: true,
			},
			httpResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewBufferString(`[
					{"id":"1","dist":0.1,"vector":[0.11,0.21,0.31]},
					{"id":"2","dist":0.2,"vector":[0.12,0.22,0.32]}
				]`)),
			},
			expectedMethod: http.MethodPost,
			expectedURL:    "https://api.turbopuffer.com/v1/namespaces/test-namespace/query",
			expectedBody:   `{"vector":[0.1,0.2,0.3],"distance_metric":"cosine_distance","top_k":5,"include_vectors":true}`,
			expectedResult: []*tpuf.QueryResult{
				{ID: "1", Dist: 0.1, Vector: []float32{0.11, 0.21, 0.31}},
				{ID: "2", Dist: 0.2, Vector: []float32{0.12, 0.22, 0.32}},
			},
		},
		{
			name:      "BM25 text search",
			namespace: "test-namespace",
			request: &tpuf.QueryRequest{
				RankBy: []interface{}{"description", "BM25", "fox jumping"},
				TopK:   3,
			},
			httpResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewBufferString(`[
					{"id":"1","dist":1.5},
					{"id":"2","dist":1.2},
					{"id":"3","dist":0.8}
				]`)),
			},
			expectedMethod: http.MethodPost,
			expectedURL:    "https://api.turbopuffer.com/v1/namespaces/test-namespace/query",
			expectedBody:   `{"rank_by":["description","BM25","fox jumping"],"top_k":3}`,
			expectedResult: []*tpuf.QueryResult{
				{ID: "1", Dist: 1.5},
				{ID: "2", Dist: 1.2},
				{ID: "3", Dist: 0.8},
			},
		},
		{
			name:      "filter-only search",
			namespace: "test-namespace",
			request: &tpuf.QueryRequest{
				Filters: &tpuf.AndFilter{
					Filters: []tpuf.Filter{
						&tpuf.BaseFilter{Attribute: "category", Operator: tpuf.OpEq, Value: "electronics"},
						&tpuf.BaseFilter{Attribute: "price", Operator: tpuf.OpGte, Value: 100},
					},
				},
				TopK: 2,
			},
			httpResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewBufferString(`[
					{"id":"1","dist":0},
					{"id":"2","dist":0}
				]`)),
			},
			expectedMethod: http.MethodPost,
			expectedURL:    "https://api.turbopuffer.com/v1/namespaces/test-namespace/query",
			expectedBody:   `{"filters":["And",[["category", "Eq", "electronics"],["price", "Gte", 100]]],"top_k":2}`,
			expectedResult: []*tpuf.QueryResult{
				{ID: "1", Dist: 0},
				{ID: "2", Dist: 0},
			},
		},
		{
			name:      "nil filter",
			namespace: "test-namespace",
			request: &tpuf.QueryRequest{
				Filters: nil,
				TopK:    2,
			},
			httpResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewBufferString(`[
					{"id":"1","dist":0},
					{"id":"2","dist":0}
				]`)),
			},
			expectedMethod: http.MethodPost,
			expectedURL:    "https://api.turbopuffer.com/v1/namespaces/test-namespace/query",
			expectedBody:   `{"top_k":2}`,
			expectedResult: []*tpuf.QueryResult{
				{ID: "1", Dist: 0},
				{ID: "2", Dist: 0},
			},
		},
		{
			name:      "query error",
			namespace: "test-namespace",
			request:   &tpuf.QueryRequest{TopK: 1},
			httpResponse: &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(bytes.NewBufferString(`{"error":"Invalid query","status":"error"}`)),
			},
			expectedError:  "failed to query documents: error: Invalid query (HTTP 400)",
			expectedMethod: http.MethodPost,
			expectedURL:    "https://api.turbopuffer.com/v1/namespaces/test-namespace/query",
			expectedBody:   `{"top_k":1}`,
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

			results, err := client.Query(context.Background(), tt.namespace, tt.request)

			if tt.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, results, "unexpected query results")
			} else {
				assert.EqualError(t, err, tt.expectedError)
				assert.Nil(t, results)
			}
		})
	}
}
