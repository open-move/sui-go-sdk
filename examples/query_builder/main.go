// Package main demonstrates the QueryBuilder for custom GraphQL queries.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/open-move/sui-go-sdk/graphql"
)

func printJSON(label string, v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Printf("JSON marshal error for %s: %v", label, err)
		return
	}
	fmt.Printf("%s:\n%s\n", label, string(data))
}

func main() {
	client := NewClient(
		WithEndpoint(TestnetEndpoint),
		WithTimeout(30*time.Second),
	)

	ctx := context.Background()

	// Example 1: Simple query with single field
	fmt.Println("=== Simple Query: Chain Identifier ===")
	{
		qb := NewQueryBuilder()
		qb.Field("chainIdentifier").Done()

		var resp struct {
			ChainIdentifier string `json:"chainIdentifier"`
		}

		if err := qb.Execute(ctx, client, &resp); err != nil {
			log.Printf("Error: %v", err)
		} else {
			fmt.Printf("Chain ID: %s\n", resp.ChainIdentifier)
			printJSON("Simple query response", resp)
		}
	}
	fmt.Println()

	// Example 2: Query with multiple fields
	fmt.Println("=== Query with Fields: Current Epoch ===")
	{
		qb := NewQueryBuilder()
		qb.Field("epoch").
			Fields("epochId", "referenceGasPrice", "startTimestamp").
			Done()

		var resp struct {
			Epoch struct {
				EpochID           uint64 `json:"epochId"`
				ReferenceGasPrice string `json:"referenceGasPrice"`
				StartTimestamp    string `json:"startTimestamp"`
			} `json:"epoch"`
		}

		if err := qb.Execute(ctx, client, &resp); err != nil {
			log.Printf("Error: %v", err)
		} else {
			fmt.Printf("Epoch: %d\n", resp.Epoch.EpochID)
			fmt.Printf("Gas Price: %s\n", resp.Epoch.ReferenceGasPrice)
			fmt.Printf("Started: %s\n", resp.Epoch.StartTimestamp)
			printJSON("Current epoch query response", resp)
		}
	}
	fmt.Println()

	// Example 3: Nested query with arguments
	fmt.Println("=== Nested Query: Epoch Checkpoints ===")
	{
		qb := NewQueryBuilder()
		qb.Field("epoch").
			Fields("epochId").
			SubField("checkpoints").
			Arg("first", 3).
			SubField("nodes").
			Fields("sequenceNumber", "digest", "timestamp").
			End().
			End().
			Done()

		var resp struct {
			Epoch struct {
				EpochID     uint64 `json:"epochId"`
				Checkpoints struct {
					Nodes []struct {
						SequenceNumber uint64 `json:"sequenceNumber"`
						Digest         string `json:"digest"`
						Timestamp      string `json:"timestamp"`
					} `json:"nodes"`
				} `json:"checkpoints"`
			} `json:"epoch"`
		}

		if err := qb.Execute(ctx, client, &resp); err != nil {
			log.Printf("Error: %v", err)
		} else {
			fmt.Printf("Epoch %d Checkpoints:\n", resp.Epoch.EpochID)
			for _, cp := range resp.Epoch.Checkpoints.Nodes {
				fmt.Printf("  - #%d: %s\n", cp.SequenceNumber, cp.Digest[:16]+"...")
			}
			printJSON("Epoch checkpoints query response", resp)
		}
	}
	fmt.Println()

	// Example 4: Query with variables
	fmt.Println("=== Query with Variables: Coin Metadata ===")
	{
		qb := NewQueryBuilder()
		coinTypeVar := qb.Variable("coinType", "String!", "0x2::sui::SUI")

		qb.Field("coinMetadata").
			ArgVar("coinType", coinTypeVar).
			Fields("name", "symbol", "decimals", "description").
			Done()

		var resp struct {
			CoinMetadata struct {
				Name        string `json:"name"`
				Symbol      string `json:"symbol"`
				Decimals    int    `json:"decimals"`
				Description string `json:"description"`
			} `json:"coinMetadata"`
		}

		if err := qb.Execute(ctx, client, &resp); err != nil {
			log.Printf("Error: %v", err)
		} else {
			fmt.Printf("Coin: %s (%s)\n", resp.CoinMetadata.Name, resp.CoinMetadata.Symbol)
			fmt.Printf("Decimals: %d\n", resp.CoinMetadata.Decimals)
			fmt.Printf("Description: %s\n", resp.CoinMetadata.Description)
			printJSON("Coin metadata query response", resp)
		}
	}
	fmt.Println()

	// Example 5: Build query string (for debugging)
	fmt.Println("=== Built Query String ===")
	{
		qb := NewQueryBuilder().Name("GetEpochInfo")
		qb.Field("epoch").
			Fields("epochId", "referenceGasPrice").
			SubField("checkpoints").
			Arg("first", 2).
			SubField("nodes").
			Fields("sequenceNumber").
			End().
			End().
			Done()

		query, vars := qb.Build()
		fmt.Println("Query:")
		fmt.Println(query)
		printJSON("Variables", vars)
	}
}
