package examples

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/bamo/tpuf-go"
)

/**
 * Here we use a filter-only query combined with pagination to delete documents matching a given
 * filter from an index.
 *
 * This example is runnable as-is, but you'll need to set the TPUF_API_TOKEN environment variable.
 */
func DeleteByFilter(namespace string) error {
	ctx := context.Background()
	client := &tpuf.Client{
		ApiToken: os.Getenv("TPUF_API_TOKEN"),
	}

	// First, upsert some documents, including some incriminating evidence.
	type DocAttrs struct {
		Category    string `json:"category"`
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	docs := []*tpuf.Upsert{
		{
			ID:     "doc1",
			Vector: []float32{0.1, 0.2, 0.3},
			Attributes: DocAttrs{
				Category:    "incriminating",
				Title:       "Secret Meeting Notes",
				Description: "Details of the midnight cheese heist",
			},
		},
		{
			ID:     "doc2",
			Vector: []float32{0.2, 0.3, 0.4},
			Attributes: DocAttrs{
				Category:    "incriminating",
				Title:       "Hidden Account Numbers",
				Description: "Offshore moon cheese accounts",
			},
		},
		{
			ID:     "doc3",
			Vector: []float32{0.3, 0.4, 0.5},
			Attributes: DocAttrs{
				Category:    "incriminating",
				Title:       "Operation Cheddar",
				Description: "Plans for the dairy domination scheme",
			},
		},
		{
			ID:     "doc4",
			Vector: []float32{0.4, 0.5, 0.6},
			Attributes: DocAttrs{
				Category:    "normal",
				Title:       "Grocery List",
				Description: "Just regular groceries",
			},
		},
		{
			ID:     "doc5",
			Vector: []float32{0.5, 0.6, 0.7},
			Attributes: DocAttrs{
				Category:    "normal",
				Title:       "Weekend Plans",
				Description: "Normal weekend activities",
			},
		},
	}

	err := client.Upsert(ctx, namespace, &tpuf.UpsertRequest{
		Schema: tpuf.Schema{
			"category": &tpuf.Attribute{
				Type: tpuf.AttributeTypeString,
			},
			"title": &tpuf.Attribute{
				Type: tpuf.AttributeTypeString,
			},
			"description": &tpuf.Attribute{
				Type: tpuf.AttributeTypeString,
			},
		},
		Upserts: docs,
	})
	if err != nil {
		return fmt.Errorf("failed to upsert documents: %w", err)
	}

	// Define a filter to delete incriminating evidence
	baseFilter := &tpuf.BaseFilter{
		Attribute: "category",
		Operator:  tpuf.OpEq,
		Value:     "incriminating",
	}

	var filter tpuf.Filter = baseFilter

	pageSize := 2
	deletedCount := 0

	for {
		// Use paginated filter-only search to retrieve 1000 results at a time.
		results, err := client.Query(ctx, namespace, &tpuf.QueryRequest{
			TopK:    pageSize,
			Filters: filter,
		})
		if err != nil {
			// Check if the error is due to namespace not found
			apiErr := &tpuf.ApiError{}
			if errors.As(err, apiErr) && apiErr.HttpStatus == 404 {
				fmt.Println("Namespace not found. Deletion process complete.")
				return nil
			}
			return fmt.Errorf("failed to query documents: %w", err)
		}

		if len(results) == 0 {
			break
		}

		idsToDelete := make([]string, len(results))
		for i, result := range results {
			idsToDelete[i] = result.ID
		}

		// Delete this batch of documents by ID.
		err = client.Delete(ctx, namespace, idsToDelete)
		if err != nil {
			return fmt.Errorf("failed to delete documents: %w", err)
		}

		deletedCount += len(idsToDelete)
		fmt.Printf("Deleted %d documents\n", deletedCount)

		// Update the filter to move on to the next batch of documents.
		filter = &tpuf.AndFilter{
			Filters: []tpuf.Filter{
				baseFilter,
				&tpuf.BaseFilter{
					Attribute: "id",
					Operator:  tpuf.OpGt,
					Value:     results[len(results)-1].ID,
				},
			},
		}
	}
	fmt.Printf("Deletion complete. Total documents deleted: %d\n", deletedCount)

	results, err := client.Query(ctx, namespace, &tpuf.QueryRequest{
		TopK:              1000,
		IncludeAttributes: true,
	})
	if err != nil {
		return fmt.Errorf("failed to query documents: %w", err)
	}
	fmt.Printf("Remaining documents:\n")
	for _, result := range results {
		fmt.Printf("%s\n", string(result.Attributes))
	}
	return nil
}
