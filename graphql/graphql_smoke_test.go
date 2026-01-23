package graphql

import (
	"context"
	"testing"
	"time"
)

func newTestClient() *Client {
	return NewClient(
		WithEndpoint(TestnetEndpoint),
		WithTimeout(20*time.Second),
	)
}

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
