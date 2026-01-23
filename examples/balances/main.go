// Package main demonstrates balance and coin queries using the Sui GraphQL SDK.
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

	// Create a client connected to testnet
	client := NewClient(
		WithEndpoint(TestnetEndpoint),
		WithTimeout(30*time.Second),
	)

	ctx := context.Background()

	// Example: Get all balances for an address
	fmt.Println("=== GetAllBalances ===")
	balances, err := client.GetAllBalances(ctx, SuiAddress(address))
	if err != nil {
		log.Printf("GetAllBalances error: %v", err)
	} else {
		printJSON("GetAllBalances result", balances)
		fmt.Printf("Found %d balance(s):\n", len(balances))
		for _, b := range balances {
			if b.CoinType != nil {
				fmt.Printf("  - %s: %s\n", b.CoinType.Repr, b.TotalBalance)
			}
		}
	}
	fmt.Println()

	// Example: Get coin metadata for SUI
	fmt.Println("=== GetCoinMetadata ===")
	coinType := "0x2::sui::SUI"
	metadata, err := client.GetCoinMetadata(ctx, coinType)
	if err != nil {
		log.Printf("GetCoinMetadata error: %v", err)
	} else if metadata != nil {
		printJSON("GetCoinMetadata result", metadata)
		fmt.Printf("Coin: %s\n", *metadata.Name)
		fmt.Printf("Symbol: %s\n", *metadata.Symbol)
		fmt.Printf("Decimals: %d\n", *metadata.Decimals)
		fmt.Printf("Description: %s\n", *metadata.Description)
	}
	fmt.Println()

	// Example: Get total supply of SUI
	fmt.Println("=== GetTotalSupply ===")
	supply, err := client.GetTotalSupply(ctx, coinType)
	if err != nil {
		log.Printf("GetTotalSupply error: %v", err)
	} else if supply != nil {
		printJSON("GetTotalSupply result", supply)
		fmt.Printf("Total SUI Supply: %s\n", *supply)
	}
	fmt.Println()
}
