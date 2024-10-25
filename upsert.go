package tpuf

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

// Upsert creates, updates or deletes documents in a namespace.
// See https://turbopuffer.com/docs/upsert
func (c *Client) Upsert(ctx context.Context, namespace string, request *UpsertRequest) error {
	return c.upsert(ctx, namespace, request, false)
}

func (c *Client) upsert(ctx context.Context, namespace string, request *UpsertRequest, allowDelete bool) error {
	url := fmt.Sprintf("/v1/vectors/%s", namespace)
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
	resp, err := c.post(ctx, url, bytes.NewBuffer(reqJson))
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var response ApiResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return fmt.Errorf("failed to decode response: %w (raw response: %s, status code: %d)", err, string(respBody), resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code %d: %w", resp.StatusCode, response)
	}

	if response.Status != ApiStatusOK {
		return fmt.Errorf("upsert failed: %w", response)
	}

	return nil
}
