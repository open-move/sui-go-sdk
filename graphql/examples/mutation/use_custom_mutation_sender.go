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

// CustomMutationResult defines the response shape for your custom mutation.
// Structure this to match exactly what your GraphQL mutation returns.
type CustomMutationResult struct {
	ExecuteTransaction struct {
		Errors  []string `json:"errors"`
		Effects *struct {
			Status string `json:"status"`
			Epoch  *struct {
				StartTimestamp string `json:"startTimestamp"`
			} `json:"epoch"`
			GasEffects *struct {
				GasSummary *struct {
					ComputationCost uint64 `json:"computationCost"`
				} `json:"gasSummary"`
			} `json:"gasEffects"`
		} `json:"effects"`
	} `json:"executeTransaction"`
}

// UseCustomMutationSender demonstrates how to execute your own custom GraphQL mutation.
// This allows full control over the mutation query and response fields.
func UseCustomMutationSender(ctx context.Context, client *graphql.Client, kp keypair.Keypair, sender, recipient string, gasPrice uint64, gasCoinRef *types.ObjectRef) *types.ObjectRef {
	fmt.Println()
	fmt.Println("=== Example 5: Execute Custom GraphQL Mutation ===")

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

	buildResult, err := tb.Build(ctx, transaction.BuildOptions{})
	if err != nil {
		log.Printf("Failed to build transaction: %v", err)
		return nil
	}

	// Sign the transaction
	signature, err := kp.SignTransaction(buildResult.TransactionBytes)
	if err != nil {
		log.Printf("Failed to sign transaction: %v", err)
		return nil
	}

	// ==========================================================================
	// Execute Your Own Custom GraphQL Mutation
	// ==========================================================================
	// Define your own mutation query with exactly the fields you need.
	// This gives you full control over:
	// - Which fields to request (reduces response size)
	// - Custom field selections not in standard responses
	// - The exact GraphQL query structure

	customMutation := `
		mutation ($tx: String!, $sigs: [String!]!) {
			executeTransaction(transactionDataBcs: $tx, signatures: $sigs) {
				errors
				effects {
					status
					epoch {
						startTimestamp
					}
					gasEffects {
						gasSummary {
							computationCost
						}
					}
				}
			}
		}
	`

	// Prepare variables - encode transaction bytes and signature as base64
	variables := map[string]any{
		"tx":   base64.StdEncoding.EncodeToString(buildResult.TransactionBytes),
		"sigs": []string{base64.StdEncoding.EncodeToString(signature)},
	}

	fmt.Println("Executing custom mutation...")
	fmt.Printf("  Mutation length: %d characters\n", len(customMutation))

	// Execute directly using client.Execute with your custom result struct
	var result CustomMutationResult
	err = client.Execute(ctx, customMutation, variables, &result)
	if err != nil {
		log.Printf("Failed to execute custom mutation: %v", err)
		return nil
	}

	// Handle response using your custom struct fields
	if len(result.ExecuteTransaction.Errors) > 0 {
		fmt.Printf("❌ Custom mutation failed with errors: %v\n", result.ExecuteTransaction.Errors)
		return nil
	}

	if result.ExecuteTransaction.Effects != nil {
		effects := result.ExecuteTransaction.Effects
		fmt.Printf("✅ Custom mutation executed successfully!\n")
		fmt.Printf("  Status: %s\n", effects.Status)
		if effects.Epoch != nil {
			fmt.Printf("  Epoch start timestamp: %s\n", effects.Epoch.StartTimestamp)
		}
		if effects.GasEffects != nil && effects.GasEffects.GasSummary != nil {
			fmt.Printf("  Computation cost: %d\n", effects.GasEffects.GasSummary.ComputationCost)
		}
	}

	// Note: This example doesn't track the updated gas coin since we're using
	// a minimal custom response. Add objectChanges to your mutation if needed.
	_ = gasCoinAddress
	return nil
}
