package graphql

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// NewTransactionBuilder creates a new transaction builder.
func NewTransactionBuilder() *TransactionBuilder {
	return &TransactionBuilder{
		Inputs:       make([]TransactionInputBuilder, 0),
		Transactions: make([]ProgrammableTransactionCommand, 0),
	}
}

// SetSender sets the transaction sender.
func (tb *TransactionBuilder) SetSender(sender SuiAddress) *TransactionBuilder {
	tb.Sender = &sender
	return tb
}

// SetGasBudget sets the gas budget.
func (tb *TransactionBuilder) SetGasBudget(budget uint64) *TransactionBuilder {
	tb.GasBudget = &budget
	return tb
}

// SetGasPrice sets the gas price.
func (tb *TransactionBuilder) SetGasPrice(price uint64) *TransactionBuilder {
	tb.GasPrice = &price
	return tb
}

// SetGasObject sets the gas payment object.
func (tb *TransactionBuilder) SetGasObject(obj ObjectRef) *TransactionBuilder {
	tb.GasObject = &obj
	return tb
}

// SetGasSponsor sets the gas sponsor (for sponsored transactions).
func (tb *TransactionBuilder) SetGasSponsor(sponsor SuiAddress) *TransactionBuilder {
	tb.GasSponsor = &sponsor
	return tb
}

// SetExpiration sets the transaction expiration epoch.
func (tb *TransactionBuilder) SetExpiration(epoch uint64) *TransactionBuilder {
	tb.Expiration = &EpochRef{EpochID: epoch}
	return tb
}

// AddPureInput adds a pure value input and returns its argument reference.
func (tb *TransactionBuilder) AddPureInput(typeName string, value interface{}) Argument {
	idx := len(tb.Inputs)
	tb.Inputs = append(tb.Inputs, TransactionInputBuilder{
		Kind:  "PURE",
		Type:  typeName,
		Value: value,
	})
	return Argument{Kind: ArgumentKindInput, Index: uint16(idx)}
}

// AddPureU64 adds a u64 pure input.
func (tb *TransactionBuilder) AddPureU64(value uint64) Argument {
	return tb.AddPureInput("u64", value)
}

// AddPureU8 adds a u8 pure input.
func (tb *TransactionBuilder) AddPureU8(value uint8) Argument {
	return tb.AddPureInput("u8", value)
}

// AddPureBool adds a bool pure input.
func (tb *TransactionBuilder) AddPureBool(value bool) Argument {
	return tb.AddPureInput("bool", value)
}

// AddPureAddress adds an address pure input.
func (tb *TransactionBuilder) AddPureAddress(addr SuiAddress) Argument {
	return tb.AddPureInput("address", addr)
}

// AddPureString adds a string pure input.
func (tb *TransactionBuilder) AddPureString(value string) Argument {
	return tb.AddPureInput("string", value)
}

// AddPureBytes adds raw bytes as a pure input.
func (tb *TransactionBuilder) AddPureBytes(data []byte) Argument {
	return tb.AddPureInput("vector<u8>", base64.StdEncoding.EncodeToString(data))
}

// AddObjectInput adds an owned object input.
func (tb *TransactionBuilder) AddObjectInput(ref ObjectRef) Argument {
	idx := len(tb.Inputs)
	tb.Inputs = append(tb.Inputs, TransactionInputBuilder{
		Kind:      "IMMUTABLE_OR_OWNED",
		ObjectRef: &ref,
	})
	return Argument{Kind: ArgumentKindInput, Index: uint16(idx)}
}

// AddSharedObjectInput adds a shared object input.
func (tb *TransactionBuilder) AddSharedObjectInput(objectID SuiAddress, initialSharedVersion UInt53, mutable bool) Argument {
	idx := len(tb.Inputs)
	tb.Inputs = append(tb.Inputs, TransactionInputBuilder{
		Kind: "SHARED",
		SharedInfo: &SharedInfo{
			ObjectID:             objectID,
			InitialSharedVersion: initialSharedVersion,
			Mutable:              mutable,
		},
	})
	return Argument{Kind: ArgumentKindInput, Index: uint16(idx)}
}

// GasCoin returns an argument reference to the gas coin.
func (tb *TransactionBuilder) GasCoin() Argument {
	return Argument{Kind: ArgumentKindGasCoin}
}

// Result returns an argument reference to a transaction result.
func (tb *TransactionBuilder) Result(index uint16) Argument {
	return Argument{Kind: ArgumentKindResult, Index: index}
}

// NestedResult returns an argument reference to a nested transaction result.
func (tb *TransactionBuilder) NestedResult(cmdIndex, resultIndex uint16) Argument {
	return Argument{Kind: ArgumentKindNestedResult, Index: cmdIndex, ResultIndex: resultIndex}
}

// MoveCall adds a Move call command.
func (tb *TransactionBuilder) MoveCall(target string, typeArgs []string, args []Argument) uint16 {
	idx := len(tb.Transactions)
	tb.Transactions = append(tb.Transactions, ProgrammableTransactionCommand{
		Kind:          "MoveCall",
		Target:        target,
		TypeArguments: typeArgs,
		Arguments:     args,
	})
	return uint16(idx)
}

// TransferObjects adds a transfer objects command.
func (tb *TransactionBuilder) TransferObjects(objects []Argument, recipient Argument) uint16 {
	idx := len(tb.Transactions)
	tb.Transactions = append(tb.Transactions, ProgrammableTransactionCommand{
		Kind:    "TransferObjects",
		Objects: objects,
		Address: &recipient,
	})
	return uint16(idx)
}

// SplitCoins adds a split coins command.
func (tb *TransactionBuilder) SplitCoins(coin Argument, amounts []Argument) uint16 {
	idx := len(tb.Transactions)
	tb.Transactions = append(tb.Transactions, ProgrammableTransactionCommand{
		Kind:    "SplitCoins",
		Coin:    &coin,
		Amounts: amounts,
	})
	return uint16(idx)
}

// MergeCoins adds a merge coins command.
func (tb *TransactionBuilder) MergeCoins(destination Argument, sources []Argument) uint16 {
	idx := len(tb.Transactions)
	tb.Transactions = append(tb.Transactions, ProgrammableTransactionCommand{
		Kind:    "MergeCoins",
		Coin:    &destination,
		Objects: sources,
	})
	return uint16(idx)
}

// MakeMoveVec adds a make move vector command.
func (tb *TransactionBuilder) MakeMoveVec(elementType string, elements []Argument) uint16 {
	idx := len(tb.Transactions)
	tb.Transactions = append(tb.Transactions, ProgrammableTransactionCommand{
		Kind:        "MakeMoveVec",
		ElementType: elementType,
		Objects:     elements,
	})
	return uint16(idx)
}

// Publish adds a publish command.
func (tb *TransactionBuilder) Publish(modules []Base64, dependencies []SuiAddress) uint16 {
	idx := len(tb.Transactions)
	tb.Transactions = append(tb.Transactions, ProgrammableTransactionCommand{
		Kind:         "Publish",
		Modules:      modules,
		Dependencies: dependencies,
	})
	return uint16(idx)
}

// Upgrade adds an upgrade command.
func (tb *TransactionBuilder) Upgrade(modules []Base64, dependencies []SuiAddress, packageID SuiAddress, ticket Argument) uint16 {
	idx := len(tb.Transactions)
	tb.Transactions = append(tb.Transactions, ProgrammableTransactionCommand{
		Kind:         "Upgrade",
		Modules:      modules,
		Dependencies: dependencies,
		Package:      &packageID,
		Ticket:       &ticket,
	})
	return uint16(idx)
}

// Build constructs the transaction data for simulation or signing.
func (tb *TransactionBuilder) Build() (*TransactionData, error) {
	if len(tb.Transactions) == 0 {
		return nil, fmt.Errorf("transaction has no commands")
	}

	txData := &TransactionData{
		Sender:     tb.Sender,
		Expiration: tb.Expiration,
		Kind: &TransactionKindData{
			ProgrammableTransaction: &ProgrammableTransactionData{
				Inputs:       tb.Inputs,
				Transactions: tb.Transactions,
			},
		},
	}

	// Set gas config if any gas parameters are provided
	if tb.GasBudget != nil || tb.GasPrice != nil || tb.GasSponsor != nil || tb.GasObject != nil {
		txData.GasConfig = &GasConfig{
			Budget: tb.GasBudget,
			Price:  tb.GasPrice,
			Owner:  tb.GasSponsor,
		}
		if tb.GasObject != nil {
			txData.GasConfig.Payment = []ObjectRef{*tb.GasObject}
		}
	}

	return txData, nil
}

// BuildJSON converts the transaction to JSON format for the GraphQL API.
func (tb *TransactionBuilder) BuildJSON() (json.RawMessage, error) {
	txData, err := tb.Build()
	if err != nil {
		return nil, err
	}
	return json.Marshal(txData)
}

// SimulateTransactionBuilder simulates a transaction built with the TransactionBuilder.
func (c *Client) SimulateTransactionBuilder(ctx context.Context, tb *TransactionBuilder, opts *SimulationOptions) (*SimulationResult, error) {
	// Auto-fill gas payment if missing and sender is set
	if tb.GasObject == nil && tb.Sender != nil {
		// Fetch coins for the sender
		coins, err := c.GetCoins(ctx, *tb.Sender, nil, nil)
		if err == nil && coins != nil && len(coins.Nodes) > 0 {
			// Use the first coin as gas payment
			// Note: For production, we should select coins based on balance and gas budget
			coin := coins.Nodes[0]
			tb.GasObject = &ObjectRef{
				ObjectID: coin.Address,
				Version:  coin.Version,
				Digest:   coin.Digest,
			}
		}
	}

	txJSON, err := tb.BuildJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction: %w", err)
	}
	return SimulateTransaction(c, ctx, txJSON, opts)
}

// SimulateTransaction simulates a transaction without executing it.
// This is a method wrapper around the SimulateTransaction function.
func (c *Client) SimulateTransaction(ctx context.Context, txJSON json.RawMessage, opts *SimulationOptions) (*SimulationResult, error) {
	return SimulateTransaction(c, ctx, txJSON, opts)
}

// EstimateGas estimates the gas cost for a transaction.
func (c *Client) EstimateGas(ctx context.Context, tb *TransactionBuilder) (*GasCostSummary, error) {
	return EstimateGas(c, ctx, tb)
}

// Note: dryRunTransactionKind has been deprecated in favor of simulateTransaction.
// Use SimulateTransaction or SimulateTransactionBuilder for transaction simulation.

// High-Level Transaction Helpers

// BuildTransferSui creates a transaction builder for transferring SUI.
func BuildTransferSui(params TransferSuiParams) *TransactionBuilder {
	tb := NewTransactionBuilder()
	tb.SetSender(params.Sender)
	tb.SetGasBudget(params.GasBudget)

	// Add amount as pure input
	amountArg := tb.AddPureU64(params.Amount)
	recipientArg := tb.AddPureAddress(params.Recipient)

	// Split coins from gas coin
	splitResult := tb.SplitCoins(tb.GasCoin(), []Argument{amountArg})

	// Transfer the split coin to recipient
	tb.TransferObjects([]Argument{tb.Result(splitResult)}, recipientArg)

	return tb
}

// BuildTransferObject creates a transaction builder for transferring an object.
func BuildTransferObject(params TransferObjectParams) *TransactionBuilder {
	tb := NewTransactionBuilder()
	tb.SetSender(params.Sender)
	tb.SetGasBudget(params.GasBudget)

	objectArg := tb.AddObjectInput(params.Object)
	recipientArg := tb.AddPureAddress(params.Recipient)

	tb.TransferObjects([]Argument{objectArg}, recipientArg)

	return tb
}

// BuildMoveCall creates a transaction builder for a Move function call.
func BuildMoveCall(params MoveCallParams) *TransactionBuilder {
	tb := NewTransactionBuilder()
	tb.SetSender(params.Sender)
	tb.SetGasBudget(params.GasBudget)

	// Build target string
	target := fmt.Sprintf("%s::%s::%s", params.Package, params.Module, params.Function)

	// Convert arguments
	args := make([]Argument, len(params.Arguments))
	for i, arg := range params.Arguments {
		switch v := arg.(type) {
		case Argument:
			args[i] = v
		case ObjectRef:
			args[i] = tb.AddObjectInput(v)
		case uint64:
			args[i] = tb.AddPureU64(v)
		case uint8:
			args[i] = tb.AddPureU8(v)
		case bool:
			args[i] = tb.AddPureBool(v)
		case SuiAddress:
			args[i] = tb.AddPureAddress(v)
		case string:
			args[i] = tb.AddPureString(v)
		case []byte:
			args[i] = tb.AddPureBytes(v)
		default:
			// For complex types, try to marshal as JSON and add as pure input
			args[i] = tb.AddPureInput("", v)
		}
	}

	tb.MoveCall(target, params.TypeArguments, args)

	return tb
}

// BuildSplitCoins creates a transaction builder for splitting coins.
func BuildSplitCoins(params SplitCoinsParams) *TransactionBuilder {
	tb := NewTransactionBuilder()
	tb.SetSender(params.Sender)
	tb.SetGasBudget(params.GasBudget)

	coinArg := tb.AddObjectInput(params.Coin)

	amountArgs := make([]Argument, len(params.Amounts))
	for i, amount := range params.Amounts {
		amountArgs[i] = tb.AddPureU64(amount)
	}

	tb.SplitCoins(coinArg, amountArgs)

	return tb
}

// BuildMergeCoins creates a transaction builder for merging coins.
func BuildMergeCoins(params MergeCoinsParams) *TransactionBuilder {
	tb := NewTransactionBuilder()
	tb.SetSender(params.Sender)
	tb.SetGasBudget(params.GasBudget)

	destArg := tb.AddObjectInput(params.Destination)

	sourceArgs := make([]Argument, len(params.Sources))
	for i, source := range params.Sources {
		sourceArgs[i] = tb.AddObjectInput(source)
	}

	tb.MergeCoins(destArg, sourceArgs)

	return tb
}

// BuildPublishPackage creates a transaction builder for publishing a package.
func BuildPublishPackage(params PublishPackageParams) *TransactionBuilder {
	tb := NewTransactionBuilder()
	tb.SetSender(params.Sender)
	tb.SetGasBudget(params.GasBudget)

	tb.Publish(params.Modules, params.Dependencies)

	return tb
}
