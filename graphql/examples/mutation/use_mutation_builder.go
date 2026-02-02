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

// UseMutationBuilder demonstrates how to use the Mutation Builder
func UseMutationBuilder(ctx context.Context, client *graphql.Client, kp keypair.Keypair, sender, recipient string, gasPrice uint64, gasCoinRef *types.ObjectRef) *types.ObjectRef {
	fmt.Println()
	fmt.Println("=== Example 4: Use Mutation Builder ===")

	amount := uint64(500000) // 0.0005 SUI
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

	// Build the transaction
	tb := transaction.New()
	tb.SetSender(sender)
	tb.SetGasBudget(gasBudget)
	tb.SetGasPrice(gasPrice)
	tb.SetGasPayment([]types.ObjectRef{gasPayment})

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

	// Display the mutation builder pattern
	fmt.Println("Demonstrating ExecuteMutationBuilder for custom execution...")

	// Build BCS bytes
	buildResult, err := tb.Build(ctx, transaction.BuildOptions{})
	if err != nil {
		log.Printf("Failed to build transaction: %v", err)
		return nil
	}

	txBcsBase64 := base64.StdEncoding.EncodeToString(buildResult.TransactionBytes)
	fmt.Printf("Transaction BCS size: %d bytes\n", len(buildResult.TransactionBytes))

	// Sign the transaction
	signature, err := kp.SignTransaction(buildResult.TransactionBytes)
	if err != nil {
		log.Printf("Failed to sign transaction: %v", err)
		return nil
	}
	signatureBase64 := base64.StdEncoding.EncodeToString(signature)

	// Demonstrate ExecuteMutationBuilder setup
	emb := graphql.ExecuteMutation().
		TxBytes(txBcsBase64).
		Signatures(signatureBase64).
		WithAllEffects()

	query, vars := emb.Build()
	fmt.Printf("ExecuteMutationBuilder configured:\n")
	fmt.Printf("  Query length: %d characters\n", len(query))
	fmt.Printf("  Variables: tx=%d chars, signatures=1\n", len(vars["tx"].(string)))
	fmt.Printf("  Amount: %d MIST (%.6f SUI)\n", amount, float64(amount)/1e9)

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
