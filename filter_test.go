package tpuf_test

import (
	"encoding/json"
	"testing"

	"github.com/bamo/tpuf-go"
	"github.com/stretchr/testify/assert"
)

func TestMarshalFilter(t *testing.T) {
	tests := []struct {
		name     string
		filter   tpuf.Filter
		expected string
	}{
		{
			name: "Basic filter",
			filter: &tpuf.BaseFilter{
				Attribute: "attr",
				Operator:  tpuf.OpEq,
				Value:     42,
			},
			expected: `["attr","Eq",42]`,
		},
		{
			name: "Multi-value filter",
			filter: &tpuf.BaseFilter{
				Attribute: "ids",
				Operator:  tpuf.OpIn,
				Value:     []int{1, 2, 3},
			},
			expected: `["ids","In",[1,2,3]]`,
		},
		{
			name: "AndFilter",
			filter: &tpuf.AndFilter{
				Filters: []tpuf.Filter{
					&tpuf.BaseFilter{
						Attribute: "attr1",
						Operator:  tpuf.OpEq,
						Value:     "value1",
					},
					&tpuf.BaseFilter{
						Attribute: "attr2",
						Operator:  tpuf.OpGt,
						Value:     10,
					},
				},
			},
			expected: `["And",[["attr1","Eq","value1"],["attr2","Gt",10]]]`,
		},
		{
			name: "OrFilter",
			filter: &tpuf.OrFilter{
				Filters: []tpuf.Filter{
					&tpuf.BaseFilter{
						Attribute: "attr1",
						Operator:  tpuf.OpEq,
						Value:     "value1",
					},
					&tpuf.BaseFilter{
						Attribute: "attr2",
						Operator:  tpuf.OpLt,
						Value:     20,
					},
				},
			},
			expected: `["Or",[["attr1","Eq","value1"],["attr2","Lt",20]]]`,
		},
		{
			name: "Nested compound filter",
			filter: &tpuf.AndFilter{
				Filters: []tpuf.Filter{
					&tpuf.BaseFilter{
						Attribute: "id",
						Operator:  tpuf.OpIn,
						Value:     []int{1, 2, 3},
					},
					&tpuf.BaseFilter{
						Attribute: "key1",
						Operator:  tpuf.OpEq,
						Value:     "one",
					},
					&tpuf.BaseFilter{
						Attribute: "filename",
						Operator:  tpuf.OpNotGlob,
						Value:     "/vendor/**",
					},
					&tpuf.OrFilter{
						Filters: []tpuf.Filter{
							&tpuf.BaseFilter{
								Attribute: "filename",
								Operator:  tpuf.OpGlob,
								Value:     "**.tsx",
							},
							&tpuf.BaseFilter{
								Attribute: "filename",
								Operator:  tpuf.OpGlob,
								Value:     "**.js",
							},
						},
					},
				},
			},
			expected: `["And",[["id","In",[1,2,3]],["key1","Eq","one"],["filename","NotGlob","/vendor/**"],["Or",[["filename","Glob","**.tsx"],["filename","Glob","**.js"]]]]]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := json.Marshal(tt.filter)
			assert.NoError(t, err)

			// Compare JSON strings
			var expectedJSON, resultJSON interface{}
			err = json.Unmarshal([]byte(tt.expected), &expectedJSON)
			assert.NoError(t, err)
			err = json.Unmarshal(result, &resultJSON)
			assert.NoError(t, err)

			assert.Equal(t, expectedJSON, resultJSON)
		})
	}

	t.Run("filter in struct", func(t *testing.T) {
		type TestStruct struct {
			Filter tpuf.Filter `json:"filter"`
		}

		testStruct := TestStruct{
			Filter: &tpuf.BaseFilter{
				Attribute: "id",
				Operator:  tpuf.OpIn,
				Value:     []int{1, 2, 3},
			},
		}

		result, err := json.Marshal(testStruct)
		assert.NoError(t, err)
		assert.Equal(t, `{"filter":["id","In",[1,2,3]]}`, string(result))
	})
}
