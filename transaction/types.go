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

type CallArg struct {
	Pure   *Pure
	Object *ObjectArg
}

func (CallArg) IsBcsEnum() {}

type ObjectArg struct {
	ImmOrOwnedObject *types.ObjectRef
	SharedObject     *types.SharedObjectRef
	Receiving        *types.ObjectRef
}

func (ObjectArg) IsBcsEnum() {}

type Argument struct {
	GasCoin      *struct{}
	Input        *uint16
	Result       *uint16
	NestedResult *NestedResult
}

func (Argument) IsBcsEnum() {}

type NestedResult struct {
	Index       uint16
	ResultIndex uint16
}

type ProgrammableMoveCall struct {
	Package       types.Address
	Module        string
	Function      string
	TypeArguments []typetag.TypeTag
	Arguments     []Argument
}

type TransferObjects struct {
	Objects []Argument
	Address Argument
}

type SplitCoins struct {
	Coin    Argument
	Amounts []Argument
}

type MergeCoins struct {
	Destination Argument
	Sources     []Argument
}

type Publish struct {
	Modules      [][]byte
	Dependencies []types.Address
}

type MakeMoveVec struct {
	Type     bcs.Option[typetag.TypeTag]
	Elements []Argument
}

type Upgrade struct {
	Modules      [][]byte
	Dependencies []types.Address
	Package      types.Address
	Ticket       Argument
}

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

type ProgrammableTransaction struct {
	Inputs   []CallArg
	Commands []Command
}

type TransactionKind struct {
	ProgrammableTransaction *ProgrammableTransaction
	ChangeEpoch             *struct{}
	Genesis                 *struct{}
	ConsensusCommitPrologue *struct{}
}

func (TransactionKind) IsBcsEnum() {}

type TransactionExpiration struct {
	None  *struct{}
	Epoch *uint64
}

func (TransactionExpiration) IsBcsEnum() {}

func ExpirationNone() TransactionExpiration {
	return TransactionExpiration{None: &struct{}{}}
}

func ExpirationEpoch(epoch uint64) TransactionExpiration {
	e := epoch
	return TransactionExpiration{Epoch: &e}
}

type GasData struct {
	Payment []types.ObjectRef
	Owner   types.Address
	Price   uint64
	Budget  uint64
}

type TransactionDataV1 struct {
	Kind       TransactionKind
	Sender     types.Address
	GasData    GasData
	Expiration TransactionExpiration
}

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
