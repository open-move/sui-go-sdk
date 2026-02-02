package transaction

import (
	"context"

	"github.com/iotaledger/bcs-go"
	"github.com/open-move/sui-go-sdk/types"
	"github.com/open-move/sui-go-sdk/typetag"
)

const (
	OwnerUnknown OwnerKind = iota
	OwnerAddress
	OwnerObject
	OwnerShared
	OwnerImmutable
	OwnerConsensusAddress
)

const (
	ReferenceUnknown ReferenceKind = iota
	ReferenceImmutable
	ReferenceMutable
)

type OwnerKind int

type ReferenceKind int

type Pure struct {
	Bytes []byte
}

// CallArg represents an argument to a Move call.
type CallArg struct {
	Pure   *Pure
	Object *ObjectArg
}

func (CallArg) IsBcsEnum() {}

// ObjectArg represents an object argument.
type ObjectArg struct {
	ImmOrOwnedObject *types.ObjectRef
	SharedObject     *types.SharedObjectRef
	Receiving        *types.ObjectRef
}

func (ObjectArg) IsBcsEnum() {}

// Argument represents a transaction argument.
type Argument struct {
	GasCoin      *struct{}
	Input        *uint16
	Result       *uint16
	NestedResult *NestedResult
}

func (Argument) IsBcsEnum() {}

// NestedResult represents a nested result index.
type NestedResult struct {
	Index       uint16
	ResultIndex uint16
}

// ProgrammableMoveCall represents a Move call command.
type ProgrammableMoveCall struct {
	Package       types.Address
	Module        string
	Function      string
	TypeArguments []typetag.TypeTag
	Arguments     []Argument
}

// TransferObjects represents a TransferObjects command.
type TransferObjects struct {
	Objects []Argument
	Address Argument
}

// SplitCoins represents a SplitCoins command.
type SplitCoins struct {
	Coin    Argument
	Amounts []Argument
}

// MergeCoins represents a MergeCoins command.
type MergeCoins struct {
	Destination Argument
	Sources     []Argument
}

// Publish represents a Publish command.
type Publish struct {
	Modules      [][]byte
	Dependencies []types.Address
}

// MakeMoveVec represents a MakeMoveVec command.
type MakeMoveVec struct {
	Type     bcs.Option[typetag.TypeTag]
	Elements []Argument
}

// Upgrade represents an Upgrade command.
type Upgrade struct {
	Modules      [][]byte
	Dependencies []types.Address
	Package      types.Address
	Ticket       Argument
}

// Command represents a programmable transaction command.
type Command struct {
	MoveCall        *ProgrammableMoveCall
	TransferObjects *TransferObjects
	SplitCoins      *SplitCoins
	MergeCoins      *MergeCoins
	Publish         *Publish
	MakeMoveVec     *MakeMoveVec
	Upgrade         *Upgrade
}

func (Command) IsBcsEnum() {}

// ProgrammableTransaction represents a programmable transaction.
type ProgrammableTransaction struct {
	Inputs   []CallArg
	Commands []Command
}

// TransactionKind represents the kind of transaction.
type TransactionKind struct {
	ProgrammableTransaction *ProgrammableTransaction
	ChangeEpoch             *struct{}
	Genesis                 *struct{}
	ConsensusCommitPrologue *struct{}
}

func (TransactionKind) IsBcsEnum() {}

// TransactionExpiration represents the transaction expiration.
type TransactionExpiration struct {
	None  *struct{}
	Epoch *uint64
}

func (TransactionExpiration) IsBcsEnum() {}

// ExpirationNone returns a TransactionExpiration with None set.
func ExpirationNone() TransactionExpiration {
	return TransactionExpiration{None: &struct{}{}}
}

// ExpirationEpoch returns a TransactionExpiration with the given epoch.
func ExpirationEpoch(epoch uint64) TransactionExpiration {
	e := epoch
	return TransactionExpiration{Epoch: &e}
}

// GasData represents gas payment and budget information.
type GasData struct {
	Payment []types.ObjectRef
	Owner   types.Address
	Price   uint64
	Budget  uint64
}

// TransactionDataV1 represents version 1 of transaction data.
type TransactionDataV1 struct {
	Kind       TransactionKind
	Sender     types.Address
	GasData    GasData
	Expiration TransactionExpiration
}

// TransactionData represents the transaction data to be signed.
type TransactionData struct {
	V1 *TransactionDataV1
}

func (TransactionData) IsBcsEnum() {}

type GasBudgetInput struct {
	Sender     types.Address
	GasOwner   types.Address
	GasPrice   uint64
	Kind       TransactionKind
	Expiration TransactionExpiration
}

type ObjectMetadata struct {
	ID           types.ObjectID
	Version      uint64
	Digest       types.Digest
	OwnerKind    OwnerKind
	OwnerVersion *uint64
}

type MoveFunction struct {
	Parameters []MoveParameter
}

type MoveParameter struct {
	Reference ReferenceKind
	TypeName  string
}

type PackageMetadata struct {
	StorageID  string
	OriginalID string
	Version    uint64
}

type Resolver interface {
	ResolveObjects(ctx context.Context, objectIDs []string) ([]ObjectMetadata, error)
	ResolveMoveFunction(ctx context.Context, packageID, module, function string) (*MoveFunction, error)
}

type PackageResolver interface {
	ResolvePackage(ctx context.Context, packageID string) (*PackageMetadata, error)
}

type GasResolver interface {
	ResolveGasPrice(ctx context.Context) (uint64, error)
	ResolveGasBudget(ctx context.Context, input GasBudgetInput) (uint64, error)
	ResolveGasPayment(ctx context.Context, owner types.Address, budget uint64) ([]types.ObjectRef, error)
}

type TransactionSigner interface {
	SignTransaction(txBytes []byte) ([]byte, error)
	SuiAddress() (string, error)
}
