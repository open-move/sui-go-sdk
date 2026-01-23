// Package main demonstrates event queries using the Sui GraphQL SDK.
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

	// Example: Query recent events
	fmt.Println("=== QueryEvents (first 5) ===")
	events, err := client.QueryEvents(ctx, nil, &PaginationArgs{
		First: Ptr(5),
	})
	if err != nil {
		log.Fatalf("QueryEvents error: %v", err)
	}

	fmt.Printf("Retrieved %d events\n", len(events.Nodes))
	for i, event := range events.Nodes {
		fmt.Printf("\nEvent %d:\n", i+1)
		if event.Sender != nil {
			fmt.Printf("  Sender: %s\n", event.Sender.Address)
		}
		if event.Timestamp != nil {
			fmt.Printf("  Timestamp: %s\n", *event.Timestamp)
		}
		if event.TransactionModule != nil {
			fmt.Printf("  Module: %s\n", event.TransactionModule.Name)
		}
	}
	fmt.Println()

	printJSON("QueryEvents result", events)

	if len(events.Nodes) > 0 {
		printJSON("First event", events.Nodes[0])
	}
}
