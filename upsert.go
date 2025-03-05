package tpuf

import (
	"context"
	"encoding/json"
	"fmt"
)

// Attributes represent a document's attributes.  Must be a json-marshalable type.
type Attributes interface{}

// Upsert represents a single document to upsert.
type Upsert struct {
	// ID is the document's unique identifier.  Required.
	ID string `json:"id"`
	// Vector is an optionalvector embedding to use for similarity search.
	Vector []float32 `json:"vector,omitempty"`
	// Attributes is a json-marshalable object representing the document's attributes.
	Attributes Attributes `json:"attributes,omitempty"`
}

type UpsertRequest struct {
	DistanceMetric    DistanceMetric `json:"distance_metric,omitempty"`
	Schema            Schema         `json:"schema,omitempty"`
	Upserts           []*Upsert      `json:"upserts,omitempty"`
	CopyFromNamespace string         `json:"copy_from_namespace,omitempty"`
}

// Upsert creates or updates documents in a namespace.
// Note that although the API supports deletion via the upsert endpoint, this client requires
// that you use the Delete method explicitly to avoid accidental deletions.
// See https://turbopuffer.com/docs/upsert
func (c *Client) Upsert(ctx context.Context, namespace string, request *UpsertRequest) error {
	return c.upsert(ctx, namespace, request, false)
}

func (c *Client) upsert(ctx context.Context, namespace string, request *UpsertRequest, allowDelete bool) error {
	path := fmt.Sprintf("/v1/namespaces/%s", namespace)
	if !allowDelete {
		for _, upsert := range request.Upserts {
			if len(upsert.Vector) == 0 {
				return fmt.Errorf("deletion must be performed using Delete, not Upsert to avoid accidental deletion")
			}
		}
	}
	reqJson, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}
	_, err = c.post(ctx, path, reqJson)
	if err != nil {
		return fmt.Errorf("failed to upsert documents: %w", err)
	}

	return nil
}
