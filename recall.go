package tpuf

import (
	"context"
	"encoding/json"
	"fmt"
)

type RecallRequest struct {
	Num     int         `json:"num,omitempty"`
	TopK    int         `json:"top_k,omitempty"`
	Filters Filter      `json:"filters,omitempty"`
	Queries [][]float32 `json:"queries,omitempty"`
}

type RecallResponse struct {
	AvgRecall          float64 `json:"avg_recall"`
	AvgExhaustiveCount float64 `json:"avg_exhaustive_count"`
	AvgAnnCount        float64 `json:"avg_ann_count"`
}

func (c *Client) Recall(ctx context.Context, namespace string, request *RecallRequest) (*RecallResponse, error) {
	path := fmt.Sprintf("/v1/vectors/%s/_debug/recall", namespace)
	reqJson, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	respData, err := c.post(ctx, path, reqJson)
	if err != nil {
		return nil, fmt.Errorf("failed to perform recall: %w", err)
	}

	var response RecallResponse
	if err := json.Unmarshal(respData, &response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &response, nil
}
