package tpuf

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

type NamespaceCursor string

type NamespacesRequest struct {
	// Prefix is an optional prefix by which to filter namespaces.
	Prefix string `json:"prefix,omitempty"`
	// PageSize is the maximum number of namespaces to return.  Default is 1000.
	PageSize int `json:"page_size,omitempty"`
	// Cursor the cursor to use for pagination.  Omit to get the first page.
	Cursor NamespaceCursor `json:"cursor,omitempty"`
}

type Namespace struct {
	ID string `json:"id"`
}

type NamespacesResponse struct {
	// Namespaces is the list of namespaces.
	Namespaces []*Namespace `json:"namespaces"`
	// NextCursor is the cursor which can be used to fetch the next page.
	NextCursor NamespaceCursor `json:"next_cursor,omitempty"`
}

// Namespaces lists all namespaces, optionally filtered by prefix.
// This query is paginated according to the input page size.  The returned NextCursor may be used to fetch the next page.
// See https://turbopuffer.com/docs/namespaces for more details.
func (c *Client) Namespaces(ctx context.Context, request *NamespacesRequest) (*NamespacesResponse, error) {
	path := "/v1/vectors"
	params := url.Values{}
	if request.PageSize > 0 {
		params.Set("page_size", strconv.Itoa(request.PageSize))
	}
	if request.Prefix != "" {
		params.Set("prefix", request.Prefix)
	}
	if request.Cursor != "" {
		params.Set("cursor", string(request.Cursor))
	}

	resp, err := c.get(ctx, path, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}
	defer resp.Body.Close()

	var response NamespacesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// DeleteNamespace deletes a namespace entirely, including all documents.
// See https://turbopuffer.com/docs/delete-namespace for more details.
func (c *Client) DeleteNamespace(ctx context.Context, namespace string) error {
	path := fmt.Sprintf("/v1/vectors/%s", namespace)
	resp, err := c.delete(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to delete namespace: %w", err)
	}
	defer resp.Body.Close()

	return nil
}
