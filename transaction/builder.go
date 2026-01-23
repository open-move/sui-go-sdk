package transaction

import (
	"context"
	"fmt"
	"math/big"

	bcs "github.com/iotaledger/bcs-go"
	v2 "github.com/open-move/sui-go-sdk/proto/sui/rpc/v2"
	"github.com/open-move/sui-go-sdk/types"
	"github.com/open-move/sui-go-sdk/utils"
)

type BuildOptions struct {
	Resolver ObjectResolver
}

type BuildResult struct {
	KindBytes         []byte
	TransactionBytes  []byte
	Transaction       *v2.Transaction
	ProgrammableKind  *ProgrammableTransaction
	ResolvedInputArgs []CallArg
}

type Builder struct {
	inputs     []input
	commands   []Command
	sender     *types.Address
	expiration *TransactionExpiration
	gas        gasConfig
	err        error
}

type input struct {
	Pure             *Pure
	Object           *ObjectArg
	UnresolvedObject *UnresolvedObject
}

type UnresolvedObject struct {
	ObjectID string
}

type gasConfig struct {
	Payment []types.ObjectRef
	Owner   *types.Address
	Price   *uint64
	Budget  *uint64
}

func New() *Builder {
	return &Builder{}
}

func (b *Builder) Err() error {
	if b == nil {
		return ErrNilBuilder
	}

	return b.err
}

func (b *Builder) SetSender(address string) *Builder {
	if b == nil {
		return b
	}

	addr, err := utils.ParseAddress(address)
	if err != nil {
		b.setErr(err)
		return b
	}

	b.sender = &addr
	return b
}

func (b *Builder) SetExpiration(expiration TransactionExpiration) *Builder {
	if b == nil {
		return b
	}

	b.expiration = &expiration
	return b
}

func (b *Builder) SetGasBudget(budget uint64) *Builder {
	if b == nil {
		return b
	}

	b.gas.Budget = &budget
	return b
}

func (b *Builder) SetGasPrice(price uint64) *Builder {
	if b == nil {
		return b
	}

	b.gas.Price = &price
	return b
}

func (b *Builder) SetGasOwner(address string) *Builder {
	if b == nil {
		return b
	}

	addr, err := utils.ParseAddress(address)
	if err != nil {
		b.setErr(err)
		return b
	}

	b.gas.Owner = &addr
	return b
}

func (b *Builder) SetGasPayment(payment []types.ObjectRef) *Builder {
	if b == nil {
		return b
	}

	b.gas.Payment = append([]types.ObjectRef(nil), payment...)
	return b
}

func (b *Builder) Gas() Argument {
	return Argument{GasCoin: &struct{}{}}
}

func (b *Builder) PureBytes(value []byte) Argument {
	if b == nil {
		return Argument{}
	}

	if b.err != nil {
		return Argument{}
	}

	return b.addInput(input{Pure: &Pure{Bytes: append([]byte(nil), value...)}})
}

func (b *Builder) PureBool(value bool) Argument {
	bytes, err := bcs.Marshal(&value)
	return b.pureEncoded(bytes, err)
}

func (b *Builder) PureU8(value uint8) Argument {
	bytes, err := bcs.Marshal(&value)
	return b.pureEncoded(bytes, err)
}

func (b *Builder) PureU16(value uint16) Argument {
	bytes, err := bcs.Marshal(&value)
	return b.pureEncoded(bytes, err)
}

func (b *Builder) PureU32(value uint32) Argument {
	bytes, err := bcs.Marshal(&value)
	return b.pureEncoded(bytes, err)
}

func (b *Builder) PureU64(value uint64) Argument {
	bytes, err := bcs.Marshal(&value)
	return b.pureEncoded(bytes, err)
}

func (b *Builder) PureU128(value *big.Int) Argument {
	if value == nil {
		b.setErr(fmt.Errorf("u128 value is nil"))
		return Argument{}
	}

	bytes, err := bcs.Marshal(value)
	if err != nil {
		b.setErr(err)
		return Argument{}
	}

	return b.PureBytes(bytes)
}

func (b *Builder) PureU256(value *big.Int) Argument {
	if value == nil {
		b.setErr(fmt.Errorf("u256 value is nil"))
		return Argument{}
	}

	bytes, err := encodeU256(value)
	if err != nil {
		b.setErr(err)
		return Argument{}
	}

	return b.PureBytes(bytes)
}

func (b *Builder) PureString(value string) Argument {
	bytes, err := bcs.Marshal(&value)
	return b.pureEncoded(bytes, err)
}

func (b *Builder) PureAddress(value string) Argument {
	if b == nil {
		return Argument{}
	}

	addr, err := utils.ParseAddress(value)
	if err != nil {
		b.setErr(err)
		return Argument{}
	}

	bytes, err := bcs.Marshal(&addr)
	if err != nil {
		b.setErr(err)
		return Argument{}
	}

	return b.PureBytes(bytes)
}

func (b *Builder) Object(id string) Argument {
	if b == nil {
		return Argument{}
	}

	normalized, err := utils.NormalizeAddress(id)
	if err != nil {
		b.setErr(err)
		return Argument{}
	}

	return b.addInput(input{UnresolvedObject: &UnresolvedObject{ObjectID: normalized}})
}

func (b *Builder) ObjectRef(ref types.ObjectRef) Argument {
	if b == nil {
		return Argument{}
	}

	return b.addInput(input{Object: &ObjectArg{ImmOrOwnedObject: &ref}})
}

func (b *Builder) SharedObject(ref types.SharedObjectRef) Argument {
	if b == nil {
		return Argument{}
	}

	return b.addInput(input{Object: &ObjectArg{SharedObject: &ref}})
}

func (b *Builder) ReceivingObject(ref types.ObjectRef) Argument {
	if b == nil {
		return Argument{}
	}

	return b.addInput(input{Object: &ObjectArg{Receiving: &ref}})
}

func (b *Builder) SplitCoins(args SplitCoins) []Argument {
	idx := b.addCommand(Command{SplitCoins: &args})
	if idx == nil {
		return nil
	}

	count := len(args.Amounts)
	results := make([]Argument, count)
	for i := 0; i < count; i++ {
		results[i] = nestedResultArg(*idx, uint16(i))
	}

	return results
}

func (b *Builder) MergeCoins(args MergeCoins) {
	b.addCommand(Command{MergeCoins: &args})
}

func (b *Builder) TransferObjects(args TransferObjects) {
	b.addCommand(Command{TransferObjects: &args})
}

func (b *Builder) MoveCall(args MoveCall) Result {
	call, err := args.toProgrammableMoveCall()
	if err != nil {
		b.setErr(err)
		return Result{}
	}

	idx := b.addCommand(Command{MoveCall: &call})
	if idx == nil {
		return Result{}
	}

	return Result{Index: *idx}
}

func (b *Builder) MakeMoveVec(args MakeMoveVecInput) Result {
	command, err := args.toCommand()
	if err != nil {
		b.setErr(err)
		return Result{}
	}
	idx := b.addCommand(Command{MakeMoveVec: &command})
	if idx == nil {
		return Result{}
	}
	return Result{Index: *idx}
}

func (b *Builder) Publish(args PublishInput) Result {
	command, err := args.toCommand()
	if err != nil {
		b.setErr(err)
		return Result{}
	}

	idx := b.addCommand(Command{Publish: &command})
	if idx == nil {
		return Result{}
	}

	return Result{Index: *idx}
}

func (b *Builder) Upgrade(args UpgradeInput) Result {
	command, err := args.toCommand()
	if err != nil {
		b.setErr(err)
		return Result{}
	}

	idx := b.addCommand(Command{Upgrade: &command})
	if idx == nil {
		return Result{}
	}

	return Result{Index: *idx}
}

func (b *Builder) Build(ctx context.Context, opts BuildOptions) (BuildResult, error) {
	if b == nil {
		return BuildResult{}, ErrNilBuilder
	}

	if b.err != nil {
		return BuildResult{}, b.err
	}

	resolvedInputs, err := b.resolveInputs(ctx, opts.Resolver)
	if err != nil {
		return BuildResult{}, err
	}

	programmable := &ProgrammableTransaction{
		Inputs:   resolvedInputs,
		Commands: append([]Command(nil), b.commands...),
	}

	kind := TransactionKind{ProgrammableTransaction: programmable}
	kindBytes, err := bcs.Marshal(&kind)
	if err != nil {
		return BuildResult{}, err
	}

	result := BuildResult{
		KindBytes:         kindBytes,
		ProgrammableKind:  programmable,
		ResolvedInputArgs: resolvedInputs,
	}
	if !b.hasFullTransaction() {
		return result, nil
	}

	gasData := b.buildGasData()
	expiration := b.expiration
	if expiration == nil {
		fallback := ExpirationNone()
		expiration = &fallback
	}

	data := TransactionData{
		V1: &TransactionDataV1{
			Kind:       kind,
			Sender:     *b.sender,
			GasData:    gasData,
			Expiration: *expiration,
		},
	}

	bytes, err := bcs.Marshal(&data)
	if err != nil {
		return BuildResult{}, err
	}

	result.TransactionBytes = bytes
	result.Transaction = &v2.Transaction{Bcs: &v2.Bcs{Name: utils.StringPtr("TransactionData"), Value: bytes}}
	return result, nil
}

func (b *Builder) addInput(in input) Argument {
	if b.err != nil {
		return Argument{}
	}

	idx, err := nextIndex(len(b.inputs))
	if err != nil {
		b.setErr(err)
		return Argument{}
	}

	b.inputs = append(b.inputs, in)
	return Argument{Input: &idx}
}

func (b *Builder) addCommand(cmd Command) *uint16 {
	if b.err != nil {
		return nil
	}

	idx, err := nextIndex(len(b.commands))
	if err != nil {
		b.setErr(err)
		return nil
	}

	b.commands = append(b.commands, cmd)
	return &idx
}

func (b *Builder) resolveInputs(ctx context.Context, resolver ObjectResolver) ([]CallArg, error) {
	resolved := make([]CallArg, len(b.inputs))
	objectIDs := make([]string, 0)
	for _, in := range b.inputs {
		if in.UnresolvedObject != nil {
			objectIDs = append(objectIDs, in.UnresolvedObject.ObjectID)
		}
	}

	objectMap := map[string]types.ObjectRef{}
	if len(objectIDs) > 0 {
		if resolver == nil {
			return nil, ErrResolverRequired
		}

		if ctx == nil {
			return nil, fmt.Errorf("nil context")
		}

		unique := uniqueStrings(objectIDs)
		refs, err := resolver.ResolveObjects(ctx, unique)
		if err != nil {
			return nil, err
		}

		if len(refs) != len(unique) {
			return nil, fmt.Errorf("resolver returned %d refs for %d object ids", len(refs), len(unique))
		}

		for i, id := range unique {
			objectMap[id] = refs[i]
		}
	}

	for i, in := range b.inputs {
		switch {
		case in.Pure != nil:
			resolved[i] = CallArg{Pure: in.Pure}
		case in.Object != nil:
			resolved[i] = CallArg{Object: in.Object}
		case in.UnresolvedObject != nil:
			ref, ok := objectMap[in.UnresolvedObject.ObjectID]
			if !ok {
				return nil, ErrUnresolvedInput
			}
			resolved[i] = CallArg{Object: &ObjectArg{ImmOrOwnedObject: &ref}}
		default:
			return nil, ErrUnresolvedInput
		}
	}

	return resolved, nil
}

func (b *Builder) hasFullTransaction() bool {
	if b.sender == nil {
		return false
	}

	if b.gas.Budget == nil || b.gas.Price == nil || len(b.gas.Payment) == 0 {
		return false
	}

	return true
}

func (b *Builder) buildGasData() GasData {
	owner := b.gas.Owner
	if owner == nil {
		owner = b.sender
	}

	return GasData{
		Payment: append([]types.ObjectRef(nil), b.gas.Payment...),
		Owner:   *owner,
		Price:   *b.gas.Price,
		Budget:  *b.gas.Budget,
	}
}

func (b *Builder) pureEncoded(bytes []byte, err error) Argument {
	if b == nil {
		return Argument{}
	}

	if err != nil {
		b.setErr(err)
		return Argument{}
	}

	return b.PureBytes(bytes)
}

func (b *Builder) setErr(err error) {
	if err != nil && b.err == nil {
		b.err = err
	}
}

const maxIndex = int(^uint16(0))

func nextIndex(length int) (uint16, error) {
	if length > maxIndex {
		return 0, fmt.Errorf("transaction index exceeds %d", maxIndex)
	}

	return uint16(length), nil
}

func nestedResultArg(index uint16, resultIndex uint16) Argument {
	return Argument{NestedResult: &NestedResult{Index: index, ResultIndex: resultIndex}}
}

func encodeU256(value *big.Int) ([]byte, error) {
	if value.Sign() < 0 {
		return nil, fmt.Errorf("u256 value must be positive")
	}

	if value.BitLen() > 256 {
		return nil, fmt.Errorf("u256 value out of range")
	}

	buf := make([]byte, 32)
	bigBytes := value.Bytes()
	copy(buf[32-len(bigBytes):], bigBytes)
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}

	return buf, nil
}

func uniqueStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(values))
	unique := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		unique = append(unique, value)
	}

	return unique
}
