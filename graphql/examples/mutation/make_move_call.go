package mutation

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"

	"github.com/open-move/sui-go-sdk/graphql"
	"github.com/open-move/sui-go-sdk/keypair"
	"github.com/open-move/sui-go-sdk/transaction"
	"github.com/open-move/sui-go-sdk/types"
	"github.com/open-move/sui-go-sdk/utils"
)

// MakeMoveCall demonstrates how to make a Move call
func MakeMoveCall(ctx context.Context, client *graphql.Client, kp keypair.Keypair, sender string, gasPrice uint64, gasCoinRef *types.ObjectRef) *types.ObjectRef {
	fmt.Println()
	fmt.Println("=== Example 2: Make a Move Call ===")

	gasBudget := uint64(10000000)

	senderAddr, err := utils.ParseAddress(sender)
	if err != nil {
		log.Printf("invalid sender address: %v", err)
		return nil
	}

	var gasPayment types.ObjectRef
	var gasCoinAddress string

	if gasCoinRef != nil {
		gasPayment = *gasCoinRef
		gasCoinAddress = gasPayment.ObjectID.String()
		fmt.Printf("Using provided gas coin: %s (version: %d)\n", gasCoinAddress, gasPayment.Version)
	} else {
		coins, err := client.GetCoins(ctx, senderAddr, nil, nil)
		if err != nil {
			log.Printf("Failed to get coins: %v", err)
			return nil
		}

		if len(coins.Nodes) == 0 {
			log.Printf("No coins available")
			return nil
		}

		gasCoin := coins.Nodes[0]
		gasPayment = types.ObjectRef{
			ObjectID: gasCoin.Address,
			Version:  uint64(gasCoin.Version),
			Digest:   gasCoin.Digest,
		}
		gasCoinAddress = gasCoin.Address.String()
		fmt.Printf("Using gas coin: %s (version: %d)\n", gasCoinAddress, gasCoin.Version)
	}

	// Build a Move call to get the current timestamp from the Clock object
	tb := transaction.New()
	tb.SetSender(sender)
	tb.SetGasBudget(gasBudget)
	tb.SetGasPrice(gasPrice)
	tb.SetGasPayment([]types.ObjectRef{gasPayment})

	// Add the Clock shared object as input
	// Clock object ID is 0x6, initial version is 1, immutable access
	clockAddr, _ := utils.ParseAddress("0x0000000000000000000000000000000000000000000000000000000000000006")
	clockRef := types.SharedObjectRef{
		ObjectID:             clockAddr,
		InitialSharedVersion: 1,
		Mutable:              false,
	}
	clockArg := tb.SharedObject(clockRef)

	// Call clock::timestamp_ms
	tb.MoveCall(transaction.MoveCall{
		Package:   "0x0000000000000000000000000000000000000000000000000000000000000002",
		Module:    "clock",
		Function:  "timestamp_ms",
		Arguments: []transaction.Argument{clockArg},
	})

	if err := tb.Err(); err != nil {
		log.Printf("Transaction build error: %v", err)
		return nil
	}

	// Build BCS bytes
	buildResult, err := tb.Build(ctx, transaction.BuildOptions{})
	if err != nil {
		log.Printf("Failed to build transaction: %v", err)
		return nil
	}

	fmt.Printf("Move Call transaction built successfully!\n")
	fmt.Printf("  BCS size: %d bytes\n", len(buildResult.TransactionBytes))
	fmt.Printf("  Target: 0x2::clock::timestamp_ms\n")
	fmt.Printf("  Clock object: 0x6 (shared)\n")

	// Sign the transaction
	signature, err := kp.SignTransaction(buildResult.TransactionBytes)
	if err != nil {
		log.Printf("Failed to sign transaction: %v", err)
		return nil
	}
	fmt.Printf("  Signature: %s...\n", base64.StdEncoding.EncodeToString(signature)[:20])

	// Execute the transaction
	fmt.Println("Executing transaction...")

	result, err := graphql.ExecuteTransaction(client, ctx, buildResult.TransactionBytes, [][]byte{signature})
	if err != nil {
		log.Printf("Failed to execute transaction: %v", err)
		return nil
	}

	if result.Effects != nil {
		fmt.Printf("✅ Transaction executed successfully!\n")
		fmt.Printf("  Transaction digest: %s\n", result.Effects.Digest.String())
		fmt.Printf("  Status: %s\n", result.Effects.Status)
		if result.Effects.ExecutionError != nil {
			fmt.Printf("  Execution error: %s\n", result.Effects.ExecutionError.Message)
		}
		return extractUpdatedGasCoin(result, gasCoinAddress)
	} else if len(result.Errors) > 0 {
		fmt.Printf("❌ Transaction failed with errors: %v\n", result.Errors)
	}
	return nil
}
