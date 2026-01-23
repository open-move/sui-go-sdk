package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/open-move/sui-go-sdk/graphql"
	"github.com/open-move/sui-go-sdk/keypair"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Get private key from environment
	privateKey := os.Getenv("SUI_PRIVATE_KEY")
	if privateKey == "" {
		log.Fatal("SUI_PRIVATE_KEY environment variable is required (bech32 format: suiprivkey1...)")
	}

	// Parse the keypair from bech32 format
	kp, err := keypair.FromBech32(privateKey)
	if err != nil {
		log.Fatalf("Failed to parse private key: %v", err)
	}

	// Get the sender address
	senderAddress, err := kp.SuiAddress()
	if err != nil {
		log.Fatalf("Failed to get sender address: %v", err)
	}

	fmt.Printf("=== Sui Transaction Execution ===\n")
	fmt.Printf("Sender address: %s\n\n", senderAddress)

	// Get transaction BCS bytes from environment or command line
	// Generate using: sui client transfer-sui --to <addr> --amount <mist> --serialize-unsigned-transaction
	txBcsBase64 := os.Getenv("TX_BCS")
	if txBcsBase64 == "" && len(os.Args) > 1 {
		txBcsBase64 = os.Args[1]
	}

	if txBcsBase64 == "" {
		fmt.Println("Usage:")
		fmt.Println("  TX_BCS=<base64_tx_bytes> go run main.go")
		fmt.Println("  go run main.go <base64_tx_bytes>")
		fmt.Println()
		fmt.Println("To generate transaction bytes:")
		fmt.Println("  sui client transfer-sui --to 0x... --sui-coin-object-id 0x... --amount 1000000 --gas-budget 10000000 --serialize-unsigned-transaction")
		fmt.Println()
		fmt.Println("Or for a Move call:")
		fmt.Println("  sui client call --package 0x... --module ... --function ... --serialize-unsigned-transaction")
		os.Exit(1)
	}

	// Create client for Sui testnet
	client := NewClient(
		WithEndpoint(TestnetEndpoint),
		WithTimeout(30*time.Second),
	)

	ctx := context.Background()

	fmt.Println("=== Transaction Details ===")
	fmt.Printf("BCS bytes (first 60 chars): %s...\n", txBcsBase64[:min(60, len(txBcsBase64))])

	// Step 1: Decode the transaction bytes
	txBytes, err := base64.StdEncoding.DecodeString(txBcsBase64)
	if err != nil {
		log.Fatalf("Failed to decode transaction bytes: %v", err)
	}
	fmt.Printf("Transaction size: %d bytes\n\n", len(txBytes))

	// Step 2: Sign the transaction using the SDK's txsigning package
	fmt.Println("=== Signing Transaction ===")
	signature, err := kp.SignTransaction(txBytes)
	if err != nil {
		log.Fatalf("Failed to sign transaction: %v", err)
	}

	signatureBase64 := base64.StdEncoding.EncodeToString(signature)
	fmt.Printf("Signature (first 60 chars): %s...\n", signatureBase64[:min(60, len(signatureBase64))])
	fmt.Printf("Signature size: %d bytes\n\n", len(signature))

	// Step 3: Execute the signed transaction
	fmt.Println("=== Executing Transaction on Testnet ===")
	result, err := client.ExecuteTransaction(
		ctx,
		Base64(txBcsBase64),
		[]Base64{Base64(signatureBase64)},
	)
	if err != nil {
		log.Fatalf("Transaction execution failed: %v", err)
	}

	// Print full result as JSON
	data, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println("Raw Result:")
	fmt.Println(string(data))

	// Check for errors
	if result != nil && len(result.Errors) > 0 {
		fmt.Println("\n=== Transaction Failed ===")
		for _, e := range result.Errors {
			fmt.Printf("Error: %s\n", e)
			if strings.Contains(e, "not available for consumption") {
				fmt.Println("\nThis error means the object version in the transaction is stale.")
				fmt.Println("Generate fresh transaction bytes and execute immediately.")
			}
		}
		os.Exit(1)
	}

	// Print success summary
	if result != nil && result.Effects != nil {
		fmt.Printf("\n=== Transaction Successful ===\n")
		fmt.Printf("Digest: %s\n", result.Effects.Digest)
		fmt.Printf("Status: %s\n", result.Effects.Status)

		if result.Effects.Epoch != nil {
			fmt.Printf("Epoch: %d\n", result.Effects.Epoch.EpochID)
		}

		if result.Effects.Checkpoint != nil {
			fmt.Printf("Checkpoint: %d\n", result.Effects.Checkpoint.SequenceNumber)
		}

		if result.Effects.GasEffects != nil && result.Effects.GasEffects.GasSummary != nil {
			gas := result.Effects.GasEffects.GasSummary
			totalGas := gas.ComputationCost + gas.StorageCost - gas.StorageRebate
			fmt.Printf("\nGas Used:\n")
			fmt.Printf("  Computation: %d MIST\n", gas.ComputationCost)
			fmt.Printf("  Storage: %d MIST\n", gas.StorageCost)
			fmt.Printf("  Rebate: %d MIST\n", gas.StorageRebate)
			fmt.Printf("  Total: %d MIST (%.6f SUI)\n", totalGas, float64(totalGas)/1e9)
		}

		fmt.Printf("\nView on explorer: https://testnet.suivision.xyz/txblock/%s\n", result.Effects.Digest)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
