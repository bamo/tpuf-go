package tpuf

import (
	"context"
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
