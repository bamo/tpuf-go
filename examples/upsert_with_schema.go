package examples

import (
	"context"
	"fmt"
	"os"

	"github.com/bamo/tpuf-go"
)

func UpsertWithSchema() error {
	ctx := context.Background()
	client := &tpuf.Client{
		ApiToken: os.Getenv("TPUF_API_TOKEN"),
	}

	namespace := "my-test-namespace"

	// Define the schema for this namespace
	schema := tpuf.Schema{
		"title": &tpuf.Attribute{
			Type: tpuf.AttributeTypeString,
			FullTextSearch: &tpuf.FullTextSearchParams{
				Stemming:        boolPtr(true),
				RemoveStopWords: boolPtr(true),
			},
		},
		"description": &tpuf.Attribute{
			Type:           tpuf.AttributeTypeString,
			FullTextSearch: &tpuf.FullTextSearchParams{},
		},
		"category": &tpuf.Attribute{
			Type: tpuf.AttributeTypeString,
		},
		"price": &tpuf.Attribute{
			Type: tpuf.AttributeTypeUint,
		},
	}

	// Create the upsert request
	request := &tpuf.UpsertRequest{
		DistanceMetric: tpuf.DistanceMetricCosine,
		Schema:         schema,
		Upserts: []*tpuf.Upsert{
			{
				ID:     "product1",
				Vector: []float32{0.1, 0.2, 0.3, 0.4},
				Attributes: map[string]interface{}{
					"title":       "Ergonomic Office Chair",
					"description": "Comfortable chair with lumbar support for long working hours",
					"category":    "furniture",
					"price":       199,
				},
			},
			{
				ID:     "product2",
				Vector: []float32{0.2, 0.3, 0.4, 0.5},
				Attributes: map[string]interface{}{
					"title":       "Wireless Noise-Canceling Headphones",
					"description": "High-quality audio with active noise cancellation",
					"category":    "electronics",
					"price":       299,
				},
			},
		},
	}

	// Perform the upsert
	err := client.Upsert(ctx, namespace, request)
	if err != nil {
		return fmt.Errorf("failed to upsert documents: %w", err)
	}

	fmt.Println("Documents upserted successfully with custom schema")
	return nil
}

func boolPtr(b bool) *bool {
	return &b
}
