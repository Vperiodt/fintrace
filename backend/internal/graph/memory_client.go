package graph

import (
	"context"
	"sync"
)

// MemoryClient is a simple in-memory implementation of the Client interface used
// for unit testing repository logic without requiring a running graph database.
type MemoryClient struct {
	mu           sync.Mutex
	writeCalls   []ExecutedQuery
	readCalls    []ExecutedQuery
	readResults  []Result
	writeResults []Result
	err          error
	connectivity error
}

// ExecutedQuery captures a cypher statement and parameters executed against the graph.
type ExecutedQuery struct {
	Query  string
	Params map[string]any
}

// NewMemoryClient instantiates the in-memory client with optional canned results.
func NewMemoryClient() *MemoryClient {
	return &MemoryClient{}
}

// WithError configures the client to return the provided error for subsequent calls.
func (m *MemoryClient) WithError(err error) *MemoryClient {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.err = err
	return m
}

// WithConnectivityError forces VerifyConnectivity to return the supplied error.
func (m *MemoryClient) WithConnectivityError(err error) *MemoryClient {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connectivity = err
	return m
}

// PushReadResult appends a result that will be returned on the next ExecuteRead call.
func (m *MemoryClient) PushReadResult(res Result) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.readResults = append(m.readResults, res)
}

// PushWriteResult appends a result that will be returned on the next ExecuteWrite call.
func (m *MemoryClient) PushWriteResult(res Result) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.writeResults = append(m.writeResults, res)
}

func (m *MemoryClient) ExecuteWrite(_ context.Context, cypher string, params map[string]any) (Result, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.err != nil {
		return Result{}, m.err
	}

	m.writeCalls = append(m.writeCalls, ExecutedQuery{
		Query:  cypher,
		Params: cloneMap(params),
	})

	if len(m.writeResults) == 0 {
		return Result{}, nil
	}

	res := m.writeResults[0]
	m.writeResults = m.writeResults[1:]
	return res, nil
}

func (m *MemoryClient) ExecuteRead(_ context.Context, cypher string, params map[string]any) (Result, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.err != nil {
		return Result{}, m.err
	}

	m.readCalls = append(m.readCalls, ExecutedQuery{
		Query:  cypher,
		Params: cloneMap(params),
	})

	if len(m.readResults) == 0 {
		return Result{}, nil
	}

	res := m.readResults[0]
	m.readResults = m.readResults[1:]
	return res, nil
}

func (m *MemoryClient) VerifyConnectivity(context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.connectivity
}

func (m *MemoryClient) Close(context.Context) error {
	return nil
}

// WriteCalls returns a snapshot of executed write queries.
func (m *MemoryClient) WriteCalls() []ExecutedQuery {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]ExecutedQuery(nil), m.writeCalls...)
}

// ReadCalls returns a snapshot of executed read queries.
func (m *MemoryClient) ReadCalls() []ExecutedQuery {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]ExecutedQuery(nil), m.readCalls...)
}

func cloneMap(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}
	dst := make(map[string]any, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
