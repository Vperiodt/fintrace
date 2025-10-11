package graph

import (
	"context"
	"errors"
)

// Client defines the minimal contract required by the repositories to interact
// with the underlying graph database.
type Client interface {
	ExecuteWrite(ctx context.Context, cypher string, params map[string]any) (Result, error)
	ExecuteRead(ctx context.Context, cypher string, params map[string]any) (Result, error)
	VerifyConnectivity(ctx context.Context) error
	Close(ctx context.Context) error
}

// Result is a simplified representation of a query response.
type Result struct {
	Records []Record
}

// Record groups key-value pairs returned from the graph engine.
type Record map[string]any

// Options configures a graph client implementation.
type Options struct {
	URI            string
	Database       string
	Username       string
	Password       string
	MaxConnections int
}

// ErrMissingURI indicates the graph URI is not provided.
var ErrMissingURI = errors.New("graph URI is required")
