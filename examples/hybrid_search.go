package examples

import (
	"context"
	"fmt"
	"os"

	"github.com/bamo/tpuf-go"
)

func HybridSearch() error {
	ctx := context.Background()
	client := &tpuf.Client{
		ApiToken: os.Getenv("TPUF_API_TOKEN"),
	}

	namespace := "my-test-namespace"

	query := "What is the capital of the moon?"

	// Replace with your favorite embedding model.
	queryEmbedding := []float32{0.1, 0.2, 0.3}

	// Retrieve 10 results using BM25 full-text search.
	keywordResults, err := client.Query(ctx, namespace, &tpuf.QueryRequest{
		RankBy: []interface{}{"text", "BM25", query},
		TopK:   10,
	})
	if err != nil {
		return err
	}

	// Retrieve 10 results using semantic search with cosine distance metric.
	semanticResults, err := client.Query(ctx, namespace, &tpuf.QueryRequest{
		Vector:         queryEmbedding,
		DistanceMetric: tpuf.DistanceMetricCosine,
		TopK:           10,
	})
	if err != nil {
		return err
	}

	allResults := append(keywordResults, semanticResults...)

	// Do some reranking here...

	for _, result := range allResults {
		fmt.Printf("%+v\n", result)
	}

	return nil
}
