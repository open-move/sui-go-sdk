package query

import (
	"context"
	"fmt"
	"log"

	"github.com/open-move/sui-go-sdk/graphql"
)

func QueryBuilder(ctx context.Context, client *graphql.Client) {
	// Example 1: Simple query with single field
	fmt.Println("=== Simple Query: Chain Identifier ===")
	{
		qb := graphql.NewQueryBuilder()
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
		qb := graphql.NewQueryBuilder()
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
		qb := graphql.NewQueryBuilder()
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
			fmt.Printf("Epoch: %d\n", resp.Epoch.EpochID)
			fmt.Printf("Checkpoints: %d\n", len(resp.Epoch.Checkpoints.Nodes))
			printJSON("Nested query response", resp)
		}
	}
}
