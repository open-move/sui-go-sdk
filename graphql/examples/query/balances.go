package query

import (
	"context"
	"fmt"
	"log"

	"github.com/open-move/sui-go-sdk/graphql"
	"github.com/open-move/sui-go-sdk/utils"
)

// Balances demonstrates how to fetch balances for an address.
func Balances(ctx context.Context, client *graphql.Client, address string) {
	// Example: Get all balances for an address
	fmt.Println("=== GetAllBalances ===")
	addr, err := utils.ParseAddress(address)
	if err != nil {
		log.Printf("invalid address: %v", err)
		return
	}

	balances, err := client.GetAllBalances(ctx, addr)
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
