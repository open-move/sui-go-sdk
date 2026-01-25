package graphql

import (
	"context"
	"testing"
	"time"
)

// newTestClient creates a new client for testing purposes.
func newTestClient() *Client {
	return NewClient(
		WithEndpoint(TestnetEndpoint),
		WithTimeout(20*time.Second),
	)
}

// TestExecuteChainIdentifier tests the execution of a query to fetch the chain identifier.
func TestExecuteChainIdentifier(t *testing.T) {
	client := newTestClient()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	const query = `query { chainIdentifier }`

	var resp struct {
		ChainIdentifier string `json:"chainIdentifier"`
	}

	if err := client.Execute(ctx, query, nil, &resp); err != nil {
		t.Fatalf("execute chainIdentifier: %v", err)
	}

	if resp.ChainIdentifier == "" {
		t.Fatalf("empty chainIdentifier")
	}
}

// TestQueryBuilderExecuteEpoch tests the query builder by fetching the epoch's reference gas price.
func TestQueryBuilderExecuteEpoch(t *testing.T) {
	client := newTestClient()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	qb := NewQueryBuilder()
	qb.Field("epoch").Fields("referenceGasPrice").Done()

	var resp struct {
		Epoch struct {
			ReferenceGasPrice string `json:"referenceGasPrice"`
		} `json:"epoch"`
	}

	if err := qb.Execute(ctx, client, &resp); err != nil {
		t.Fatalf("query builder execute epoch: %v", err)
	}

	if resp.Epoch.ReferenceGasPrice == "" {
		t.Fatalf("empty referenceGasPrice")
	}
}
