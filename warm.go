package tpuf

import (
	"context"
	"encoding/json"
	"fmt"
)

type WarmCacheResult struct {
	Status string `json:"status"`
	Message string `json:"message"`
}

// Hints to turbopuffer that it should warm the hint cache for the given namespace.
// See https://turbopuffer.com/docs/warm-cache for more details.
func (c *Client) WarmCache(ctx context.Context, namespace string) (*WarmCacheResult, error) {
	path := fmt.Sprintf("/v1/namespaces/%s/hint_cache_warm", namespace)

	respData, err := c.get(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to warm cache: %w", err)
	}

	var warmCacheResult WarmCacheResult
	if err := json.Unmarshal(respData, &warmCacheResult); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &warmCacheResult, nil
}
