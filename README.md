# tpuf-go

`tpuf-go` is a Golang API client for Turbopuffer (turbopuffer.com), a vector database and search engine. This client provides an easy-to-use interface for interacting with the Turbopuffer API, allowing you to perform operations such as upserting documents, querying data, and managing namespaces.

## Installation

To install the `tpuf-go` package, use the following command:

```go get github.com/bamo/tpuf-go```

## Initializing a Client

To use the Turbopuffer API, you first need to initialize a client with your API token:

```go
import "github.com/bamo/tpuf-go"

client := &tpuf.Client{
    ApiToken: "your-api-token-here",
}
```

## Upserting Documents

The `Upsert` method allows you to create or update documents in a namespace. Here's an example of how to use it:

```go
namespace := "my-namespace"
// Define the schema for this namespace.  Optional unless you need full-text
// search, UUIDs, or other non-default behavior.
schema := tpuf.Schema{
        "title": &tpuf.Attribute{
            Type: tpuf.AttributeTypeString,
            // Full-text search enabled, with custom behavior.
            FullTextSearch: &tpuf.FullTextSearchParams{
                Stemming: true,
            },
        },
        "text": &tpuf.Attribute{
            Type: tpuf.AttributeTypeString,
            // Full-text search enabled, but with all default behavior.
            FullTextSearch: &tpuf.FullTextSearchParams{},
        },
        "category": &tpuf.Attribute{
            Type: tpuf.AttributeTypeString,
        },
}
request := &tpuf.UpsertRequest{
    DistanceMetric: tpuf.DistanceMetricCosine,
    Schema: schema,
    Upserts: []*tpuf.Upsert{
        {
            ID:     "doc1",
            Vector: []float32{0.1, 0.2, 0.3},
            Attributes: map[string]interface{}{
                "title":       "Sample Document",
                "description": "This is a sample document for demonstration purposes.",
                "category":    "example",
            },
        },
    },
}

err := client.Upsert(context.Background(), namespace, request)
if err != nil {
    return err
}
```

In this example, we're upserting a document with an ID, vector, and attributes. We're also defining a schema for the namespace, specifying that "title" and "description" should be full-text searchable.

## Querying Documents

The `Query` method allows you to search for documents using various methods. Here are examples of different types of queries:

### Vector Search

Example: searchfor the 5 closest vectors to the given vector using ANN search with the cosine distance metric.

```go
request := &tpuf.QueryRequest{
    Vector:         []float32{0.1, 0.2, 0.3},
    DistanceMetric: tpuf.DistanceMetricCosine,
    TopK:           5,
}

results, err := client.Query(context.Background(), namespace, request)
if err != nil {
    return nil, err
}

// Use results...
```

### BM25 Full Text Search

Example: perform a full-text search on the "text" field for the phrase "What is the capital of the moon?", returning the top 3 results.

```go
request := &tpuf.QueryRequest{
    RankBy: []interface{}{"text", "BM25", "What is the capital of the moon?"},
    TopK:   3,
}

results, err := client.Query(context.Background(), namespace, request)
if err != nil {
    return nil, err
}

// Use results...
```

### Filter-only Search

Example: retrieve up to 10 documents where the "category" is "example".  More filters must be used to paginate the results once the first page is retrieved.

```go
request := &tpuf.QueryRequest{
    Filters: &tpuf.AndFilter{
        Filters: []tpuf.Filter{
            &tpuf.BaseFilter{Attribute: "category", Operator: tpuf.OpEq, Value: "example"},
        },
    },
    // Return the first 10 matching results, ordered by ID ascending.
    TopK: 10,
}

results, err := client.Query(context.Background(), namespace, request)
if err != nil {
    return nil, err
}

// Use results...
```

## More Information

For more detailed information about the available methods and their parameters, please refer to the package documentation and the [Turbopuffer API documentation](https://turbopuffer.com/docs).

## License

This project is licensed under the MIT License. See the LICENSE file for details.