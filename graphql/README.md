# Sui GraphQL Go SDK

The Sui GraphQL Go SDK provides a robust and type-safe way to interact with the Sui blockchain using its GraphQL API. This module allows you to query data, execute transactions, and build complex queries using a fluent API.

## Features

- **Full GraphQL Support**: Access all Sui GraphQL RPC methods.
- **Fluent Query Builder**: Construct complex queries dynamically with type safety.
- **Mutation Builder**: Easily build and execute transactions.
- **Type-Safe Responses**: Automatically map GraphQL responses to Go structs.
- **Custom Queries**: Support for raw GraphQL queries with variable substitution.
- **Pagination**: Built-in support for connection-based pagination.

## Installation

```bash
go get github.com/open-move/sui-go-sdk
```

## Quick Start

### Connecting to the Sui Network

You can connect to the Sui network by creating a new client. By default, it connects to Mainnet, but you can configure it for Testnet, Devnet, or a local node.

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/open-move/sui-go-sdk/graphql"
)

func main() {
	// Connect to Mainnet (default)
	client := graphql.NewClient()

	// Or connect to Testnet
	// client := graphql.NewClient(graphql.WithEndpoint(graphql.TestnetEndpoint))

	// Or connect with custom timeout
	// client := graphql.NewClient(
	//     graphql.WithEndpoint(graphql.MainnetEndpoint),
	//     graphql.WithTimeout(10 * time.Second),
	// )
}
```

### Examples

#### Querying Balances

Fetch all balances for a specific address.

```go
func GetBalances(ctx context.Context, client *graphql.Client, address string) {
	balances, err := client.GetAllBalances(ctx, graphql.SuiAddress(address))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for _, b := range balances {
		if b.CoinType != nil {
			fmt.Printf("Coin: %s, Balance: %s\n", b.CoinType.Repr, b.TotalBalance)
		}
	}
}
```

#### Querying Objects

Retrieve object details by ID.

```go
func GetObject(ctx context.Context, client *graphql.Client, objectID string) {
	obj, err := client.GetObject(ctx, graphql.SuiAddress(objectID), &graphql.ObjectDataOptions{
		ShowType:    true,
		ShowContent: true,
		ShowOwner:   true,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if obj != nil {
		fmt.Printf("Object Type: %v\n", obj.Type)
	}
}
```

#### Using the Query Builder

Construct dynamic queries using the fluent Query Builder API.

```go
func BuildQuery(ctx context.Context, client *graphql.Client) {
	qb := graphql.NewQueryBuilder()
	qb.Field("chainIdentifier").Done()

	var resp struct {
		ChainIdentifier string `json:"chainIdentifier"`
	}

	if err := qb.Execute(ctx, client, &resp); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Chain ID: %s\n", resp.ChainIdentifier)
}
```

#### Executing Custom Queries

Execute raw GraphQL queries for maximum flexibility.

```go
func CustomQuery(ctx context.Context, client *graphql.Client) {
	query := `
        query getChainId {
            chainIdentifier
        }
    `
	var result struct {
		ChainIdentifier string `json:"chainIdentifier"`
	}

	err := client.Query(ctx, query, nil, &result)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	fmt.Printf("Chain ID: %s\n", result.ChainIdentifier)
}
```

#### Querying Transactions

Retrieve transaction details by digest or query multiple transactions.

```go
func GetTransaction(ctx context.Context, client *graphql.Client, digest string) {
	// Get a single transaction
	tx, err := client.GetTransactionBlock(ctx, digest, &graphql.TransactionBlockOptions{
		ShowInput:   true,
		ShowEffects: true,
		ShowEvents:  true,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Status: %v\n", tx.Effects.Status)

	// Query transactions by sender
	filter := &graphql.TransactionFilter{
		SentAddress: graphql.Ptr(graphql.SuiAddress("0xSenderAddress")),
	}
	txs, err := client.QueryTransactionBlocks(ctx, filter, &graphql.PaginationArgs{
		First: graphql.Ptr(10),
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Found %d transactions\n", len(txs.Nodes))
}
```

#### Querying Events

Fetch events emitted by transactions.

```go
func GetEvents(ctx context.Context, client *graphql.Client) {
	// Query events by sender
	filter := &graphql.EventFilter{
		Sender: graphql.Ptr(graphql.SuiAddress("0xSenderAddress")),
	}
	events, err := client.QueryEvents(ctx, filter, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for _, event := range events.Nodes {
		fmt.Printf("Event type: %v\n", event.Contents.Type.Repr)
	}
}
```

#### Managing Coins

Retrieve specific coins for an address.

```go
func GetCoins(ctx context.Context, client *graphql.Client, address string) {
	// Get SUI coins
	coins, err := client.GetCoins(ctx, graphql.SuiAddress(address), nil, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for _, coin := range coins.Nodes {
		fmt.Printf("Coin ID: %s, Balance: %v\n", coin.Address, coin.Contents)
	}
}
```

#### Dynamic Fields

Access dynamic fields of an object.

```go
func GetDynamicField(ctx context.Context, client *graphql.Client, parentID string) {
	// Get a specific dynamic field
	name := graphql.DynamicFieldName{
		Type:  "0x2::kiosk::ItemKey",
		Value: "...", // BCS encoded value or JSON
	}
	field, err := client.GetDynamicFieldObject(ctx, graphql.SuiAddress(parentID), name)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Field Value: %v\n", field.Value)
}
```

#### Simulating Transactions

Simulate a transaction to estimate gas and check effects before execution.

```go
func SimulateTx(ctx context.Context, client *graphql.Client, txBytes []byte) {
	// Simulate transaction from bytes
	result, err := graphql.SimulateTransactionBcs(ctx, client, graphql.Base64(txBytes), nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if result.Effects.Status.Status == "success" {
		fmt.Printf("Gas Cost: %v\n", result.Effects.GasEffects.GasSummary.ComputationCost)
	}
}
```

#### Executing Transactions

Execute a signed transaction.

```go
func ExecuteTx(ctx context.Context, client *graphql.Client, txBytes []byte, signatures []string) {
	// Convert signatures to Base64
	sigs := make([]graphql.Base64, len(signatures))
	for i, s := range signatures {
		sigs[i] = graphql.Base64(s)
	}

	// Execute
	result, err := graphql.ExecuteTransaction(client, ctx, graphql.Base64(txBytes), sigs)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Digest: %s\n", result.Effects.Digest)
}
```

#### Pagination

Handle large datasets using pagination.

```go
func IterateCoins(ctx context.Context, client *graphql.Client, address string) {
	var cursor *string
	for {
		coins, err := client.GetCoins(ctx, graphql.SuiAddress(address), nil, &graphql.PaginationArgs{
			First: graphql.Ptr(10),
			After: cursor,
		})
		if err != nil {
			break
		}

		// Process coins...

		if !coins.PageInfo.HasNextPage {
			break
		}
		cursor = coins.PageInfo.EndCursor
	}
}
```

## Documentation

For more detailed examples, check the `examples/` directory.
