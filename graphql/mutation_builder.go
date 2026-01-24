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

// =============================================================================
// Chained Transaction Builder (for complex multi-step transactions)
// =============================================================================

// ChainedTransactionBuilder allows building complex transactions step by step.
type ChainedTransactionBuilder struct {
	*TransactionBuilder
	client  *Client
	results []Argument
}

// NewChainedTransaction creates a new chained transaction builder.
func NewChainedTransaction(client *Client, sender SuiAddress) *ChainedTransactionBuilder {
	tb := NewTransactionBuilder()
	tb.SetSender(sender)
	return &ChainedTransactionBuilder{
		TransactionBuilder: tb,
		client:             client,
		results:            make([]Argument, 0),
	}
}

// Pay creates a payment from the gas coin to a recipient.
// Returns the command index for chaining.
func (ctb *ChainedTransactionBuilder) Pay(recipient SuiAddress, amount uint64) *ChainedTransactionBuilder {
	amountArg := ctb.AddPureU64(amount)
	recipientArg := ctb.AddPureAddress(recipient)
	splitIdx := ctb.SplitCoins(ctb.GasCoin(), []Argument{amountArg})
	ctb.TransferObjects([]Argument{ctb.Result(splitIdx)}, recipientArg)
	return ctb
}

// PayMultiple creates multiple payments from the gas coin.
func (ctb *ChainedTransactionBuilder) PayMultiple(payments []Payment) *ChainedTransactionBuilder {
	for _, p := range payments {
		amountArg := ctb.AddPureU64(p.Amount)
		recipientArg := ctb.AddPureAddress(p.Recipient)
		splitIdx := ctb.SplitCoins(ctb.GasCoin(), []Argument{amountArg})
		ctb.TransferObjects([]Argument{ctb.Result(splitIdx)}, recipientArg)
	}
	return ctb
}

// Payment represents a single payment.
type Payment struct {
	Recipient SuiAddress
	Amount    uint64
}

// Call executes a Move function call.
func (ctb *ChainedTransactionBuilder) Call(target string, typeArgs []string, args []Argument) *ChainedTransactionBuilder {
	ctb.MoveCall(target, typeArgs, args)
	return ctb
}

// Transfer transfers objects to a recipient.
func (ctb *ChainedTransactionBuilder) Transfer(objects []Argument, recipient SuiAddress) *ChainedTransactionBuilder {
	recipientArg := ctb.AddPureAddress(recipient)
	ctb.TransferObjects(objects, recipientArg)
	return ctb
}

// WithGasBudget sets the gas budget.
func (ctb *ChainedTransactionBuilder) WithGasBudget(budget uint64) *ChainedTransactionBuilder {
	ctb.SetGasBudget(budget)
	return ctb
}

// WithGasPrice sets the gas price.
func (ctb *ChainedTransactionBuilder) WithGasPrice(price uint64) *ChainedTransactionBuilder {
	ctb.SetGasPrice(price)
	return ctb
}

// WithSponsor sets the gas sponsor for sponsored transactions.
func (ctb *ChainedTransactionBuilder) WithSponsor(sponsor SuiAddress) *ChainedTransactionBuilder {
	ctb.SetGasSponsor(sponsor)
	return ctb
}

// Simulate simulates the transaction and returns the result.
func (ctb *ChainedTransactionBuilder) Simulate(ctx context.Context) (*SimulationResult, error) {
	return ctb.client.SimulateTransactionBuilder(ctx, ctb.TransactionBuilder, nil)
}

// EstimateGas estimates the gas cost for the transaction.
func (ctb *ChainedTransactionBuilder) EstimateGas(ctx context.Context) (*GasCostSummary, error) {
	return ctb.client.EstimateGas(ctx, ctb.TransactionBuilder)
}

// =============================================================================
// Batch Transaction Builder
// =============================================================================

// BatchTransactionBuilder helps build multiple transactions for batch processing.
type BatchTransactionBuilder struct {
	transactions []*TransactionBuilder
	client       *Client
}

// NewBatchTransaction creates a new batch transaction builder.
func NewBatchTransaction(client *Client) *BatchTransactionBuilder {
	return &BatchTransactionBuilder{
		transactions: make([]*TransactionBuilder, 0),
		client:       client,
	}
}

// Add adds a transaction to the batch.
func (btb *BatchTransactionBuilder) Add(tb *TransactionBuilder) *BatchTransactionBuilder {
	btb.transactions = append(btb.transactions, tb)
	return btb
}

// AddTransferSui adds a SUI transfer to the batch.
func (btb *BatchTransactionBuilder) AddTransferSui(params TransferSuiParams) *BatchTransactionBuilder {
	btb.transactions = append(btb.transactions, BuildTransferSui(params))
	return btb
}

// AddTransferObject adds an object transfer to the batch.
func (btb *BatchTransactionBuilder) AddTransferObject(params TransferObjectParams) *BatchTransactionBuilder {
	btb.transactions = append(btb.transactions, BuildTransferObject(params))
	return btb
}

// AddMoveCall adds a Move call to the batch.
func (btb *BatchTransactionBuilder) AddMoveCall(params MoveCallParams) *BatchTransactionBuilder {
	btb.transactions = append(btb.transactions, BuildMoveCall(params))
	return btb
}

// SimulateAll simulates all transactions in the batch.
func (btb *BatchTransactionBuilder) SimulateAll(ctx context.Context) ([]*SimulationResult, error) {
	results := make([]*SimulationResult, len(btb.transactions))

	for i, tb := range btb.transactions {
		result, err := btb.client.SimulateTransactionBuilder(ctx, tb, nil)
		if err != nil {
			return nil, fmt.Errorf("transaction %d simulation failed: %w", i, err)
		}
		results[i] = result
	}

	return results, nil
}

// EstimateAllGas estimates gas for all transactions.
func (btb *BatchTransactionBuilder) EstimateAllGas(ctx context.Context) ([]*GasCostSummary, error) {
	results := make([]*GasCostSummary, len(btb.transactions))

	for i, tb := range btb.transactions {
		gas, err := btb.client.EstimateGas(ctx, tb)
		if err != nil {
			return nil, fmt.Errorf("transaction %d gas estimation failed: %w", i, err)
		}
		results[i] = gas
	}

	return results, nil
}

// Count returns the number of transactions in the batch.
func (btb *BatchTransactionBuilder) Count() int {
	return len(btb.transactions)
}

// Get returns the transaction at the specified index.
func (btb *BatchTransactionBuilder) Get(index int) *TransactionBuilder {
	if index < 0 || index >= len(btb.transactions) {
		return nil
	}
	return btb.transactions[index]
}

// =============================================================================
// Sponsored Transaction Builder
// =============================================================================

// SponsoredTransactionBuilder helps build sponsored transactions.
type SponsoredTransactionBuilder struct {
	*TransactionBuilder
	sponsor SuiAddress
	client  *Client
}

// NewSponsoredTransaction creates a new sponsored transaction builder.
func NewSponsoredTransaction(client *Client, sender, sponsor SuiAddress) *SponsoredTransactionBuilder {
	tb := NewTransactionBuilder()
	tb.SetSender(sender)
	tb.SetGasSponsor(sponsor)
	return &SponsoredTransactionBuilder{
		TransactionBuilder: tb,
		sponsor:            sponsor,
		client:             client,
	}
}

// Pay creates a payment from the sender's gas coin to a recipient.
func (stb *SponsoredTransactionBuilder) Pay(recipient SuiAddress, amount uint64) *SponsoredTransactionBuilder {
	amountArg := stb.AddPureU64(amount)
	recipientArg := stb.AddPureAddress(recipient)
	splitIdx := stb.SplitCoins(stb.GasCoin(), []Argument{amountArg})
	stb.TransferObjects([]Argument{stb.Result(splitIdx)}, recipientArg)
	return stb
}

// Call executes a Move function call.
func (stb *SponsoredTransactionBuilder) Call(target string, typeArgs []string, args []Argument) *SponsoredTransactionBuilder {
	stb.MoveCall(target, typeArgs, args)
	return stb
}

// WithGasBudget sets the gas budget.
func (stb *SponsoredTransactionBuilder) WithGasBudget(budget uint64) *SponsoredTransactionBuilder {
	stb.SetGasBudget(budget)
	return stb
}

// Simulate simulates the sponsored transaction.
func (stb *SponsoredTransactionBuilder) Simulate(ctx context.Context) (*SimulationResult, error) {
	return stb.client.SimulateTransactionBuilder(ctx, stb.TransactionBuilder, nil)
}
