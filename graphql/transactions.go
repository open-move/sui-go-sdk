package graphql

import (
	"context"

	"fmt"
)

// =============================================================================
// High-Level Transaction Operations
// =============================================================================

// TransferSui transfers SUI from the sender to a recipient.
// Returns comprehensive transaction details including effects, balance changes, and object changes.
//
// Note: This function builds and simulates the transaction. To execute, you need to:
// 1. Sign the transaction bytes
// 2. Call ExecuteSignedTransaction with the signatures
//
// Example:
//
//	result, err := client.TransferSui(ctx, TransferSuiParams{
//		Sender:    "0x...",
//		Recipient: "0x...",
//		Amount:    1000000000, // 1 SUI
//		GasBudget: 10000000,
//	})
func (c *Client) TransferSui(ctx context.Context, params TransferSuiParams) (*TransactionResult, error) {
	tb := BuildTransferSui(params)
	return c.SimulateAndBuildResult(ctx, tb)
}

// TransferObject transfers an object from the sender to a recipient.
// Returns comprehensive transaction details.
//
// Example:
//
//	result, err := client.TransferObject(ctx, TransferObjectParams{
//		Sender:    "0x...",
//		Recipient: "0x...",
//		Object: ObjectRef{
//			ObjectID: "0x...",
//			Version:  123,
//			Digest:   "...",
//		},
//		GasBudget: 10000000,
//	})
func (c *Client) TransferObject(ctx context.Context, params TransferObjectParams) (*TransactionResult, error) {
	tb := BuildTransferObject(params)
	return c.SimulateAndBuildResult(ctx, tb)
}

// MoveCall executes a Move function call.
// Returns comprehensive transaction details including any returned values.
//
// Example:
//
//	result, err := client.MoveCall(ctx, MoveCallParams{
//		Sender:        "0x...",
//		Package:       "0x2",
//		Module:        "coin",
//		Function:      "balance",
//		TypeArguments: []string{"0x2::sui::SUI"},
//		Arguments:     []interface{}{coinObjectRef},
//		GasBudget:     10000000,
//	})
func (c *Client) MoveCall(ctx context.Context, params MoveCallParams) (*TransactionResult, error) {
	tb := BuildMoveCall(params)
	return c.SimulateAndBuildResult(ctx, tb)
}

// SplitCoins splits coins from a source coin.
// Returns comprehensive transaction details including the newly created coin objects.
//
// Example:
//
//	result, err := client.SplitCoins(ctx, SplitCoinsParams{
//		Sender:    "0x...",
//		Coin:      coinRef,
//		Amounts:   []uint64{1000000000, 2000000000}, // Split into 1 SUI and 2 SUI
//		GasBudget: 10000000,
//	})
func (c *Client) SplitCoins(ctx context.Context, params SplitCoinsParams) (*TransactionResult, error) {
	tb := BuildSplitCoins(params)
	return c.SimulateAndBuildResult(ctx, tb)
}

// MergeCoins merges multiple coins into a destination coin.
// Returns comprehensive transaction details.
//
// Example:
//
//	result, err := client.MergeCoins(ctx, MergeCoinsParams{
//		Sender:      "0x...",
//		Destination: destCoinRef,
//		Sources:     []ObjectRef{coin1, coin2},
//		GasBudget:   10000000,
//	})
func (c *Client) MergeCoins(ctx context.Context, params MergeCoinsParams) (*TransactionResult, error) {
	tb := BuildMergeCoins(params)
	return c.SimulateAndBuildResult(ctx, tb)
}

// PublishPackage publishes a Move package.
// Returns comprehensive transaction details including the published package ID.
//
// Example:
//
//	result, err := client.PublishPackage(ctx, PublishPackageParams{
//		Sender:       "0x...",
//		Modules:      []Base64{compiledModule1, compiledModule2},
//		Dependencies: []SuiAddress{"0x1", "0x2"},
//		GasBudget:    100000000,
//	})
func (c *Client) PublishPackage(ctx context.Context, params PublishPackageParams) (*TransactionResult, error) {
	tb := BuildPublishPackage(params)
	return c.SimulateAndBuildResult(ctx, tb)
}

// =============================================================================
// Transaction Execution with Signatures
// =============================================================================

// ExecuteSignedTransaction executes a signed transaction and returns comprehensive results.
// The transaction bytes should be the serialized transaction, and signatures should be
// the base64-encoded signatures.
//
// Example:
//
//	result, err := client.ExecuteSignedTransaction(ctx, txBytes, []string{signature})
func (c *Client) ExecuteSignedTransaction(ctx context.Context, txBytes Base64, signatures []string) (*TransactionResult, error) {
	// Build the comprehensive query for transaction execution
	query := `
		mutation ExecuteTransaction($tx: String!, $sigs: [String!]!) {
			executeTransaction(transactionDataBcs: $tx, signatures: $sigs) {
				effects {
					digest
					status
					executionError { message }
					lamportVersion
					epoch { epochId }
					checkpoint { sequenceNumber }
					timestamp
					gasEffects {
						gasSummary {
							computationCost
							storageCost
							storageRebate
							nonRefundableStorageFee
						}
						gasObject {
							address
							version
							digest
						}
					}
					objectChanges {
						nodes {
							address
							idCreated
							idDeleted
							inputState {
								address
								version
								digest
								owner {
									__typename
									... on AddressOwner { address { address } }
									... on Shared { initialSharedVersion }
								}
								asMoveObject {
									type { repr }
								}
							}
							outputState {
									address
									version
									digest
									owner {
										__typename
										... on AddressOwner { address { address } }
										... on Shared { initialSharedVersion }
									}
								asMoveObject {
									type { repr }
								}
							}
						}
					}
					balanceChanges {
						nodes {
							owner {
								__typename
								... on AddressOwner { address { address } }
								... on Shared { initialSharedVersion }
							}
							coinType { repr }
							amount
						}
					}
					dependencies {
						nodes {
							digest
						}
					}
				}
				errors
			}
		}
	`

	vars := map[string]any{
		"tx":   string(txBytes),
		"sigs": signatures,
	}

	var result struct {
		ExecuteTransaction *struct {
			Effects *TransactionEffects `json:"effects"`
			Errors  []string            `json:"errors"`
		} `json:"executeTransaction"`
	}

	if err := c.Execute(ctx, query, vars, &result); err != nil {
		return nil, err
	}

	if result.ExecuteTransaction == nil {
		return nil, fmt.Errorf("no execution result returned")
	}

	// Convert to comprehensive TransactionResult
	return c.convertEffectsToResult(result.ExecuteTransaction.Effects, result.ExecuteTransaction.Errors)
}

// =============================================================================
// Internal Helpers
// =============================================================================

// SimulateAndBuildResult simulates a transaction and returns comprehensive results.
func (c *Client) SimulateAndBuildResult(ctx context.Context, tb *TransactionBuilder) (*TransactionResult, error) {
	// Build comprehensive simulation query using SimulateTransactionBuilder to get auto-gas injection
	simResult, err := c.SimulateTransactionBuilder(ctx, tb, nil)
	if err != nil {
		return nil, err
	}

	if simResult == nil {
		return nil, fmt.Errorf("no simulation result returned")
	}

	var errors []string
	if simResult.Error != nil && *simResult.Error != "" {
		errors = append(errors, *simResult.Error)
	}

	return c.convertEffectsToResult(simResult.Effects, errors)
}

// convertEffectsToResult converts TransactionEffects to the comprehensive TransactionResult format.
func (c *Client) convertEffectsToResult(effects *TransactionEffects, errors []string) (*TransactionResult, error) {
	if effects == nil {
		return &TransactionResult{Errors: errors}, nil
	}

	result := &TransactionResult{
		Digest: effects.Digest,
		Errors: errors,
	}

	// Build effects result
	effectsResult := &TransactionEffectsResult{
		TransactionDigest: effects.Digest,
	}

	// Status
	effectsResult.Status = &StatusResult{
		Status: string(effects.Status),
	}
	if effects.ExecutionError != nil {
		effectsResult.Status.Error = &effects.ExecutionError.Message
	}

	// Epoch
	if effects.Epoch != nil {
		effectsResult.ExecutedEpoch = fmt.Sprintf("%d", effects.Epoch.EpochID)
	}

	// Gas used
	if effects.GasEffects != nil && effects.GasEffects.GasSummary != nil {
		gas := effects.GasEffects.GasSummary
		effectsResult.GasUsed = &GasUsedResult{
			ComputationCost:         fmt.Sprintf("%d", gas.ComputationCost),
			StorageCost:             fmt.Sprintf("%d", gas.StorageCost),
			StorageRebate:           fmt.Sprintf("%d", gas.StorageRebate),
			NonRefundableStorageFee: fmt.Sprintf("%d", gas.NonRefundableStorageFee),
		}
	}

	// Gas object
	if effects.GasEffects != nil && effects.GasEffects.GasObject != nil {
		gasObj := effects.GasEffects.GasObject
		effectsResult.GasObject = &ObjectOwnerResult{
			Reference: &ObjectRefResult{
				ObjectID: string(gasObj.Address),
				Version:  fmt.Sprintf("%d", gasObj.Version),
				Digest:   gasObj.Digest,
			},
		}
	}

	// Dependencies
	if effects.Dependencies != nil {
		for _, dep := range effects.Dependencies.Nodes {
			effectsResult.Dependencies = append(effectsResult.Dependencies, dep.Digest)
		}
	}

	result.Effects = effectsResult

	// Object changes
	if effects.ObjectChanges != nil {
		for _, change := range effects.ObjectChanges.Nodes {
			objChange := ObjectChangeResult{
				ObjectID: string(change.Address),
			}

			if change.IDCreated != nil && *change.IDCreated {
				objChange.Type = "created"
			} else if change.IDDeleted != nil && *change.IDDeleted {
				objChange.Type = "deleted"
			} else {
				objChange.Type = "mutated"
			}

			if change.OutputState != nil {
				objChange.Version = fmt.Sprintf("%d", change.OutputState.Version)
				objChange.Digest = change.OutputState.Digest
				if change.OutputState.Owner != nil {
					if change.OutputState.Owner.Address != nil {
						objChange.Owner = map[string]string{
							"AddressOwner": string(change.OutputState.Owner.Address.Address),
						}
					} else if change.OutputState.Owner.Typename == "Immutable" {
						objChange.Owner = "Immutable"
					} else if change.OutputState.Owner.InitialSharedVersion != nil {
						objChange.Owner = map[string]any{
							"Shared": map[string]any{
								"initial_shared_version": *change.OutputState.Owner.InitialSharedVersion,
							},
						}
					}
				}
				if change.OutputState.AsMoveObject != nil && change.OutputState.AsMoveObject.Contents != nil {
					objChange.ObjectType = change.OutputState.AsMoveObject.Contents.Type.Repr
				}
			}

			if change.InputState != nil {
				objChange.PreviousVersion = fmt.Sprintf("%d", change.InputState.Version)
			}

			result.ObjectChanges = append(result.ObjectChanges, objChange)
		}
	}

	// Balance changes
	if effects.BalanceChanges != nil {
		for _, change := range effects.BalanceChanges.Nodes {
			balChange := BalanceChangeResult{
				Amount: string(change.Amount),
			}
			if change.CoinType != nil {
				balChange.CoinType = change.CoinType.Repr
			}
			if change.Owner != nil {
				balChange.Owner = map[string]string{
					"AddressOwner": string(change.Owner.Address),
				}
			}
			result.BalanceChanges = append(result.BalanceChanges, balChange)
		}
	}

	// Timestamp
	if effects.Timestamp != nil {
		ts := string(*effects.Timestamp)
		result.TimestampMs = &ts
	}

	// Checkpoint
	if effects.Checkpoint != nil {
		cp := fmt.Sprintf("%d", effects.Checkpoint.SequenceNumber)
		result.Checkpoint = &cp
	}

	return result, nil
}

// =============================================================================
// Programmable Transaction Builder Convenience Methods
// =============================================================================

// TransferSuiFromGas creates a transaction that transfers SUI from the gas coin.
// This is a convenience method for simple SUI transfers.
func (c *Client) TransferSuiFromGas(ctx context.Context, sender, recipient SuiAddress, amount, gasBudget uint64) (*TransactionResult, error) {
	return c.TransferSui(ctx, TransferSuiParams{
		Sender:    sender,
		Recipient: recipient,
		Amount:    amount,
		GasBudget: gasBudget,
	})
}

// CallMoveFunction is a convenience method for calling Move functions with automatic argument conversion.
func (c *Client) CallMoveFunction(ctx context.Context, sender SuiAddress, target string, typeArgs []string, args []interface{}, gasBudget uint64) (*TransactionResult, error) {
	// Parse target (format: "package::module::function")
	parts := splitTarget(target)
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid target format, expected 'package::module::function', got '%s'", target)
	}

	return c.MoveCall(ctx, MoveCallParams{
		Sender:        sender,
		Package:       SuiAddress(parts[0]),
		Module:        parts[1],
		Function:      parts[2],
		TypeArguments: typeArgs,
		Arguments:     args,
		GasBudget:     gasBudget,
	})
}

// splitTarget splits a Move function target into package, module, and function.
func splitTarget(target string) []string {
	var parts []string
	var current string
	colonCount := 0

	for _, c := range target {
		if c == ':' {
			colonCount++
			if colonCount == 2 {
				parts = append(parts, current)
				current = ""
				colonCount = 0
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}

	return parts
}

// =============================================================================
// Batch Transaction Operations
// =============================================================================

// ExecuteMultipleTransactions executes multiple transactions and returns all results.
// Transactions are executed sequentially.
func (c *Client) ExecuteMultipleTransactions(ctx context.Context, txs []struct {
	TxBytes    Base64
	Signatures []string
}) ([]*TransactionResult, error) {
	results := make([]*TransactionResult, len(txs))

	for i, tx := range txs {
		result, err := c.ExecuteSignedTransaction(ctx, tx.TxBytes, tx.Signatures)
		if err != nil {
			return results, fmt.Errorf("transaction %d failed: %w", i, err)
		}
		results[i] = result
	}

	return results, nil
}

// =============================================================================
// Transaction Status Checking
// =============================================================================

// WaitForTransaction waits for a transaction to be confirmed and returns the full result.
func (c *Client) WaitForTransaction(ctx context.Context, digest string) (*TransactionResult, error) {
	// Query the transaction with full details
	query := `
		query GetTransaction($digest: String!) {
			transaction(digest: $digest) {
				digest
				sender { address }
				effects {
					digest
					status
					executionError { message }
					lamportVersion
					epoch { epochId }
					checkpoint { sequenceNumber }
					timestamp
					gasEffects {
						gasSummary {
							computationCost
							storageCost
							storageRebate
							nonRefundableStorageFee
						}
						gasObject {
							address
							version
							digest
						}
					}
					objectChanges {
						nodes {
							address
							idCreated
							idDeleted
							inputState {
									address
									version
									digest
									owner {
										__typename
										... on AddressOwner { address { address } }
										... on Shared { initialSharedVersion }
									}
								asMoveObject {
									contents {
										type { repr }
									}
								}
							}
							outputState {
								address
								version
								digest
								owner {
									__typename
									... on AddressOwner { address { address } }
									... on Shared { initialSharedVersion }
								}
								asMoveObject {
									contents {
										type { repr }
									}
								}
							}
						}
					}
					balanceChanges {
						nodes {
							owner {
								__typename
								... on AddressOwner { address { address } }
								... on Shared { initialSharedVersion }
							}
							coinType { repr }
							amount
						}
					}
					dependencies {
						nodes {
							digest
						}
					}
				}
			}
		}
	`

	var result struct {
		Transaction *struct {
			Digest  string              `json:"digest"`
			Sender  *Address            `json:"sender"`
			Effects *TransactionEffects `json:"effects"`
		} `json:"transaction"`
	}

	if err := c.Execute(ctx, query, map[string]any{"digest": digest}, &result); err != nil {
		return nil, err
	}

	if result.Transaction == nil {
		return nil, fmt.Errorf("transaction not found: %s", digest)
	}

	return c.convertEffectsToResult(result.Transaction.Effects, nil)
}

// GetTransactionResult fetches comprehensive transaction details by digest.
// This is an alias for WaitForTransaction but with clearer semantics for querying past transactions.
func (c *Client) GetTransactionResult(ctx context.Context, digest string) (*TransactionResult, error) {
	return c.WaitForTransaction(ctx, digest)
}
