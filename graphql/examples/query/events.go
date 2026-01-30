package query

import (
	"context"
	"fmt"
	"log"

	"github.com/open-move/sui-go-sdk/graphql"
	"github.com/open-move/sui-go-sdk/utils"
)

// Events demonstrates how to query events.
func Events(ctx context.Context, client *graphql.Client) {
	// Example: Query recent events
	fmt.Println("=== QueryEvents (first 5) ===")
	events, err := client.QueryEvents(ctx, nil, &graphql.PaginationArgs{
		First: utils.Ptr(5),
	})
	if err != nil {
		log.Printf("QueryEvents error: %v", err)
		return
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
