package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/open-move/sui-go-sdk/examples/mutation"
	"github.com/open-move/sui-go-sdk/examples/query"
	"github.com/open-move/sui-go-sdk/graphql"
	"github.com/open-move/sui-go-sdk/keypair"
	"github.com/open-move/sui-go-sdk/types"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found")
	}

	// Initialize the GraphQL client (using Testnet)
	client := graphql.NewClient(graphql.WithEndpoint(graphql.TestnetEndpoint))
	ctx := context.Background()

	// Load keypair from environment variable
	privKey := os.Getenv("SUI_PRIVATE_KEY")
	if privKey == "" {
		log.Fatal("SUI_PRIVATE_KEY not set in .env")
	}

	kp, err := keypair.FromBech32(privKey)
	if err != nil {
		log.Fatalf("Failed to parse private key: %v", err)
	}

	sender, err := kp.SuiAddress()
	if err != nil {
		log.Fatalf("Failed to get sender address: %v", err)
	}
	fmt.Printf("Using sender address: %s\n", sender)

	// Use a dummy recipient address
	recipient := "0x0000000000000000000000000000000000000000000000000000000000000000"

	// --- Query Examples ---
	fmt.Println("\n========================================")
	fmt.Println("Running Query Examples")
	fmt.Println("========================================")

	// Use an address that likely has some activity on testnet for better query results,
	// or use the sender (which is empty)
	// Using a known address from examples
	exampleAddress := "0xf41564ce5236f344bc79abb0c6ca22bb31edc4ec64b995824e986b81e71eb031"

	query.Balances(ctx, client, exampleAddress)
	query.Events(ctx, client)
	query.Objects(ctx, client)
	query.Packages(ctx, client)
	query.Protocol(ctx, client)
	query.QueryBuilder(ctx, client)
	query.Transactions(ctx, client, exampleAddress)

	// Custom Queries
	exampleDigest := "3KmTo5yvbkeg9mrQafUkfFoYVRt45zTg2SoGWBJw615V"
	query.CustomQueries(ctx, client, exampleDigest, exampleAddress)

	// --- Mutation Examples ---
	fmt.Println("\n========================================")
	fmt.Println("Running Mutation Examples")
	fmt.Println("========================================")
	fmt.Println("Note: These will likely fail or return nil because the generated account has no SUI for gas.")

	// Get gas price first
	gasPrice, err := mutation.GetGasPrice(ctx, client)
	if err != nil {
		log.Printf("Failed to get gas price: %v", err)
		gasPrice = 1000 // Fallback
	}

	// We can try to chain them, but they return *types.ObjectRef which might be nil if failed.
	// Also, if the first one fails (due to no gas), others will too.
	// But we will call them all as requested.

	var gasCoinRef *types.ObjectRef // Initially nil, will try to fetch from network

	// 1. Send SUI
	fmt.Println("\n--- SendSui ---")
	gasCoinRef = mutation.SendSui(ctx, client, kp, sender, recipient, gasPrice, gasCoinRef)

	// 2. Make Move Call
	fmt.Println("\n--- MakeMoveCall ---")
	gasCoinRef = mutation.MakeMoveCall(ctx, client, kp, sender, gasPrice, gasCoinRef)

	// 3. Send Objects
	fmt.Println("\n--- SendObjects ---")
	gasCoinRef = mutation.SendObjects(ctx, client, kp, sender, recipient, gasPrice, gasCoinRef)

	// 4. Use Mutation Builder
	fmt.Println("\n--- UseMutationBuilder ---")
	gasCoinRef = mutation.UseMutationBuilder(ctx, client, kp, sender, recipient, gasPrice, gasCoinRef)

	// 5. Use Custom Mutation Sender
	fmt.Println("\n--- UseCustomMutationSender ---")
	gasCoinRef = mutation.UseCustomMutationSender(ctx, client, kp, sender, recipient, gasPrice, gasCoinRef)

	fmt.Println("\nAll examples executed.")
}
