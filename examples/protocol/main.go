// Package main demonstrates protocol and system queries using the Sui GraphQL SDK.
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

	// Example: Get chain identifier
	fmt.Println("=== GetChainIdentifier ===")
	chainID, err := client.GetChainIdentifier(ctx)
	if err != nil {
		log.Printf("GetChainIdentifier error: %v", err)
	} else {
		printJSON("GetChainIdentifier result", chainID)
		fmt.Printf("Chain Identifier: %s\n", chainID)
	}
	fmt.Println()

	// Example: Get reference gas price
	fmt.Println("=== GetReferenceGasPrice ===")
	gasPrice, err := client.GetReferenceGasPrice(ctx)
	if err != nil {
		log.Printf("GetReferenceGasPrice error: %v", err)
	} else if gasPrice != nil {
		printJSON("GetReferenceGasPrice result", gasPrice)
		fmt.Printf("Reference Gas Price: %s MIST\n", *gasPrice)
	}
	fmt.Println()

	// Example: Get service configuration
	fmt.Println("=== GetServiceConfig ===")
	config, err := client.GetServiceConfig(ctx)
	if err != nil {
		log.Printf("GetServiceConfig error: %v", err)
	} else if config != nil {
		printJSON("GetServiceConfig result", config)
		fmt.Printf("Max Query Depth: %d\n", config.MaxQueryDepth)
		fmt.Printf("Max Query Nodes: %d\n", config.MaxQueryNodes)
		fmt.Printf("Max Output Nodes: %d\n", config.MaxOutputNodes)
		fmt.Printf("Query Timeout: %dms\n", config.QueryTimeoutMs)
		fmt.Printf("Max Payload Size: %d bytes\n", config.MaxQueryPayloadSize)
	}
	fmt.Println()
}
