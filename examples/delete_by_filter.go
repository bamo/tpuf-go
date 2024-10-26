package examples

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/bamo/tpuf-go"
)

func DeleteByFilter() error {
	ctx := context.Background()
	client := &tpuf.Client{
		ApiToken: os.Getenv("TPUF_API_TOKEN"),
	}

	namespace := "my-test-namespace"

	// Define a filter to delete incriminating evidence
	baseFilter := &tpuf.BaseFilter{
		Attribute: "category",
		Operator:  tpuf.OpEq,
		Value:     "incriminating",
	}

	var filter tpuf.Filter = baseFilter

	pageSize := 1000
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
	return nil
}
