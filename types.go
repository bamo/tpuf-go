package tpuf

// DistanceMetric represents the available distance functions used to calculate vector similarity.
type DistanceMetric string

const (
	DistanceMetricCosine    DistanceMetric = "cosine_distance"
	DistanceMetricEuclidean DistanceMetric = "euclidean_squared"
)

// AttributeType is the data type of an attribute.
type AttributeType string

const (
	AttributeTypeString      AttributeType = "string"
	AttributeTypeUint        AttributeType = "uint"
	AttributeTypeUUID        AttributeType = "uuid"
	AttributeTypeBool        AttributeType = "bool"
	AttributeTypeStringArray AttributeType = "[]string"
	AttributeTypeUintArray   AttributeType = "[]uint"
	AttributeTypeUUIDArray   AttributeType = "[]uuid"
)

type FullTextSearchParams struct {
	// Language determines language-aware stemming and stopword removal. Default is english.
	// See https://turbopuffer.com/docs/schema#supported-languages-for-full-text-search
	Language string `json:"language,omitempty"`
	// Whether to apply language-specific stemming. Default is false.
	Stemming *bool `json:"stemming,omitempty"`
	// Whether to remove common stop words. Default is true.
	RemoveStopWords *bool `json:"remove_stop_words,omitempty"`
	// Whether searching is case-sensitive. Default is false.
	CaseSensitive *bool `json:"case_sensitive,omitempty"`
}

// Attribute represents a single document attribute.
type Attribute struct {
	// Type is the type of the attribute.  Can usually be inferred, except for UUIDs.
	Type AttributeType `json:"type"`
	// Whether this attribute is filterable.  Default is true unless full text search is enabled for this attribute.
	Filterable *bool `json:"filterable,omitempty"`
	// Whether this attribute is full text searchable using BM25.  Defaults to disabled.
	// For behavior consistent with full_text_search=true, simply use empty FullTextSearchParams.
	FullTextSearch *FullTextSearchParams `json:"full_text_search,omitempty"`
}

// Schema represents the schema of a namespace. Allows customization of document attributes.
// See https://turbopuffer.com/docs/schema
type Schema map[string]*Attribute
