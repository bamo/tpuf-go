package tpuf_test

import (
	"encoding/json"
	"testing"

	"github.com/bamo/tpuf-go"
	"github.com/stretchr/testify/assert"
)

func TestSchemaMarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		schema   tpuf.Schema
		expected string
	}{
		{
			name: "Basic schema with full text search and UUID",
			schema: tpuf.Schema{
				"text": &tpuf.Attribute{
					Type: tpuf.AttributeTypeString,
					FullTextSearch: &tpuf.FullTextSearchParams{
						Language:        "english",
						Stemming:        boolPtr(false),
						RemoveStopWords: boolPtr(true),
						CaseSensitive:   boolPtr(false),
					},
				},
				"relatedID": &tpuf.Attribute{
					Type: tpuf.AttributeTypeUUID,
				},
			},
			expected: `{"text":{"type":"string","full_text_search":{"language":"english","stemming":false,"remove_stop_words":true,"case_sensitive":false}},"relatedID":{"type":"uuid"}}`,
		},
		{
			name: "Schema with filterable attribute",
			schema: tpuf.Schema{
				"age": &tpuf.Attribute{
					Type:       tpuf.AttributeTypeUint,
					Filterable: boolPtr(true),
				},
			},
			expected: `{"age":{"type":"uint","filterable":true}}`,
		},
		{
			name:     "Empty schema",
			schema:   tpuf.Schema{},
			expected: `{}`,
		},
		{
			name: "Attribute with only type specified",
			schema: tpuf.Schema{
				"name": &tpuf.Attribute{
					Type: tpuf.AttributeTypeString,
				},
			},
			expected: `{"name":{"type":"string"}}`,
		},
		{
			name: "Attribute with empty FullTextSearch",
			schema: tpuf.Schema{
				"description": &tpuf.Attribute{
					Type:           tpuf.AttributeTypeString,
					FullTextSearch: &tpuf.FullTextSearchParams{},
				},
			},
			expected: `{"description":{"type":"string","full_text_search":{}}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			marshaled, err := json.Marshal(tt.schema)
			assert.NoError(t, err, "Failed to marshal schema")

			assert.JSONEq(t, tt.expected, string(marshaled))

			var unmarshaled tpuf.Schema
			err = json.Unmarshal(marshaled, &unmarshaled)
			assert.NoError(t, err, "Failed to unmarshal schema")

			assert.Equal(t, tt.schema, unmarshaled)
		})
	}
}

// Helper function to create a pointer to a bool
func boolPtr(b bool) *bool {
	return &b
}
