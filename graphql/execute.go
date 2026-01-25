package graphql

import (
	"context"
	"fmt"
)

// =============================================================================
// Transaction Simulation
// =============================================================================

// SimulateTransaction simulates a transaction from BCS-encoded bytes.
func SimulateTransaction(c *Client, ctx context.Context, txBcs Base64, opts *SimulationOptions) (*SimulationResult, error) {
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

// ExecuteOptions defines options for transaction execution.
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

// ExecuteTransactionWithOptions executes a signed transaction with custom options.
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
