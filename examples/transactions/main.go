package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/open-move/sui-go-sdk/graphql"
	"github.com/open-move/sui-go-sdk/keypair"
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
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Get private key from environment and derive address
	privateKey := os.Getenv("SUI_PRIVATE_KEY")
	if privateKey == "" {
		log.Fatal("SUI_PRIVATE_KEY environment variable is required")
	}

	kp, err := keypair.FromBech32(privateKey)
	if err != nil {
		log.Fatalf("Failed to parse private key: %v", err)
	}

	address, err := kp.SuiAddress()
	if err != nil {
		log.Fatalf("Failed to get address: %v", err)
	}

	fmt.Printf("Using address: %s\n\n", address)

	client := NewClient(
		WithEndpoint(TestnetEndpoint),
		WithTimeout(30*time.Second),
	)

	ctx := context.Background()

	fmt.Println("=== GetTransactionBlock ===")
	digest := "3KmTo5yvbkeg9mrQafUkfFoYVRt45zTg2SoGWBJw615V"
	options := &TransactionBlockOptions{
		ShowInput:          true,
		ShowRawInput:       true,
		ShowEffects:        true,
		ShowEvents:         true,
		ShowObjectChanges:  true,
		ShowBalanceChanges: true,
	}

	tx, err := client.GetTransactionBlock(ctx, digest, options)
	if err != nil {
		log.Printf("GetTransactionBlock error: %v", err)
	} else if tx != nil {
		printJSON("GetTransactionBlock result", tx)
	}
	fmt.Println()

	fmt.Println("=== GetMultipleTransactionBlocks ===")
	digests := []string{
		"3KmTo5yvbkeg9mrQafUkfFoYVRt45zTg2SoGWBJw615V",
	}

	txs, err := client.GetMultipleTransactionBlocks(ctx, digests, nil)
	if err != nil {
		log.Printf("GetMultipleTransactionBlocks error: %v", err)
	} else {
		printJSON("GetMultipleTransactionBlocks result", txs)
	}
	fmt.Println()

	fmt.Println("=== QueryTransactionBlocks ===")
	filter := &TransactionFilter{
		SentAddress: Ptr(SuiAddress(address)),
	}

	txsConnection, err := client.QueryTransactionBlocks(ctx, filter, &PaginationArgs{
		First: Ptr(10),
	})
	if err != nil {
		log.Printf("QueryTransactionBlocks error: %v", err)
	} else if txsConnection != nil {
		printJSON("QueryTransactionBlocks result", txsConnection)
	}
}
