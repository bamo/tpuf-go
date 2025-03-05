package tpuf

import (
	"context"
	"encoding/json"
	"fmt"
)

// Delete deletes documents from a namespace.
// See https://turbopuffer.com/docs/upsert#document-deletion
func (c *Client) Delete(ctx context.Context, namespace string, ids []string) error {
	var upserts []*Upsert
	for _, id := range ids {
		upserts = append(upserts, &Upsert{ID: id})
	}
	return c.upsert(ctx, namespace, &UpsertRequest{
		Upserts: upserts,
	}, true)
}

type DeleteByFilterRequest struct {
	Filter Filter `json:"delete_by_filter"`
}

// DeleteByFilter deletes documents from a namespace based on a filter.
// See https://turbopuffer.com/docs/upsert#document-deletion
func (c *Client) DeleteByFilter(ctx context.Context, namespace string, request *DeleteByFilterRequest) error {
	path := fmt.Sprintf("/v1/namespaces/%s", namespace)
	reqJson, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}
	_, err = c.post(ctx, path, reqJson)
	if err != nil {
		return fmt.Errorf("failed to delete by filter: %w", err)
	}
	return nil
}
