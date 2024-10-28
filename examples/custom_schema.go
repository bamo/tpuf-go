package examples

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/bamo/tpuf-go"
)

/**
 * In this example, we create a namespace with a custom schema and upsert some documents.
 * We then use a query to retrieve documents from the index, and deserialize the attributes.
 *
 * This example is runnable as-is, but you'll need to set the TPUF_API_TOKEN environment variable.
 */
func UpsertAndQueryWithCustomSchema(namespace string) error {
	ctx := context.Background()
	client := &tpuf.Client{
		ApiToken: os.Getenv("TPUF_API_TOKEN"),
	}

	// First, we define the schema for this namespace
	schema := tpuf.Schema{
		"colony_name": &tpuf.Attribute{
			Type: tpuf.AttributeTypeString,
			FullTextSearch: &tpuf.FullTextSearchParams{
				Stemming:        boolPtr(true),
				RemoveStopWords: boolPtr(true),
			},
		},
		"alien_description": &tpuf.Attribute{
			Type:           tpuf.AttributeTypeString,
			FullTextSearch: &tpuf.FullTextSearchParams{},
		},
		"gravity_level": &tpuf.Attribute{
			Type: tpuf.AttributeTypeString,
		},
		"cheese_reserves": &tpuf.Attribute{
			Type: tpuf.AttributeTypeUint,
		},
	}

	// We can use a json-marshalable struct instead of a map[string]interface{} for
	// more structured attributes.
	type ColonyAttrs struct {
		ColonyName       string `json:"colony_name"`
		AlienDescription string `json:"alien_description"`
		GravityLevel     string `json:"gravity_level"`
		CheeseReserves   uint   `json:"cheese_reserves"`
	}

	// Create the upsert request
	request := &tpuf.UpsertRequest{
		DistanceMetric: tpuf.DistanceMetricCosine,
		Schema:         schema,
		Upserts: []*tpuf.Upsert{
			{
				ID:     "c64da516-cb16-4a99-8e0d-450b2c0cd1c2",
				Vector: []float32{0.1, 0.2, 0.3, 0.4},
				Attributes: ColonyAttrs{
					ColonyName:       "Lunar Cheese Factory",
					AlienDescription: "Small green creatures obsessed with dairy products",
					GravityLevel:     "bouncy",
					CheeseReserves:   9001,
				},
			},
			{
				ID:     "76a40b10-47cd-44f7-af15-c5498e48f1d9",
				Vector: []float32{0.2, 0.3, 0.4, 0.5},
				Attributes: ColonyAttrs{
					ColonyName:       "Venusian Sauna Resort",
					AlienDescription: "Heat-resistant blobs that love extreme temperatures",
					GravityLevel:     "crushy",
					CheeseReserves:   42,
				},
			},
			{
				ID:     "df9756fa-39f9-4ef2-8d5d-f012891b93f4",
				Vector: []float32{0.3, 0.4, 0.5, 0.6},
				Attributes: ColonyAttrs{
					ColonyName:       "Floating Noodle Bowl",
					AlienDescription: "Spaghetti-like beings that thrive in zero gravity",
					GravityLevel:     "floaty",
					CheeseReserves:   314,
				},
			},
		},
	}

	// Perform the upsert
	err := client.Upsert(ctx, namespace, request)
	if err != nil {
		return fmt.Errorf("failed to upsert space colonies: %w", err)
	}

	fmt.Println("Space colonies upserted successfully with custom schema")

	// Now, query the space colonies
	// In reality, you would use your favorite embedding model here.
	generateEmbedding := func(text string) ([]float32, error) {
		return []float32{0.1, 0.2, 0.3, 0.4}, nil
	}

	query := "What is the name of our colony on the moon?"
	queryEmbedding, err := generateEmbedding(query)
	if err != nil {
		return fmt.Errorf("failed to generate query embedding: %w", err)
	}

	queryRequest := &tpuf.QueryRequest{
		Vector:            queryEmbedding,
		DistanceMetric:    tpuf.DistanceMetricCosine,
		IncludeAttributes: true,
		TopK:              3,
	}
	results, err := client.Query(ctx, namespace, queryRequest)
	if err != nil {
		return fmt.Errorf("failed to query space colonies: %w", err)
	}

	for _, result := range results {
		fmt.Printf("Colony: %s (distance: %f)\n", result.ID, result.Dist)
		// We can unmarshal the attributes into our structured format.
		var attrs ColonyAttrs
		if err := json.Unmarshal(result.Attributes, &attrs); err != nil {
			return fmt.Errorf("failed to unmarshal attributes: %w", err)
		}
		fmt.Printf("  Attributes: %+v\n", attrs)
	}

	return nil
}

func boolPtr(b bool) *bool {
	return &b
}
