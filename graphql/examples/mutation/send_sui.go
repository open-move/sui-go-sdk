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

// SendSui demonstrates how to send SUI coins
func SendSui(ctx context.Context, client *graphql.Client, kp keypair.Keypair, sender, recipient string, gasPrice uint64, gasCoinRef *types.ObjectRef) *types.ObjectRef {
	fmt.Println()
	fmt.Println("=== Example 1: Send SUI ===")

	amount := uint64(1000000) // 0.001 SUI (1 SUI = 1,000,000,000 MIST)
	gasBudget := uint64(10000000)

	var gasPayment types.ObjectRef
	var gasCoinAddress string

	if gasCoinRef != nil {
		// Use provided gas coin
		gasPayment = *gasCoinRef
		gasCoinAddress = gasPayment.ObjectID.String()
		fmt.Printf("Using provided gas coin: %s (version: %d)\n", gasCoinAddress, gasPayment.Version)
	} else {
		// Fetch coins from network
		coins, err := client.GetCoins(ctx, graphql.SuiAddress(sender), nil, nil)
		if err != nil {
			log.Printf("Failed to get coins: %v", err)
			return nil
		}

		if len(coins.Nodes) == 0 {
			log.Printf("No coins available")
			return nil
		}

		gasCoin := coins.Nodes[0]
		gasPayment, err = utils.ParseObjectRef(string(gasCoin.Address), uint64(gasCoin.Version), gasCoin.Digest)
		if err != nil {
			log.Printf("Failed to parse gas coin ref: %v", err)
			return nil
		}
		gasCoinAddress = string(gasCoin.Address)
		fmt.Printf("Using gas coin: %s (version: %d)\n", gasCoin.Address, gasCoin.Version)
	}

	// Build the transaction using transaction.Builder (BCS serialization)
	tb := transaction.New()
	tb.SetSender(sender)
	tb.SetGasBudget(gasBudget)
	tb.SetGasPrice(gasPrice)
	tb.SetGasPayment([]types.ObjectRef{gasPayment})

	// Split coins from gas coin and transfer
	amountArg := tb.PureU64(amount)
	recipientArg := tb.PureAddress(recipient)
	splitResult := tb.SplitCoins(transaction.SplitCoins{
		Coin:    tb.Gas(),
		Amounts: []transaction.Argument{amountArg},
	})
	tb.TransferObjects(transaction.TransferObjects{
		Objects: splitResult,
		Address: recipientArg,
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

	fmt.Printf("Transaction built successfully!\n")
	fmt.Printf("  BCS size: %d bytes\n", len(buildResult.TransactionBytes))
	fmt.Printf("  Amount: %d MIST (%.6f SUI)\n", amount, float64(amount)/1e9)
	fmt.Printf("  Recipient: %s\n", recipient)

	// Sign the transaction
	signature, err := kp.SignTransaction(buildResult.TransactionBytes)
	if err != nil {
		log.Printf("Failed to sign transaction: %v", err)
		return nil
	}
	fmt.Printf("  Signature: %s...\n", base64.StdEncoding.EncodeToString(signature)[:20])

	// Execute the transaction
	fmt.Println("Executing transaction...")
	txBcsBase64 := base64.StdEncoding.EncodeToString(buildResult.TransactionBytes)
	signatureBase64 := base64.StdEncoding.EncodeToString(signature)

	result, err := graphql.ExecuteTransaction(client, ctx, graphql.Base64(txBcsBase64), []graphql.Base64{graphql.Base64(signatureBase64)})
	if err != nil {
		log.Printf("Failed to execute transaction: %v", err)
		return nil
	}

	if result.Effects != nil {
		fmt.Printf("✅ Transaction executed successfully!\n")
		fmt.Printf("  Transaction digest: %s\n", result.Effects.Digest)
		fmt.Printf("  Status: %s\n", result.Effects.Status)
		if result.Effects.ExecutionError != nil {
			fmt.Printf("  Execution error: %s\n", result.Effects.ExecutionError.Message)
		}
		// Return the updated gas coin
		return extractUpdatedGasCoin(result, gasCoinAddress)
	} else if len(result.Errors) > 0 {
		fmt.Printf("❌ Transaction failed with errors: %v\n", result.Errors)
	}
	return nil
}
