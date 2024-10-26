package tpuf

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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

	resp, err := c.get(ctx, path, params)
	if err != nil {
		return nil, fmt.Errorf("failed to export documents: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusAccepted {
		// TODO: handle retries.
		return nil, fmt.Errorf("export data not ready, retry after a few seconds")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to export documents: %w", c.toApiError(resp))
	}

	var exportResp ExportResponse
	if err := json.NewDecoder(resp.Body).Decode(&exportResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &exportResp, nil
}
