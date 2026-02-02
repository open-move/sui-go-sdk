package graphql

import (
	"context"
	"encoding/base64"
	"fmt"
)

// =============================================================================
// MutationBuilder - Fluent API for Building GraphQL Mutations
// =============================================================================

// MutationBuilder provides a fluent API for building custom GraphQL mutations.
// It extends the QueryBuilder with mutation-specific functionality.
type MutationBuilder struct {
	*QueryBuilder
}

// NewMutation creates a new mutation builder with a name.
func NewMutation(name string) *MutationBuilder {
	qb := NewMutationBuilder()
	qb.Name(name)
	return &MutationBuilder{QueryBuilder: qb}
}

// SimulateMutation creates a mutation for simulating a transaction.
func SimulateMutation() *SimulateMutationBuilder {
	return &SimulateMutationBuilder{
		mb: NewMutation("SimulateTransaction"),
	}
}

// ExecuteMutation creates a mutation for executing a transaction.
func ExecuteMutation() *ExecuteMutationBuilder {
	return &ExecuteMutationBuilder{
		mb: NewMutation("ExecuteTransaction"),
	}
}

// =============================================================================
// SimulateMutationBuilder
// =============================================================================

// SimulateMutationBuilder helps build simulation mutations.
type SimulateMutationBuilder struct {
	mb           *MutationBuilder
	txBytes      string
	skipChecks   bool
	effectFields []string
}

// TxBytes sets the transaction bytes.
func (smb *SimulateMutationBuilder) TxBytes(bytes string) *SimulateMutationBuilder {
	smb.txBytes = bytes
	return smb
}

// TxBytesBase64 sets the transaction bytes from raw bytes.
func (smb *SimulateMutationBuilder) TxBytesBase64(data []byte) *SimulateMutationBuilder {
	smb.txBytes = base64.StdEncoding.EncodeToString(data)
	return smb
}

// SkipChecks sets whether to skip validation checks.
func (smb *SimulateMutationBuilder) SkipChecks(skip bool) *SimulateMutationBuilder {
	smb.skipChecks = skip
	return smb
}

// WithEffects specifies which effect fields to return.
func (smb *SimulateMutationBuilder) WithEffects(fields ...string) *SimulateMutationBuilder {
	smb.effectFields = fields
	return smb
}

// WithAllEffects requests all effect fields.
func (smb *SimulateMutationBuilder) WithAllEffects() *SimulateMutationBuilder {
	smb.effectFields = []string{
		"digest", "status", "lamportVersion", "timestamp",
		"gasEffects", "objectChanges", "balanceChanges",
	}
	return smb
}

// Build generates the GraphQL mutation.
func (smb *SimulateMutationBuilder) Build() (string, map[string]any) {
	effectsBlock := smb.buildEffectsBlock()

	query := fmt.Sprintf(`
		mutation SimulateTransaction($txBytes: String!, $skipChecks: Boolean) {
			simulateTransaction(txBytes: $txBytes, skipChecks: $skipChecks) {
				effects {
					%s
				}
				error
			}
		}
	`, effectsBlock)

	vars := map[string]any{
		"txBytes":    smb.txBytes,
		"skipChecks": smb.skipChecks,
	}

	return query, vars
}

// buildEffectsBlock builds the effects selection block for the query.
// buildEffectsBlock constructs the effects selection block for the mutation.
func (smb *SimulateMutationBuilder) buildEffectsBlock() string {
	if len(smb.effectFields) == 0 {
		return `
			digest
			status
			executionError { message }
			gasEffects {
				gasSummary {
					computationCost
					storageCost
					storageRebate
					nonRefundableStorageFee
				}
			}
		`
	}

	result := ""
	for _, field := range smb.effectFields {
		switch field {
		case "digest", "status", "lamportVersion", "timestamp":
			result += field + "\n"
		case "gasEffects":
			result += `
				gasEffects {
					gasSummary {
						computationCost
						storageCost
						storageRebate
						nonRefundableStorageFee
					}
				}
			`
		case "objectChanges":
			result += `
				objectChanges {
					nodes {
						address
						idCreated
						idDeleted
					}
				}
			`
		case "balanceChanges":
			result += `
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
		case "executionError":
			result += "executionError { message }\n"
		}
	}
	return result
}

// Execute runs the simulation mutation.
func (smb *SimulateMutationBuilder) Execute(ctx context.Context, client *Client) (*SimulationResult, error) {
	query, vars := smb.Build()

	var result struct {
		SimulateTransaction *SimulationResult `json:"simulateTransaction"`
	}

	err := client.Execute(ctx, query, vars, &result)
	if err != nil {
		return nil, err
	}

	return result.SimulateTransaction, nil
}

// =============================================================================
// ExecuteMutationBuilder
// =============================================================================

// ExecuteMutationBuilder helps build execution mutations.
type ExecuteMutationBuilder struct {
	mb           *MutationBuilder
	txBytes      string
	signatures   []string
	effectFields []string
}

// TxBytes sets the transaction bytes.
func (emb *ExecuteMutationBuilder) TxBytes(bytes string) *ExecuteMutationBuilder {
	emb.txBytes = bytes
	return emb
}

// TxBytesBase64 sets the transaction bytes from raw bytes.
func (emb *ExecuteMutationBuilder) TxBytesBase64(data []byte) *ExecuteMutationBuilder {
	emb.txBytes = base64.StdEncoding.EncodeToString(data)
	return emb
}

// Signatures sets the transaction signatures.
func (emb *ExecuteMutationBuilder) Signatures(sigs ...string) *ExecuteMutationBuilder {
	emb.signatures = sigs
	return emb
}

// SignaturesBase64 sets signatures from raw bytes.
func (emb *ExecuteMutationBuilder) SignaturesBase64(sigs ...[]byte) *ExecuteMutationBuilder {
	emb.signatures = make([]string, len(sigs))
	for i, sig := range sigs {
		emb.signatures[i] = base64.StdEncoding.EncodeToString(sig)
	}
	return emb
}

// WithEffects specifies which effect fields to return.
func (emb *ExecuteMutationBuilder) WithEffects(fields ...string) *ExecuteMutationBuilder {
	emb.effectFields = fields
	return emb
}

// WithAllEffects requests all effect fields.
func (emb *ExecuteMutationBuilder) WithAllEffects() *ExecuteMutationBuilder {
	emb.effectFields = []string{
		"digest", "status", "lamportVersion", "timestamp",
		"gasEffects", "objectChanges", "balanceChanges", "epoch", "checkpoint",
	}
	return emb
}

// Build generates the GraphQL mutation.
func (emb *ExecuteMutationBuilder) Build() (string, map[string]any) {
	effectsBlock := emb.buildEffectsBlock()

	query := fmt.Sprintf(`
		mutation ExecuteTransaction($tx: String!, $sigs: [String!]!) {
			executeTransaction(transactionDataBcs: $tx, signatures: $sigs) {
				effects {
					%s
				}
				errors
			}
		}
	`, effectsBlock)

	vars := map[string]any{
		"tx":   emb.txBytes,
		"sigs": emb.signatures,
	}

	return query, vars
}

// buildEffectsBlock constructs the effects selection block for the mutation.
func (emb *ExecuteMutationBuilder) buildEffectsBlock() string {
	if len(emb.effectFields) == 0 {
		return `
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
	}

	result := ""
	for _, field := range emb.effectFields {
		switch field {
		case "digest", "status", "lamportVersion", "timestamp":
			result += field + "\n"
		case "epoch":
			result += "epoch { epochId }\n"
		case "checkpoint":
			result += "checkpoint { sequenceNumber }\n"
		case "gasEffects":
			result += `
				gasEffects {
					gasSummary {
						computationCost
						storageCost
						storageRebate
						nonRefundableStorageFee
					}
				}
			`
		case "objectChanges":
			result += `
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
		case "balanceChanges":
			result += `
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
		case "executionError":
			result += "executionError { message }\n"
		}
	}
	return result
}

// Execute runs the execution mutation.
func (emb *ExecuteMutationBuilder) Execute(ctx context.Context, client *Client) (*ExecuteTransactionResult, error) {
	query, vars := emb.Build()

	var result struct {
		ExecuteTransaction *ExecuteTransactionResult `json:"executeTransaction"`
	}

	err := client.Execute(ctx, query, vars, &result)
	if err != nil {
		return nil, err
	}

	return result.ExecuteTransaction, nil
}
