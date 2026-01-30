package transaction

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	bcs "github.com/iotaledger/bcs-go"
	v2 "github.com/open-move/sui-go-sdk/proto/sui/rpc/v2"
	"github.com/open-move/sui-go-sdk/types"
	"github.com/open-move/sui-go-sdk/utils"
)

type BuildOptions struct {
	Resolver    Resolver
	GasResolver GasResolver
}

type BuildResult struct {
	KindBytes         []byte
	TransactionBytes  []byte
	Transaction       *v2.Transaction
	ProgrammableKind  *ProgrammableTransaction
	ResolvedInputArgs []CallArg
}

type Transaction struct {
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

func New() *Transaction {
	return &Transaction{}
}

func (b *Transaction) Err() error {
	if b == nil {
		return ErrNilTransaction
	}

	return b.err
}

func (b *Transaction) HasSender() bool {
	return b != nil && b.sender != nil
}

func (b *Transaction) SetSender(address string) *Transaction {
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

func (b *Transaction) SetExpiration(expiration TransactionExpiration) *Transaction {
	if b == nil {
		return b
	}

	b.expiration = &expiration
	return b
}

func (b *Transaction) SetGasBudget(budget uint64) *Transaction {
	if b == nil {
		return b
	}

	b.gas.Budget = &budget
	return b
}

func (b *Transaction) SetGasPrice(price uint64) *Transaction {
	if b == nil {
		return b
	}

	b.gas.Price = &price
	return b
}

func (b *Transaction) SetGasOwner(address string) *Transaction {
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

func (b *Transaction) SetGasPayment(payment []types.ObjectRef) *Transaction {
	if b == nil {
		return b
	}

	b.gas.Payment = append([]types.ObjectRef(nil), payment...)
	return b
}

func (b *Transaction) Gas() Argument {
	return Argument{GasCoin: &struct{}{}}
}

func (b *Transaction) PureBytes(value []byte) Argument {
	if b == nil {
		return Argument{}
	}

	if b.err != nil {
		return Argument{}
	}

	return b.addInput(input{Pure: &Pure{Bytes: append([]byte(nil), value...)}})
}

func (b *Transaction) PureBool(value bool) Argument {
	bytes, err := bcs.Marshal(&value)
	return b.pureEncoded(bytes, err)
}

func (b *Transaction) PureU8(value uint8) Argument {
	bytes, err := bcs.Marshal(&value)
	return b.pureEncoded(bytes, err)
}

func (b *Transaction) PureU16(value uint16) Argument {
	bytes, err := bcs.Marshal(&value)
	return b.pureEncoded(bytes, err)
}

func (b *Transaction) PureU32(value uint32) Argument {
	bytes, err := bcs.Marshal(&value)
	return b.pureEncoded(bytes, err)
}

func (b *Transaction) PureU64(value uint64) Argument {
	bytes, err := bcs.Marshal(&value)
	return b.pureEncoded(bytes, err)
}

func (b *Transaction) PureU128(value *big.Int) Argument {
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

func (b *Transaction) PureU256(value *big.Int) Argument {
	if value == nil {
		b.setErr(fmt.Errorf("u256 value is nil"))
		return Argument{}
	}

	bytes, err := utils.EncodeU256(value)
	if err != nil {
		b.setErr(err)
		return Argument{}
	}

	return b.PureBytes(bytes)
}

func (b *Transaction) PureString(value string) Argument {
	bytes, err := bcs.Marshal(&value)
	return b.pureEncoded(bytes, err)
}

func (b *Transaction) PureAddress(value string) Argument {
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

func (b *Transaction) Object(id string) Argument {
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

func (b *Transaction) ObjectRef(ref types.ObjectRef) Argument {
	if b == nil {
		return Argument{}
	}

	return b.addInput(input{Object: &ObjectArg{ImmOrOwnedObject: &ref}})
}

func (b *Transaction) SharedObject(ref types.SharedObjectRef) Argument {
	if b == nil {
		return Argument{}
	}

	return b.addInput(input{Object: &ObjectArg{SharedObject: &ref}})
}

func (b *Transaction) ReceivingObject(ref types.ObjectRef) Argument {
	if b == nil {
		return Argument{}
	}

	return b.addInput(input{Object: &ObjectArg{Receiving: &ref}})
}

func (b *Transaction) SplitCoins(args SplitCoins) []Argument {
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

func (b *Transaction) MergeCoins(args MergeCoins) {
	b.addCommand(Command{MergeCoins: &args})
}

func (b *Transaction) TransferObjects(args TransferObjects) {
	b.addCommand(Command{TransferObjects: &args})
}

func (b *Transaction) MoveCall(args MoveCall) Result {
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

func (b *Transaction) MakeMoveVec(args MakeMoveVecInput) Result {
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

func (b *Transaction) Publish(args PublishInput) Result {
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

func (b *Transaction) Upgrade(args UpgradeInput) Result {
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

func (b *Transaction) Build(ctx context.Context, opts BuildOptions) (BuildResult, error) {
	if b == nil {
		return BuildResult{}, ErrNilTransaction
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

	expiration := b.expiration

	result := BuildResult{
		KindBytes:         kindBytes,
		ProgrammableKind:  programmable,
		ResolvedInputArgs: resolvedInputs,
	}
	if expiration == nil {
		fallback := ExpirationNone()
		expiration = &fallback
	}

	if !b.hasFullTransaction() && opts.GasResolver != nil {
		if err = b.resolveGas(ctx, opts.GasResolver, kind, *expiration); err != nil {
			return BuildResult{}, err
		}
	}

	if !b.hasFullTransaction() {
		return result, nil
	}

	gasData := b.buildGasData()
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
	result.Transaction = &v2.Transaction{Bcs: &v2.Bcs{Name: utils.Ptr("TransactionData"), Value: bytes}}
	return result, nil
}

func (b *Transaction) addInput(in input) Argument {
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

func (b *Transaction) addCommand(cmd Command) *uint16 {
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

type inputUsage struct {
	mutable   bool
	receiving bool
}

func (b *Transaction) resolveInputUsage(ctx context.Context, resolver Resolver) ([]inputUsage, error) {
	usage := make([]inputUsage, len(b.inputs))

	markMutable := func(arg Argument) {
		if arg.Input == nil {
			return
		}
		idx := int(*arg.Input)
		if idx < 0 || idx >= len(usage) {
			return
		}
		if b.inputs[idx].UnresolvedObject == nil {
			return
		}
		usage[idx].mutable = true
	}

	for _, cmd := range b.commands {
		switch {
		case cmd.SplitCoins != nil:
			markMutable(cmd.SplitCoins.Coin)
			for _, amount := range cmd.SplitCoins.Amounts {
				markMutable(amount)
			}
		case cmd.MergeCoins != nil:
			markMutable(cmd.MergeCoins.Destination)
			for _, source := range cmd.MergeCoins.Sources {
				markMutable(source)
			}
		case cmd.TransferObjects != nil:
			for _, obj := range cmd.TransferObjects.Objects {
				markMutable(obj)
			}
		case cmd.MakeMoveVec != nil:
			for _, elem := range cmd.MakeMoveVec.Elements {
				markMutable(elem)
			}
		}
	}

	for _, cmd := range b.commands {
		if cmd.MoveCall == nil {
			continue
		}
		moveCall := cmd.MoveCall
		needsResolution := false
		for _, arg := range moveCall.Arguments {
			if arg.Input == nil {
				continue
			}
			idx := int(*arg.Input)
			if idx < 0 || idx >= len(b.inputs) {
				continue
			}
			if b.inputs[idx].UnresolvedObject != nil {
				needsResolution = true
				break
			}
		}
		if !needsResolution {
			continue
		}

		sig, err := resolver.ResolveMoveFunction(ctx, moveCall.Package.String(), moveCall.Module, moveCall.Function)
		if err != nil {
			return nil, err
		}
		params := trimTxContext(sig.Parameters)
		if len(params) < len(moveCall.Arguments) {
			return nil, fmt.Errorf("move call %s::%s::%s expects %d args, got %d", moveCall.Package.String(), moveCall.Module, moveCall.Function, len(params), len(moveCall.Arguments))
		}

		for i, arg := range moveCall.Arguments {
			if arg.Input == nil {
				continue
			}
			idx := int(*arg.Input)
			if idx < 0 || idx >= len(usage) {
				continue
			}
			if b.inputs[idx].UnresolvedObject == nil {
				continue
			}
			param := params[i]
			if param.Reference != ReferenceImmutable {
				usage[idx].mutable = true
			}
			if isReceivingType(param) {
				usage[idx].receiving = true
			}
		}
	}

	return usage, nil
}

func buildObjectArg(meta ObjectMetadata, usage inputUsage) (*ObjectArg, error) {
	switch meta.OwnerKind {
	case OwnerShared, OwnerConsensusAddress:
		if meta.OwnerVersion == nil {
			return nil, fmt.Errorf("shared object missing initial shared version")
		}
		shared := types.SharedObjectRef{
			ObjectID:             meta.ID,
			InitialSharedVersion: *meta.OwnerVersion,
			Mutable:              usage.mutable,
		}
		return &ObjectArg{SharedObject: &shared}, nil
	case OwnerImmutable, OwnerAddress, OwnerObject, OwnerUnknown:
		if usage.receiving {
			ref := types.ObjectRef{ObjectID: meta.ID, Version: meta.Version, Digest: meta.Digest}
			return &ObjectArg{Receiving: &ref}, nil
		}
		ref := types.ObjectRef{ObjectID: meta.ID, Version: meta.Version, Digest: meta.Digest}
		return &ObjectArg{ImmOrOwnedObject: &ref}, nil
	default:
		return nil, fmt.Errorf("unsupported owner kind %d", meta.OwnerKind)
	}
}

func trimTxContext(params []MoveParameter) []MoveParameter {
	if len(params) == 0 {
		return params
	}
	last := params[len(params)-1]
	if last.TypeName == "0x2::tx_context::TxContext" {
		return params[:len(params)-1]
	}
	return params
}

func isReceivingType(param MoveParameter) bool {
	if param.TypeName == "" {
		return false
	}

	return param.TypeName == "0x2::transfer::Receiving" || strings.HasPrefix(param.TypeName, "0x2::transfer::Receiving<")
}

func (b *Transaction) resolveInputs(ctx context.Context, resolver Resolver) ([]CallArg, error) {
	objectIDs := make([]string, 0)
	resolved := make([]CallArg, len(b.inputs))

	for _, in := range b.inputs {
		if in.UnresolvedObject != nil {
			objectIDs = append(objectIDs, in.UnresolvedObject.ObjectID)
		}
	}

	hasUnresolved := len(objectIDs) > 0
	var usage []inputUsage
	if hasUnresolved {
		if resolver == nil {
			return nil, ErrResolverRequired
		}
		if ctx == nil {
			return nil, fmt.Errorf("nil context")
		}
		resolvedUsage, err := b.resolveInputUsage(ctx, resolver)
		if err != nil {
			return nil, err
		}
		usage = resolvedUsage
	} else {
		usage = make([]inputUsage, len(b.inputs))
	}

	objectMap := map[string]ObjectMetadata{}
	if hasUnresolved {
		unique := utils.UniqueValues(objectIDs)
		refs, err := resolver.ResolveObjects(ctx, unique)
		if err != nil {
			return nil, err
		}
		if len(refs) != len(unique) {
			return nil, fmt.Errorf("resolver returned %d objects for %d object ids", len(refs), len(unique))
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
			if in.Object.SharedObject != nil && usage[i].mutable && !in.Object.SharedObject.Mutable {
				updated := *in.Object.SharedObject
				updated.Mutable = true
				in.Object.SharedObject = &updated
			}
			resolved[i] = CallArg{Object: in.Object}
		case in.UnresolvedObject != nil:
			meta, ok := objectMap[in.UnresolvedObject.ObjectID]
			if !ok {
				return nil, ErrUnresolvedInput
			}
			objArg, err := buildObjectArg(meta, usage[i])
			if err != nil {
				return nil, err
			}
			resolved[i] = CallArg{Object: objArg}
		default:
			return nil, ErrUnresolvedInput
		}
	}

	return resolved, nil
}

func (b *Transaction) resolveGas(ctx context.Context, resolver GasResolver, kind TransactionKind, expiration TransactionExpiration) error {
	if resolver == nil {
		return nil
	}
	if b.sender == nil {
		return nil
	}
	if ctx == nil {
		return fmt.Errorf("nil context")
	}

	if b.gas.Price == nil {
		price, err := resolver.ResolveGasPrice(ctx)
		if err != nil {
			return err
		}
		b.gas.Price = &price
	}

	if b.gas.Budget == nil {
		if b.gas.Price == nil {
			return fmt.Errorf("gas price required to resolve budget")
		}

		owner := b.gas.Owner
		if owner == nil {
			owner = b.sender
		}

		budget, err := resolver.ResolveGasBudget(ctx, GasBudgetInput{
			Sender:     *b.sender,
			GasOwner:   *owner,
			GasPrice:   *b.gas.Price,
			Kind:       kind,
			Expiration: expiration,
		})

		if err != nil {
			return err
		}

		b.gas.Budget = &budget
	}

	if len(b.gas.Payment) == 0 {
		if b.gas.Budget == nil {
			return fmt.Errorf("gas budget required to resolve payment")
		}

		owner := b.gas.Owner
		if owner == nil {
			owner = b.sender
		}

		payment, err := resolver.ResolveGasPayment(ctx, *owner, *b.gas.Budget)
		if err != nil {
			return err
		}
		b.gas.Payment = payment
	}

	return nil
}

func (b *Transaction) hasFullTransaction() bool {
	if b.sender == nil {
		return false
	}

	if b.gas.Budget == nil || b.gas.Price == nil || len(b.gas.Payment) == 0 {
		return false
	}

	return true
}

func (b *Transaction) buildGasData() GasData {
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

func (b *Transaction) pureEncoded(bytes []byte, err error) Argument {
	if b == nil {
		return Argument{}
	}

	if err != nil {
		b.setErr(err)
		return Argument{}
	}

	return b.PureBytes(bytes)
}

func (b *Transaction) setErr(err error) {
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
