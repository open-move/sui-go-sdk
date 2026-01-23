package main

import (
	"context"
	"encoding/base64"
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
	fmt.Printf("%s:\n%s\n\n", label, string(data))
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

	senderAddr := SuiAddress(address)
	fmt.Printf("Using address: %s\n\n", address)

	client := NewClient(
		WithEndpoint(TestnetEndpoint),
		WithTimeout(30*time.Second),
	)

	ctx := context.Background()

	// Run all examples
	exampleCustomMutationBuilder(ctx, client, senderAddr)
	exampleSimulateMutation(ctx, client, senderAddr)
	exampleMoveCall(ctx, client, senderAddr)
	exampleSendSui(ctx, client, senderAddr)
	exampleSendObject(ctx, client, senderAddr)
	exampleExecuteTransaction(ctx, client, senderAddr, kp)
}

// =============================================================================
// Example 1: Custom Mutation Builder
// =============================================================================
// The MutationBuilder provides a fluent API for building and executing
// custom GraphQL mutations with fine-grained control over response fields.

func exampleCustomMutationBuilder(ctx context.Context, client *Client, sender SuiAddress) {
	fmt.Println("=== Custom Mutation Builder Example ===")

	// Use the ChainedTransactionBuilder for building complex transactions
	// This provides a fluent API for chaining multiple operations
	ctb := NewChainedTransaction(client, sender)

	// Build a transaction that splits coins (simulates a payment)
	recipient := SuiAddress("0x559ef1509af6e837d4153b3b08d9534d3df3f336a5cb6498fa248ce6cb2172e6")
	ctb.Pay(recipient, 1000000) // Pay 0.001 SUI
	ctb.WithGasBudget(10000000)

	// Simulate the transaction to see results without executing
	result, err := ctb.Simulate(ctx)
	if err != nil {
		log.Printf("ChainedTransaction simulation error: %v", err)
		return
	}
	printJSON("ChainedTransaction Simulation Result", result)

	// Estimate gas for the transaction
	gas, err := ctb.EstimateGas(ctx)
	if err != nil {
		log.Printf("Gas estimation error: %v", err)
	} else {
		printJSON("Estimated Gas Cost", gas)
	}
	fmt.Println()
}

// =============================================================================
// Example 2: Simulate Mutation (Custom Mutations)
// =============================================================================
// Use SimulateMutationBuilder for custom simulation queries with specific
// response fields.

func exampleSimulateMutation(ctx context.Context, client *Client, sender SuiAddress) {
	fmt.Println("=== Custom Simulate Mutation Example ===")

	// First, build a transaction using the TransactionBuilder
	tb := NewTransactionBuilder()
	tb.SetSender(sender)
	tb.SetGasBudget(10000000)

	// Add a simple SUI transfer
	recipient := SuiAddress("0x559ef1509af6e837d4153b3b08d9534d3df3f336a5cb6498fa248ce6cb2172e6")
	amountArg := tb.AddPureU64(2000000) // 0.002 SUI
	recipientArg := tb.AddPureAddress(recipient)
	splitIdx := tb.SplitCoins(tb.GasCoin(), []Argument{amountArg})
	tb.TransferObjects([]Argument{tb.Result(splitIdx)}, recipientArg)

	// Simulate using the client's simulation method
	result, err := client.SimulateTransactionBuilder(ctx, tb, nil)
	if err != nil {
		log.Printf("SimulateTransactionBuilder error: %v", err)
		return
	}
	printJSON("Simulation Result", result)

	// Alternative: Use the SimulateMutationBuilder for more control
	// This allows customizing which fields are returned
	fmt.Println("--- Using SimulateMutationBuilder for Custom Fields ---")

	// The SimulateMutationBuilder lets you specify exactly which effect
	// fields to include in the response
	smb := SimulateMutation().
		WithAllEffects() // Request all available effect fields

	// Build the mutation (for reference - actual execution requires txBytes)
	query, vars := smb.Build()
	fmt.Printf("Generated Query Structure:\n%s\nVariables: %v\n\n", query[:100]+"...", vars)
	fmt.Println()
}

// =============================================================================
// Example 3: Move Calls
// =============================================================================
// Execute Move function calls with type arguments and parameters.

func exampleMoveCall(ctx context.Context, client *Client, sender SuiAddress) {
	fmt.Println("=== Move Call Examples ===")

	// Example 1: Call a Move function using MoveCallParams
	fmt.Println("--- MoveCall with MoveCallParams ---")
	moveCallParams := MoveCallParams{
		Sender:        sender,
		Package:       SuiAddress("0x2"), // Sui framework
		Module:        "coin",
		Function:      "value",
		TypeArguments: []string{"0x2::sui::SUI"},
		Arguments:     []interface{}{}, // No arguments for this read-only call
		GasBudget:     10000000,
	}

	result, err := client.SimulateMoveCall(ctx, moveCallParams)
	if err != nil {
		log.Printf("SimulateMoveCall error: %v", err)
	} else {
		printJSON("MoveCall (coin::value) Simulation", result)
	}

	// Example 2: Using CallMoveFunction convenience method
	fmt.Println("--- Using CallMoveFunction ---")
	// This parses the target string automatically (package::module::function format)
	callResult, err := client.CallMoveFunction(
		ctx,
		sender,
		"0x2::coin::value",        // Target in package::module::function format
		[]string{"0x2::sui::SUI"}, // Type arguments
		[]interface{}{},           // Arguments
		10000000,                  // Gas budget
	)
	if err != nil {
		log.Printf("CallMoveFunction error: %v", err)
	} else {
		printJSON("CallMoveFunction Result", callResult)
	}

	// Example 3: Build a custom Move call transaction with TransactionBuilder
	fmt.Println("--- Custom MoveCall with TransactionBuilder ---")
	tb := NewTransactionBuilder()
	tb.SetSender(sender)
	tb.SetGasBudget(10000000)

	// Call clock::timestamp_ms from sui framework
	tb.MoveCall(
		"0x6::clock::timestamp_ms", // Clock module
		[]string{},                 // No type arguments
		[]Argument{ // Arguments
			tb.AddSharedObjectInput( // Clock is a shared object
				SuiAddress("0x6"),
				UInt53(1),
				false, // immutable access
			),
		},
	)

	simResult, err := client.SimulateTransactionBuilder(ctx, tb, nil)
	if err != nil {
		log.Printf("Custom MoveCall simulation error: %v", err)
	} else {
		printJSON("Custom MoveCall Simulation", simResult)
	}
	fmt.Println()
}

// =============================================================================
// Example 4: Send SUI
// =============================================================================
// Transfer SUI tokens to another address.

func exampleSendSui(ctx context.Context, client *Client, sender SuiAddress) {
	fmt.Println("=== Send SUI Examples ===")

	recipient := SuiAddress("0x559ef1509af6e837d4153b3b08d9534d3df3f336a5cb6498fa248ce6cb2172e6")

	// Example 1: Using TransferSuiParams
	fmt.Println("--- TransferSui with Params ---")
	params := TransferSuiParams{
		Sender:    sender,
		Recipient: recipient,
		Amount:    5000000, // 0.005 SUI
		GasBudget: 10000000,
	}

	// Simulate the transfer first
	simResult, err := client.SimulateTransferSui(ctx, params)
	if err != nil {
		log.Printf("SimulateTransferSui error: %v", err)
	} else {
		printJSON("TransferSui Simulation", simResult)
	}

	// Get comprehensive transaction result (simulation mode)
	fmt.Println("--- TransferSui Full Result ---")
	txResult, err := client.TransferSui(ctx, params)
	if err != nil {
		log.Printf("TransferSui error: %v", err)
	} else {
		printJSON("TransferSui Full Result", txResult)
	}

	// Example 2: Using convenience method
	fmt.Println("--- TransferSuiFromGas ---")
	result, err := client.TransferSuiFromGas(
		ctx,
		sender,
		recipient,
		3000000, // 0.003 SUI
		10000000,
	)
	if err != nil {
		log.Printf("TransferSuiFromGas error: %v", err)
	} else {
		printJSON("TransferSuiFromGas Result", result)
	}

	// Example 3: Estimate gas for transfer
	fmt.Println("--- Estimate Gas for Transfer ---")
	gasCost, err := client.EstimateGasForTransferSui(ctx, params)
	if err != nil {
		log.Printf("EstimateGasForTransferSui error: %v", err)
	} else {
		printJSON("Estimated Gas for TransferSui", gasCost)
	}

	// Example 4: Using ChainedTransaction for multiple payments
	fmt.Println("--- Multiple Payments with ChainedTransaction ---")
	recipient2 := SuiAddress("0x0000000000000000000000000000000000000000000000000000000000000001")

	ctb := NewChainedTransaction(client, sender)
	ctb.PayMultiple([]Payment{
		{Recipient: recipient, Amount: 1000000},  // 0.001 SUI
		{Recipient: recipient2, Amount: 1000000}, // 0.001 SUI
	})
	ctb.WithGasBudget(10000000)

	multiPayResult, err := ctb.Simulate(ctx)
	if err != nil {
		log.Printf("Multiple payments simulation error: %v", err)
	} else {
		printJSON("Multiple Payments Simulation", multiPayResult)
	}
	fmt.Println()
}

// =============================================================================
// Example 5: Send Objects
// =============================================================================
// Transfer objects to another address.

func exampleSendObject(ctx context.Context, client *Client, sender SuiAddress) {
	fmt.Println("=== Send Object Examples ===")

	recipient := SuiAddress("0x559ef1509af6e837d4153b3b08d9534d3df3f336a5cb6498fa248ce6cb2172e6")

	// First, get an object to transfer (using coins as example objects)
	coins, err := client.GetCoins(ctx, sender, nil, nil)
	if err != nil {
		log.Printf("GetCoins error: %v", err)
		return
	}

	if len(coins.Nodes) < 2 {
		log.Printf("Need at least 2 coins for transfer example, found: %d", len(coins.Nodes))
		// Create a placeholder for demonstration
		fmt.Println("--- Demonstrating Object Transfer Structure ---")

		// Build a transfer transaction structure for documentation purposes
		tb := NewTransactionBuilder()
		tb.SetSender(sender)
		tb.SetGasBudget(10000000)

		// Example of how to add an object and transfer it
		exampleObjectRef := ObjectRef{
			ObjectID: SuiAddress("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"),
			Version:  UInt53(100),
			Digest:   "ExampleDigest123",
		}
		objectArg := tb.AddObjectInput(exampleObjectRef)
		recipientArg := tb.AddPureAddress(recipient)
		tb.TransferObjects([]Argument{objectArg}, recipientArg)

		txData, err := tb.Build()
		if err != nil {
			log.Printf("Build error: %v", err)
		} else {
			printJSON("Transfer Object Transaction Structure", txData)
		}
		return
	}

	// Use the second coin (first is typically used for gas)
	coin := coins.Nodes[1]
	objectRef := ObjectRef{
		ObjectID: coin.Address,
		Version:  coin.Version,
		Digest:   coin.Digest,
	}

	fmt.Printf("Transferring object: %s (version: %d)\n", coin.Address, coin.Version)

	// Example 1: Using TransferObjectParams
	fmt.Println("--- TransferObject with Params ---")
	params := TransferObjectParams{
		Sender:    sender,
		Recipient: recipient,
		Object:    objectRef,
		GasBudget: 10000000,
	}

	// Simulate the object transfer
	simResult, err := client.SimulateTransferObject(ctx, params)
	if err != nil {
		log.Printf("SimulateTransferObject error: %v", err)
	} else {
		printJSON("TransferObject Simulation", simResult)
	}

	// Get comprehensive transaction result (simulation mode)
	fmt.Println("--- TransferObject Full Result ---")
	txResult, err := client.TransferObject(ctx, params)
	if err != nil {
		log.Printf("TransferObject error: %v", err)
	} else {
		printJSON("TransferObject Full Result", txResult)
	}

	// Example 2: Using TransactionBuilder directly
	fmt.Println("--- Custom Object Transfer with TransactionBuilder ---")
	tb := NewTransactionBuilder()
	tb.SetSender(sender)
	tb.SetGasBudget(10000000)

	// Add the object and recipient
	objectArg := tb.AddObjectInput(objectRef)
	recipientArg := tb.AddPureAddress(recipient)

	// Transfer the object
	tb.TransferObjects([]Argument{objectArg}, recipientArg)

	customResult, err := client.SimulateTransactionBuilder(ctx, tb, nil)
	if err != nil {
		log.Printf("Custom TransferObject simulation error: %v", err)
	} else {
		printJSON("Custom TransferObject Simulation", customResult)
	}

	// Example 3: Transfer multiple objects in one transaction
	if len(coins.Nodes) >= 3 {
		fmt.Println("--- Transfer Multiple Objects ---")
		tb2 := NewTransactionBuilder()
		tb2.SetSender(sender)
		tb2.SetGasBudget(15000000)

		recipientArg2 := tb2.AddPureAddress(recipient)

		// Add multiple objects
		var objectArgs []Argument
		for i := 1; i < 3 && i < len(coins.Nodes); i++ {
			c := coins.Nodes[i]
			objRef := ObjectRef{
				ObjectID: c.Address,
				Version:  c.Version,
				Digest:   c.Digest,
			}
			objectArgs = append(objectArgs, tb2.AddObjectInput(objRef))
		}

		// Transfer all objects at once
		tb2.TransferObjects(objectArgs, recipientArg2)

		multiResult, err := client.SimulateTransactionBuilder(ctx, tb2, nil)
		if err != nil {
			log.Printf("Multiple objects transfer simulation error: %v", err)
		} else {
			printJSON("Multiple Objects Transfer Simulation", multiResult)
		}
	}

	// Example 4: Using ChainedTransaction for object transfer
	fmt.Println("--- Object Transfer with ChainedTransaction ---")
	ctb := NewChainedTransaction(client, sender)
	ctb.Transfer(
		[]Argument{ctb.AddObjectInput(objectRef)},
		recipient,
	)
	ctb.WithGasBudget(10000000)

	chainedResult, err := ctb.Simulate(ctx)
	if err != nil {
		log.Printf("ChainedTransaction object transfer error: %v", err)
	} else {
		printJSON("ChainedTransaction Object Transfer", chainedResult)
	}
	fmt.Println()
}

// =============================================================================
// Example 6: Execute Transaction (Sign and Submit)
// =============================================================================
// Demonstrates the complete flow for executing a transaction on-chain.
// This requires BCS-serialized transaction bytes which can be obtained from
// external sources or by using the Sui TypeScript SDK for serialization.

func exampleExecuteTransaction(ctx context.Context, client *Client, sender SuiAddress, kp keypair.Keypair) {
	fmt.Println("=== Execute Transaction Example ===")
	fmt.Println("This example demonstrates how to sign and execute transactions.")
	fmt.Println()

	// The complete flow for executing a transaction is:
	// 1. Build the transaction (using TransactionBuilder or pre-built params)
	// 2. Simulate to verify and estimate gas
	// 3. Serialize the transaction to BCS format
	// 4. Sign the BCS bytes with the keypair
	// 5. Submit the signed transaction

	// Step 1 & 2: Build and simulate the transaction
	// Note: Some simulation methods may return errors due to transaction JSON format
	// The MoveCall methods work correctly for simulation
	fmt.Println("Step 1 & 2: Build and simulate transaction (using MoveCall)")
	moveCallResult, err := client.CallMoveFunction(
		ctx,
		sender,
		"0x2::coin::value",        // Target in package::module::function format
		[]string{"0x2::sui::SUI"}, // Type arguments
		[]interface{}{},           // Arguments
		10000000,                  // Gas budget
	)
	if err != nil {
		log.Printf("MoveCall simulation failed: %v", err)
	} else if moveCallResult.Effects != nil && moveCallResult.Effects.Status != nil {
		fmt.Printf("MoveCall simulation successful! Status: %s\n", moveCallResult.Effects.Status.Status)
		if moveCallResult.Effects.GasUsed != nil {
			fmt.Printf("Estimated gas: computation=%s, storage=%s\n",
				moveCallResult.Effects.GasUsed.ComputationCost,
				moveCallResult.Effects.GasUsed.StorageCost)
		}
	}
	fmt.Println()

	// Step 3: About BCS serialization
	// The Sui GraphQL API uses JSON for transaction simulation but requires
	// BCS-serialized bytes for execution. There are several ways to get BCS bytes:
	//
	// Option A: Use the Sui TypeScript SDK to build and serialize transactions
	// Option B: Use the Sui CLI to build transactions
	// Option C: Use a separate BCS serialization library (not yet implemented in this SDK)
	//
	// For now, we demonstrate with a mock BCS transaction.
	fmt.Println("Step 3: BCS Serialization")
	fmt.Println("Note: Full BCS serialization is typically done via the Sui TypeScript SDK")
	fmt.Println("or Sui CLI. This SDK provides signing capabilities once you have BCS bytes.")
	fmt.Println()

	// Step 4: Demonstrate signing (with mock transaction bytes)
	fmt.Println("Step 4: Sign the transaction")
	// This is a mock transaction for demonstration - in practice, you would use
	// actual BCS-serialized transaction bytes
	mockTxBytes := []byte("mock_transaction_bytes_for_demonstration")

	signature, err := kp.SignTransaction(mockTxBytes)
	if err != nil {
		log.Printf("Signing failed: %v", err)
		return
	}

	signatureBase64 := base64.StdEncoding.EncodeToString(signature)
	fmt.Printf("Transaction signed successfully!\n")
	fmt.Printf("Signature (base64, first 50 chars): %s...\n", signatureBase64[:min(50, len(signatureBase64))])
	fmt.Println()

	// Step 5: Execute the transaction (commented out as we're using mock data)
	fmt.Println("Step 5: Execute transaction")
	fmt.Println("With real BCS bytes, you would call:")
	fmt.Println("  result, err := client.ExecuteSignedTransaction(ctx, txBytes, []string{signature})")
	fmt.Println()

	// Example of how execution would work with real transaction bytes:
	/*
		// Assuming you have real BCS-serialized transaction bytes:
		realTxBytes := Base64("base64_encoded_bcs_transaction")

		// Sign the transaction
		txBytesRaw, _ := base64.StdEncoding.DecodeString(string(realTxBytes))
		sig, err := kp.SignTransaction(txBytesRaw)
		if err != nil {
			log.Fatalf("Failed to sign: %v", err)
		}
		sigBase64 := base64.StdEncoding.EncodeToString(sig)

		// Execute the signed transaction
		result, err := client.ExecuteSignedTransaction(ctx, realTxBytes, []string{sigBase64})
		if err != nil {
			log.Fatalf("Execution failed: %v", err)
		}

		fmt.Printf("Transaction executed! Digest: %s\n", result.Digest)
		if result.Effects != nil && result.Effects.Status != nil {
			fmt.Printf("Status: %s\n", result.Effects.Status.Status)
		}
	*/

	fmt.Println("=== End of Execute Transaction Example ===")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
