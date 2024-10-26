package tpuf

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

type ExportResponse struct {
	IDs        []string                     `json:"ids"`
	Vectors    [][]float32                  `json:"vectors"`
	Attributes map[string][]json.RawMessage `json:"attributes"`
	NextCursor string                       `json:"next_cursor"`
}

// Export paginates through all documents in a namespace.
// It returns documents in a column-oriented layout.
// Use the NextCursor from the response to retrieve the next page of results.
func (c *Client) Export(ctx context.Context, namespace string, cursor string) (*ExportResponse, error) {
	path := fmt.Sprintf("/v1/vectors/%s", namespace)

	params := url.Values{}
	if cursor != "" {
		params.Set("cursor", string(cursor))
	}

	respData, err := c.get(ctx, path, params)
	if err != nil {
		return nil, fmt.Errorf("failed to export documents: %w", err)
	}

	var exportResp ExportResponse
	if err := json.Unmarshal(respData, &exportResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &exportResp, nil
}
