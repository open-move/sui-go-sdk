package graphql

import (
	"context"
	"encoding/json"
	"fmt"
)

// =============================================================================
// Transaction Simulation
// =============================================================================

// SimulateTransaction simulates a transaction without executing it.
// This is useful for estimating gas costs and verifying transaction validity.
func SimulateTransaction(c *Client, ctx context.Context, txJSON json.RawMessage, opts *SimulationOptions) (*SimulationResult, error) {
	query := `
		query SimulateTransaction($transaction: JSON!, $checksEnabled: Boolean) {
			simulateTransaction(transaction: $transaction, checksEnabled: $checksEnabled) {
				effects {
					digest
					status
					executionError { message }
					lamportVersion
					gasEffects {
						gasSummary {
							computationCost
							storageCost
							storageRebate
							nonRefundableStorageFee
						}
					}
					epoch { epochId }
					timestamp
					objectChanges {
					nodes {
						address
						idCreated
						idDeleted
					}
				}
				balanceChanges {
					nodes {
						owner {
							__typename
							address
						}
						coinType { repr }
						amount
					}
				}
				}
				error
			}
		}
	`

	checksEnabled := true
	if opts != nil && opts.ChecksEnabled != nil {
		checksEnabled = *opts.ChecksEnabled
	}

	vars := map[string]any{
		"transaction":   txJSON,
		"checksEnabled": checksEnabled,
	}

	var result struct {
		SimulateTransaction *SimulationResult `json:"simulateTransaction"`
	}

	err := c.Execute(ctx, query, vars, &result)
	if err != nil {
		return nil, err
	}

	return result.SimulateTransaction, nil
}

// SimulateTransactionBcs simulates a transaction from BCS-encoded bytes.
func SimulateTransactionBcs(c *Client, ctx context.Context, txBcs Base64, opts *SimulationOptions) (*SimulationResult, error) {
	query := `
		mutation SimulateTransaction($txBytes: String!, $skipChecks: Boolean) {
			simulateTransaction(txBytes: $txBytes, skipChecks: $skipChecks) {
				effects {
					digest
					status
					executionError { message }
					lamportVersion
					gasEffects {
						gasSummary {
							computationCost
							storageCost
							storageRebate
							nonRefundableStorageFee
						}
					}
					epoch { epochId }
					timestamp
				}
				error
			}
		}
	`

	skipChecks := false
	if opts != nil && opts.ChecksEnabled != nil && !*opts.ChecksEnabled {
		skipChecks = true
	}

	vars := map[string]any{
		"txBytes":    string(txBcs),
		"skipChecks": skipChecks,
	}

	var result struct {
		SimulateTransaction *SimulationResult `json:"simulateTransaction"`
	}

	err := c.Execute(ctx, query, vars, &result)
	if err != nil {
		return nil, err
	}

	return result.SimulateTransaction, nil
}

// =============================================================================
// Transaction Execution
// =============================================================================

// ExecuteTransaction executes a signed transaction.
// The transaction must include valid signatures.
func ExecuteTransaction(c *Client, ctx context.Context, txBcs Base64, signatures []Base64) (*ExecuteTransactionResult, error) {
	query := `
		mutation ExecuteTransaction($tx: String!, $sigs: [String!]!) {
			executeTransaction(transactionDataBcs: $tx, signatures: $sigs) {
				effects {
					digest
					status
					executionError { message }
					lamportVersion
					gasEffects {
						gasSummary {
							computationCost
							storageCost
							storageRebate
							nonRefundableStorageFee
						}
					}
					epoch { epochId }
					checkpoint { sequenceNumber }
					timestamp
					objectChanges {
						nodes {
							address
							idCreated
							idDeleted
							inputState { address version digest }
							outputState { address version digest }
						}
					}
					balanceChanges {
						nodes {
							owner { address }
							coinType { repr }
							amount
						}
					}
				}
				errors
			}
		}
	`

	// Convert signatures to strings
	sigs := make([]string, len(signatures))
	for i, sig := range signatures {
		sigs[i] = string(sig)
	}

	vars := map[string]any{
		"tx":   string(txBcs),
		"sigs": sigs,
	}

	var result struct {
		ExecuteTransaction *ExecuteTransactionResult `json:"executeTransaction"`
	}

	err := c.Execute(ctx, query, vars, &result)
	if err != nil {
		return nil, err
	}

	return result.ExecuteTransaction, nil
}

// ExecuteTransactionWithOptions executes a signed transaction with custom options.
type ExecuteOptions struct {
	// WaitForEffects waits for effects to be available (slower but more complete)
	WaitForEffects bool
	// ShowEvents includes events in the response
	ShowEvents bool
	// ShowObjectChanges includes object changes in the response
	ShowObjectChanges bool
	// ShowBalanceChanges includes balance changes in the response
	ShowBalanceChanges bool
}

// ExecuteTransactionWithOptions executes a transaction with specified options.
func ExecuteTransactionWithOptions(c *Client, ctx context.Context, txBcs Base64, signatures []Base64, opts *ExecuteOptions) (*ExecuteTransactionResult, error) {
	if opts == nil {
		opts = &ExecuteOptions{
			WaitForEffects:     true,
			ShowEvents:         true,
			ShowObjectChanges:  true,
			ShowBalanceChanges: true,
		}
	}

	// Build dynamic query based on options
	effectsFields := `
		digest
		status
		executionError { message }
		lamportVersion
		gasEffects {
			gasSummary {
				computationCost
				storageCost
				storageRebate
				nonRefundableStorageFee
			}
		}
		epoch { epochId }
		checkpoint { sequenceNumber }
		timestamp
	`

	if opts.ShowObjectChanges {
		effectsFields += `
			objectChanges {
				nodes {
					address
					idCreated
					idDeleted
					inputState { address version digest }
					outputState { address version digest }
				}
			}
		`
	}

	if opts.ShowBalanceChanges {
		effectsFields += `
					balanceChanges {
				nodes {
					owner {
						__typename
						... on AddressOwner { address { address } }
					}
					coinType { repr }
					amount
				}
			}
		`
	}

	query := fmt.Sprintf(`
		mutation ExecuteTransaction($tx: String!, $sigs: [String!]!) {
			executeTransaction(transactionDataBcs: $tx, signatures: $sigs) {
				effects {
					%s
				}
				errors
			}
		}
	`, effectsFields)

	sigs := make([]string, len(signatures))
	for i, sig := range signatures {
		sigs[i] = string(sig)
	}

	vars := map[string]any{
		"tx":   string(txBcs),
		"sigs": sigs,
	}

	var result struct {
		ExecuteTransaction *ExecuteTransactionResult `json:"executeTransaction"`
	}

	err := c.Execute(ctx, query, vars, &result)
	if err != nil {
		return nil, err
	}

	return result.ExecuteTransaction, nil
}

// =============================================================================
// ZkLogin Verification
// =============================================================================

// VerifyZkLoginSignature verifies a zkLogin signature.
func VerifyZkLoginSignature(c *Client, ctx context.Context, bytes Base64, signature Base64, intentScope ZkLoginIntentScope, author SuiAddress) (*ZkLoginVerifyResult, error) {
	query := `
		mutation VerifyZkLoginSignature($bytes: String!, $signature: String!, $intentScope: ZkLoginIntentScope!, $author: SuiAddress!) {
			verifyZkloginSignature(bytes: $bytes, signature: $signature, intentScope: $intentScope, author: $author) {
				success
				error
			}
		}
	`

	vars := map[string]any{
		"bytes":       string(bytes),
		"signature":   string(signature),
		"intentScope": intentScope,
		"author":      author,
	}

	var result struct {
		VerifyZkloginSignature *ZkLoginVerifyResult `json:"verifyZkloginSignature"`
	}

	err := c.Execute(ctx, query, vars, &result)
	if err != nil {
		return nil, err
	}

	return result.VerifyZkloginSignature, nil
}

// =============================================================================
// High-Level Transaction Execution Helpers
// =============================================================================

// TransferSuiResult contains the result of a SUI transfer.
type TransferSuiResult struct {
	Digest     string              `json:"digest"`
	Status     ExecutionStatus     `json:"status"`
	Error      *string             `json:"error"`
	GasSummary *GasCostSummary     `json:"gasSummary"`
	Effects    *TransactionEffects `json:"effects"`
	Simulation *SimulationResult   `json:"simulation,omitempty"`
}

// SimulateTransferSui simulates a SUI transfer transaction.
func SimulateTransferSui(c *Client, ctx context.Context, params TransferSuiParams) (*SimulationResult, error) {
	tb := BuildTransferSui(params)
	return c.SimulateTransactionBuilder(ctx, tb, nil)
}

// SimulateTransferObject simulates an object transfer transaction.
func SimulateTransferObject(c *Client, ctx context.Context, params TransferObjectParams) (*SimulationResult, error) {
	tb := BuildTransferObject(params)
	return c.SimulateTransactionBuilder(ctx, tb, nil)
}

// SimulateMoveCall simulates a Move function call.
func SimulateMoveCall(c *Client, ctx context.Context, params MoveCallParams) (*SimulationResult, error) {
	tb := BuildMoveCall(params)
	return c.SimulateTransactionBuilder(ctx, tb, nil)
}

// SimulateSplitCoins simulates a coin split transaction.
func SimulateSplitCoins(c *Client, ctx context.Context, params SplitCoinsParams) (*SimulationResult, error) {
	tb := BuildSplitCoins(params)
	return c.SimulateTransactionBuilder(ctx, tb, nil)
}

// SimulateMergeCoins simulates a coin merge transaction.
func SimulateMergeCoins(c *Client, ctx context.Context, params MergeCoinsParams) (*SimulationResult, error) {
	tb := BuildMergeCoins(params)
	return c.SimulateTransactionBuilder(ctx, tb, nil)
}

// SimulatePublishPackage simulates a package publish transaction.
func SimulatePublishPackage(c *Client, ctx context.Context, params PublishPackageParams) (*SimulationResult, error) {
	tb := BuildPublishPackage(params)
	return c.SimulateTransactionBuilder(ctx, tb, nil)
}

// =============================================================================
// Gas Estimation
// =============================================================================

// EstimateGas estimates the gas cost for a transaction.
func EstimateGas(c *Client, ctx context.Context, tb *TransactionBuilder) (*GasCostSummary, error) {
	result, err := c.SimulateTransactionBuilder(ctx, tb, nil)
	if err != nil {
		return nil, err
	}

	if result == nil || result.Effects == nil {
		return nil, fmt.Errorf("simulation returned no effects")
	}

	if result.Error != nil && *result.Error != "" {
		return nil, fmt.Errorf("simulation error: %s", *result.Error)
	}

	if result.Effects.GasEffects == nil {
		return nil, fmt.Errorf("simulation returned no gas effects")
	}

	return result.Effects.GasEffects.GasSummary, nil
}

// EstimateGasForTransferSui estimates gas for a SUI transfer.
func EstimateGasForTransferSui(c *Client, ctx context.Context, params TransferSuiParams) (*GasCostSummary, error) {
	tb := BuildTransferSui(params)
	return EstimateGas(c, ctx, tb)
}

// EstimateGasForMoveCall estimates gas for a Move function call.
func EstimateGasForMoveCall(c *Client, ctx context.Context, params MoveCallParams) (*GasCostSummary, error) {
	tb := BuildMoveCall(params)
	return EstimateGas(c, ctx, tb)
}

// =============================================================================
// Dry Run (Deprecated - use Simulation)
// =============================================================================

// DryRunTransaction is deprecated. Use SimulateTransaction instead.
// This method is kept for backwards compatibility.
func DryRunTransaction(c *Client, ctx context.Context, txBcs Base64) (*SimulationResult, error) {
	return SimulateTransactionBcs(c, ctx, txBcs, nil)
}
