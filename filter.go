package tpuf

import (
	"encoding/json"
)

// Supported operators for filtering.
// See https://turbopuffer.com/docs/query#full-list-of-operators
type Operator string

const (
	OpEq       Operator = "Eq"
	OpNotEq    Operator = "NotEq"
	OpIn       Operator = "In"
	OpNotIn    Operator = "NotIn"
	OpLt       Operator = "Lt"
	OpLte      Operator = "Lte"
	OpGt       Operator = "Gt"
	OpGte      Operator = "Gte"
	OpGlob     Operator = "Glob"
	OpNotGlob  Operator = "NotGlob"
	OpIGlob    Operator = "IGlob"
	OpNotIGlob Operator = "NotIGlob"
)

// Filter represents a Turbopuffer filter.
// This may be a simple filter, such as a single attribute with an operator and value,
// or a more complex filter, such as an "And" or "Or" filter with multiple sub-filters.
// See https://turbopuffer.com/docs/query#filtering-parameters
type Filter interface {
	tpuf_SerializeFilter() interface{}
	json.Marshaler
}

// BaseFilter represents a simple filter with an attribute, operator, and value.
type BaseFilter struct {
	Attribute string
	Operator  Operator
	Value     interface{}
}

func (bf *BaseFilter) tpuf_SerializeFilter() interface{} {
	return []interface{}{bf.Attribute, bf.Operator, bf.Value}
}

func (f *BaseFilter) MarshalJSON() ([]byte, error) {
	if f == nil {
		return []byte("null"), nil
	}
	return json.Marshal(f.tpuf_SerializeFilter())
}

// AndFilter represents a filter that requires all of its sub-filters to be true.
type AndFilter struct {
	Filters []Filter
}

func (af *AndFilter) tpuf_SerializeFilter() interface{} {
	serialized := make([]interface{}, 2)
	serialized[0] = "And"
	subFilters := make([]interface{}, 0, len(af.Filters))
	for _, filter := range af.Filters {
		if filter == nil {
			continue
		}
		subFilters = append(subFilters, filter.tpuf_SerializeFilter())
	}
	serialized[1] = subFilters
	return serialized
}

func (f *AndFilter) MarshalJSON() ([]byte, error) {
	if f == nil {
		return []byte("null"), nil
	}
	return json.Marshal(f.tpuf_SerializeFilter())
}

// OrFilter represents a filter that requires at least one of its sub-filters to be true.
type OrFilter struct {
	Filters []Filter
}

func (of *OrFilter) tpuf_SerializeFilter() interface{} {
	serialized := make([]interface{}, 2)
	serialized[0] = "Or"
	subFilters := make([]interface{}, 0, len(of.Filters))
	for _, filter := range of.Filters {
		if filter == nil {
			continue
		}
		subFilters = append(subFilters, filter.tpuf_SerializeFilter())
	}
	serialized[1] = subFilters
	return serialized
}

func (f *OrFilter) MarshalJSON() ([]byte, error) {
	if f == nil {
		return []byte("null"), nil
	}
	return json.Marshal(f.tpuf_SerializeFilter())
}
