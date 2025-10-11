package graph

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// NewNeo4jClient establishes a Bolt connection using the official Neo4j driver.
// Neptune's openCypher endpoint is wire-compatible with the Bolt protocol,
// allowing the same driver to be reused for both local Neo4j and AWS Neptune.
func NewNeo4jClient(ctx context.Context, opts Options) (Client, error) {
	if opts.URI == "" {
		return nil, ErrMissingURI
	}

	auth := neo4j.NoAuth()
	if opts.Username != "" {
		auth = neo4j.BasicAuth(opts.Username, opts.Password, "")
	}

	driver, err := neo4j.NewDriverWithContext(opts.URI, auth, func(c *neo4j.Config) {
		if opts.MaxConnections > 0 {
			c.MaxConnectionPoolSize = opts.MaxConnections
		}
	})
	if err != nil {
		return nil, fmt.Errorf("create neo4j driver: %w", err)
	}

	if err := driver.VerifyConnectivity(ctx); err != nil {
		_ = driver.Close(ctx)
		return nil, fmt.Errorf("verify graph connectivity: %w", err)
	}

	return &neo4jClient{
		driver:   driver,
		database: opts.Database,
	}, nil
}

type neo4jClient struct {
	driver   neo4j.DriverWithContext
	database string
}

func (c *neo4jClient) ExecuteWrite(ctx context.Context, cypher string, params map[string]any) (Result, error) {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: c.database,
		AccessMode:   neo4j.AccessModeWrite,
	})
	defer session.Close(ctx)

	res, err := session.Run(ctx, cypher, params)
	if err != nil {
		return Result{}, err
	}

	return consumeResult(ctx, res)
}

func (c *neo4jClient) ExecuteRead(ctx context.Context, cypher string, params map[string]any) (Result, error) {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: c.database,
		AccessMode:   neo4j.AccessModeRead,
	})
	defer session.Close(ctx)

	res, err := session.Run(ctx, cypher, params)
	if err != nil {
		return Result{}, err
	}

	return consumeResult(ctx, res)
}

func (c *neo4jClient) VerifyConnectivity(ctx context.Context) error {
	return c.driver.VerifyConnectivity(ctx)
}

func (c *neo4jClient) Close(ctx context.Context) error {
	return c.driver.Close(ctx)
}

func consumeResult(ctx context.Context, res neo4j.ResultWithContext) (Result, error) {
	var records []Record
	for res.Next(ctx) {
		rec := res.Record()
		record := make(Record, len(rec.Keys))
		for _, key := range rec.Keys {
			value, _ := rec.Get(key)
			record[key] = value
		}
		records = append(records, record)
	}
	if err := res.Err(); err != nil {
		return Result{}, err
	}
	return Result{Records: records}, nil
}
