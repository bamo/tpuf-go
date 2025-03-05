package tpuf_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/bamo/tpuf-go"
	"github.com/stretchr/testify/assert"
)

func TestExport(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		cursor         string
		httpResponses  []*http.Response
		httpErrors     []error
		expectedError  string
		expectedMethod string
		expectedURL    string
		expectedResult *tpuf.ExportResponse
	}{
		{
			name:      "successful export without cursor",
			namespace: "test-namespace",
			cursor:    "",
			httpResponses: []*http.Response{
				{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(bytes.NewBufferString(`{
						"ids": ["1", "2"],
						"vectors": [[0.1, 0.1], [0.2, 0.2]],
						"attributes": {
							"key1": ["one", "two"],
							"key2": ["a", "b"]
						},
						"next_cursor": "eyJmaWxlX2lkIjoxMTMzfQ"
					}`)),
				},
			},
			expectedMethod: http.MethodGet,
			expectedURL:    "https://api.turbopuffer.com/v1/namespaces/test-namespace",
			expectedResult: &tpuf.ExportResponse{
				IDs:     []string{"1", "2"},
				Vectors: [][]float32{{0.1, 0.1}, {0.2, 0.2}},
				Attributes: map[string][]json.RawMessage{
					"key1": {json.RawMessage(`"one"`), json.RawMessage(`"two"`)},
					"key2": {json.RawMessage(`"a"`), json.RawMessage(`"b"`)},
				},
				NextCursor: "eyJmaWxlX2lkIjoxMTMzfQ",
			},
		},
		{
			name:      "successful export with cursor",
			namespace: "test-namespace",
			cursor:    "eyJmaWxlX2lkIjoxMTMzfQ",
			httpResponses: []*http.Response{
				{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(bytes.NewBufferString(`{
						"ids": ["3", "4"],
						"vectors": [[0.3, 0.3], [0.4, 0.4]],
						"attributes": {
							"key1": ["three", "four"],
							"key2": ["c", "d"]
						},
						"next_cursor": ""
					}`)),
				},
			},
			expectedMethod: http.MethodGet,
			expectedURL:    "https://api.turbopuffer.com/v1/namespaces/test-namespace?cursor=eyJmaWxlX2lkIjoxMTMzfQ",
			expectedResult: &tpuf.ExportResponse{
				IDs:     []string{"3", "4"},
				Vectors: [][]float32{{0.3, 0.3}, {0.4, 0.4}},
				Attributes: map[string][]json.RawMessage{
					"key1": {json.RawMessage(`"three"`), json.RawMessage(`"four"`)},
					"key2": {json.RawMessage(`"c"`), json.RawMessage(`"d"`)},
				},
				NextCursor: "",
			},
		},
		{
			name:      "export not ready then success",
			namespace: "test-namespace",
			cursor:    "",
			httpResponses: []*http.Response{
				{
					StatusCode: http.StatusAccepted,
					Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
				},
				{
					StatusCode: http.StatusAccepted,
					Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
				},
				{
					StatusCode: http.StatusOK,
					Body: io.NopCloser(bytes.NewBufferString(`{
						"ids": ["5", "6"],
						"vectors": [[0.5, 0.5], [0.6, 0.6]],
						"attributes": {
							"key1": ["five", "six"],
							"key2": ["e", "f"]
						},
						"next_cursor": ""
					}`)),
				},
			},
			expectedMethod: http.MethodGet,
			expectedURL:    "https://api.turbopuffer.com/v1/namespaces/test-namespace",
			expectedResult: &tpuf.ExportResponse{
				IDs:     []string{"5", "6"},
				Vectors: [][]float32{{0.5, 0.5}, {0.6, 0.6}},
				Attributes: map[string][]json.RawMessage{
					"key1": {json.RawMessage(`"five"`), json.RawMessage(`"six"`)},
					"key2": {json.RawMessage(`"e"`), json.RawMessage(`"f"`)},
				},
				NextCursor: "",
			},
		},
		{
			name:      "export error",
			namespace: "test-namespace",
			cursor:    "",
			httpResponses: []*http.Response{
				{
					StatusCode: http.StatusBadRequest,
					Body:       io.NopCloser(bytes.NewBufferString(`{"error":"Invalid request","status":"error"}`)),
				},
			},
			expectedError:  "failed to export documents: error: Invalid request (HTTP 400)",
			expectedMethod: http.MethodGet,
			expectedURL:    "https://api.turbopuffer.com/v1/namespaces/test-namespace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeTimer := &fakeTimer{}
			requestCount := 0
			client := &tpuf.Client{
				ApiToken: "test-token",
				HttpClient: &fakeHttpClient{
					doFunc: func(req *http.Request) (*http.Response, error) {
						assert.Equal(t, tt.expectedMethod, req.Method, "unexpected request method")
						assert.Equal(t, tt.expectedURL, req.URL.String(), "unexpected request URL")
						assert.Equal(t, "Bearer test-token", req.Header.Get("Authorization"), "unexpected Authorization header")
						assert.Equal(t, "application/json", req.Header.Get("Accept"), "unexpected Accept header")

						response := tt.httpResponses[requestCount]
						var err error
						if requestCount < len(tt.httpErrors) {
							err = tt.httpErrors[requestCount]
						}
						requestCount++
						return response, err
					},
				},
				Timer: fakeTimer,
			}

			result, err := client.Export(context.Background(), tt.namespace, tt.cursor)

			if tt.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result, "unexpected export result")
			} else {
				assert.EqualError(t, err, tt.expectedError)
				assert.Nil(t, result)
			}
		})
	}
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
