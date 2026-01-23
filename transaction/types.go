package transaction

import (
	"github.com/iotaledger/bcs-go"
	"github.com/open-move/sui-go-sdk/types"
)

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
	TypeArguments []TypeTag
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
	Type     bcs.Option[TypeTag]
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

type StructTag struct {
	Address    types.Address
	Module     string
	Name       string
	TypeParams []TypeTag
}

type TypeTag struct {
	Bool    *struct{}
	U8      *struct{}
	U64     *struct{}
	U128    *struct{}
	Address *struct{}
	Signer  *struct{}
	Vector  *TypeTag
	Struct  *StructTag
	U16     *struct{}
	U32     *struct{}
	U256    *struct{}
}

func (TypeTag) IsBcsEnum() {}

func TypeTagBool() TypeTag {
	return TypeTag{Bool: &struct{}{}}
}

func TypeTagU8() TypeTag {
	return TypeTag{U8: &struct{}{}}
}

func TypeTagU16() TypeTag {
	return TypeTag{U16: &struct{}{}}
}

func TypeTagU32() TypeTag {
	return TypeTag{U32: &struct{}{}}
}

func TypeTagU64() TypeTag {
	return TypeTag{U64: &struct{}{}}
}

func TypeTagU128() TypeTag {
	return TypeTag{U128: &struct{}{}}
}

func TypeTagU256() TypeTag {
	return TypeTag{U256: &struct{}{}}
}

func TypeTagAddress() TypeTag {
	return TypeTag{Address: &struct{}{}}
}

func TypeTagSigner() TypeTag {
	return TypeTag{Signer: &struct{}{}}
}

func TypeTagVector(inner TypeTag) TypeTag {
	innerCopy := inner
	return TypeTag{Vector: &innerCopy}
}

func TypeTagStruct(tag StructTag) TypeTag {
	tagCopy := tag
	return TypeTag{Struct: &tagCopy}
}

func NewStructTag(address types.Address, module, name string, typeParams []TypeTag) StructTag {
	return StructTag{
		Address:    address,
		Module:     module,
		Name:       name,
		TypeParams: append([]TypeTag(nil), typeParams...),
	}
}
