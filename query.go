package tpuf

import (
	"context"
	"encoding/json"
	"fmt"
)

type QueryRequest struct {
	// Vector is the vector to search for.
	Vector []float32 `json:"vector,omitempty"`
	// DistanceMetric is the distance metric to use for vector search.
	// Required if Vector is set.
	DistanceMetric DistanceMetric `json:"distance_metric,omitempty"`
	// RankBy is the fields to rank by for BM25 search.
	// Either Vector or RankBy, but not both, may be set.
	RankBy []interface{} `json:"rank_by,omitempty"`
	// TopK is the maximum number of results to return.  Default 10.
	TopK int `json:"top_k,omitempty"`
	// IncludeVectors includes the vectors of the results.  Default false.
	IncludeVectors bool `json:"include_vectors,omitempty"`
	// IncludeAttributes specifies which attributes to include in the results.
	// May be a list of specific attribute names, or true to include all attributes.
	IncludeAttributes interface{} `json:"include_attributes,omitempty"`
	// Filters is the filter to apply to the query, which may be a basic or compound filter.
	// See filter.go for more details.
	Filters Filter `json:"filters,omitempty"`
}

type QueryResult struct {
	Dist       float64         `json:"dist"`
	ID         string          `json:"id"`
	Vector     []float32       `json:"vector,omitempty"`
	Attributes json.RawMessage `json:"attributes,omitempty"`
}

// Query queries documents in the given namespace.
// See https://turbopuffer.com/docs/query
// Supports vector search, BM25 full-text search, and filter-only search.
// For vector search, provide a Vector and DistanceMetric.
// For BM25 search, provide RankBy.
// For filter-only search, omit both Vector and RankBy.
func (c *Client) Query(ctx context.Context, namespace string, request *QueryRequest) ([]*QueryResult, error) {
	path := fmt.Sprintf("/v1/vectors/%s/query", namespace)
	reqJson, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.post(ctx, path, reqJson)
	if err != nil {
		return nil, fmt.Errorf("failed to query documents: %w", err)
	}
	defer resp.Body.Close()

	var results []*QueryResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return results, nil
}
