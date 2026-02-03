package grpc

import (
	"context"
	"fmt"
	"math"
	"sync"

	v2 "github.com/open-move/sui-go-sdk/proto/sui/rpc/v2"
	"github.com/open-move/sui-go-sdk/transaction"
	"github.com/open-move/sui-go-sdk/types"
	"github.com/open-move/sui-go-sdk/utils"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

const (
	defaultGasCoinType = "0x2::sui::SUI"
	gasBudgetBufferPct = 10
	minGasBudgetBuffer = uint64(1000)
)

type Resolver struct {
	client *Client
	mu     sync.Mutex

	objectCache   map[string]transaction.ObjectMetadata
	functionCache map[string]*transaction.MoveFunction
	packageCache  map[string]*transaction.PackageMetadata
}

// NewResolver returns a resolver backed by the provided gRPC client.
func NewResolver(client *Client) *Resolver {
	return &Resolver{
		client:        client,
		objectCache:   make(map[string]transaction.ObjectMetadata),
		functionCache: make(map[string]*transaction.MoveFunction),
		packageCache:  make(map[string]*transaction.PackageMetadata),
	}
}

// ResolveObjects resolves object IDs into metadata using the gRPC client.
func (r *Resolver) ResolveObjects(ctx context.Context, objectIDs []string) ([]transaction.ObjectMetadata, error) {
	if r == nil || r.client == nil {
		return nil, fmt.Errorf("nil client")
	}
	if len(objectIDs) == 0 {
		return nil, nil
	}

	normalized := make([]string, len(objectIDs))
	for i, id := range objectIDs {
		value, err := utils.NormalizeAddress(id)
		if err != nil {
			return nil, err
		}
		normalized[i] = value
	}

	pending := make([]string, 0)
	indexByID := make(map[string][]int)
	results := make([]transaction.ObjectMetadata, len(objectIDs))

	for i, id := range normalized {
		indexByID[id] = append(indexByID[id], i)
	}

	r.mu.Lock()
	for id, indices := range indexByID {
		if meta, ok := r.objectCache[id]; ok {
			for _, idx := range indices {
				results[idx] = meta
			}
			continue
		}
		pending = append(pending, id)
	}
	r.mu.Unlock()

	if len(pending) > 0 {
		requests := make([]ObjectRequest, len(pending))
		for i, id := range pending {
			requests[i] = ObjectRequest{ObjectID: id}
		}
		mask := &fieldmaskpb.FieldMask{Paths: []string{"object_id", "version", "digest", "owner"}}
		responses, err := r.client.BatchGetObjects(ctx, requests, mask)
		if err != nil {
			return nil, err
		}
		if len(responses) != len(pending) {
			return nil, fmt.Errorf("resolver returned %d objects for %d ids", len(responses), len(pending))
		}
		r.mu.Lock()
		for i, resp := range responses {
			if resp.Err != nil {
				r.mu.Unlock()
				return nil, fmt.Errorf("resolve object %s: %w", pending[i], resp.Err)
			}
			meta, err := objectMetadataFromObject(resp.Object)
			if err != nil {
				r.mu.Unlock()
				return nil, fmt.Errorf("resolve object %s: %w", pending[i], err)
			}
			r.objectCache[pending[i]] = meta
			for _, idx := range indexByID[pending[i]] {
				results[idx] = meta
			}
		}
		r.mu.Unlock()
	}

	return results, nil
}

// ResolveMoveFunction fetches Move function metadata for the requested target.
func (r *Resolver) ResolveMoveFunction(ctx context.Context, packageID, module, function string) (*transaction.MoveFunction, error) {
	if r == nil || r.client == nil {
		return nil, fmt.Errorf("nil client")
	}
	key := fmt.Sprintf("%s::%s::%s", packageID, module, function)
	if key == "::" {
		return nil, fmt.Errorf("invalid move function target")
	}

	r.mu.Lock()
	if cached, ok := r.functionCache[key]; ok {
		r.mu.Unlock()
		return cached, nil
	}
	r.mu.Unlock()

	req := &v2.GetFunctionRequest{PackageId: &packageID, ModuleName: &module, Name: &function}
	resp, err := r.client.movePackageClient.GetFunction(ctx, req)
	if err != nil {
		return nil, err
	}
	fn := resp.GetFunction()
	if fn == nil {
		return nil, fmt.Errorf("move function not found")
	}
	converted := convertMoveFunction(fn)

	r.mu.Lock()
	r.functionCache[key] = converted
	r.mu.Unlock()

	return converted, nil
}

// ResolvePackage fetches Move package metadata for the provided package ID.
func (r *Resolver) ResolvePackage(ctx context.Context, packageID string) (*transaction.PackageMetadata, error) {
	if r == nil || r.client == nil {
		return nil, fmt.Errorf("nil client")
	}
	if packageID == "" {
		return nil, fmt.Errorf("empty package id")
	}

	r.mu.Lock()
	if cached, ok := r.packageCache[packageID]; ok {
		r.mu.Unlock()
		return cached, nil
	}
	r.mu.Unlock()

	req := &v2.GetPackageRequest{PackageId: &packageID}
	resp, err := r.client.movePackageClient.GetPackage(ctx, req)
	if err != nil {
		return nil, err
	}
	pkg := resp.GetPackage()
	if pkg == nil {
		return nil, fmt.Errorf("package not found")
	}

	meta := &transaction.PackageMetadata{
		StorageID:  pkg.GetStorageId(),
		OriginalID: pkg.GetOriginalId(),
		Version:    pkg.GetVersion(),
	}

	r.mu.Lock()
	r.packageCache[packageID] = meta
	r.mu.Unlock()

	return meta, nil
}

// ResolveGasPrice fetches the current reference gas price.
func (r *Resolver) ResolveGasPrice(ctx context.Context) (uint64, error) {
	if r == nil || r.client == nil {
		return 0, fmt.Errorf("nil client")
	}
	if ctx == nil {
		return 0, fmt.Errorf("nil context")
	}

	return r.client.ReferenceGasPrice(ctx)
}

// ResolveGasBudget estimates a gas budget using simulation.
func (r *Resolver) ResolveGasBudget(ctx context.Context, input transaction.GasBudgetInput) (uint64, error) {
	if r == nil || r.client == nil {
		return 0, fmt.Errorf("nil client")
	}
	if ctx == nil {
		return 0, fmt.Errorf("nil context")
	}

	tx, err := buildGasEstimationTransaction(input)
	if err != nil {
		return 0, err
	}

	checks := v2.SimulateTransactionRequest_ENABLED
	doSelect := true
	resp, err := r.client.SimulateTransaction(ctx, tx, &SimulateTransactionOptions{
		ReadMask: &fieldmaskpb.FieldMask{Paths: []string{
			"transaction.effects.status",
			"transaction.effects.gas_used",
		}},
		Checks:         &checks,
		DoGasSelection: &doSelect,
	})

	if err != nil {
		return 0, err
	}

	executed := resp.GetTransaction()
	if executed == nil {
		return 0, fmt.Errorf("simulate transaction response missing transaction")
	}
	effects := executed.GetEffects()
	if effects == nil {
		return 0, fmt.Errorf("simulate transaction response missing effects")
	}
	status := effects.GetStatus()
	if status != nil && !status.GetSuccess() {
		description := "unknown execution error"
		if errDetails := status.GetError(); errDetails != nil {
			if msg := errDetails.GetDescription(); msg != "" {
				description = msg
			}
		}
		return 0, fmt.Errorf("simulate transaction failed: %s", description)
	}
	gasUsed := effects.GetGasUsed()
	if gasUsed == nil {
		return 0, fmt.Errorf("simulate transaction response missing gas usage")
	}

	base := gasUsed.GetComputationCost()
	if cost := gasUsed.GetStorageCost(); cost > 0 {
		if base > math.MaxUint64-cost {
			base = math.MaxUint64
		} else {
			base += cost
		}
	}
	rebate := gasUsed.GetStorageRebate()
	if rebate > base {
		base = 0
	} else {
		base -= rebate
	}
	if fee := gasUsed.GetNonRefundableStorageFee(); fee > 0 {
		if base > math.MaxUint64-fee {
			base = math.MaxUint64
		} else {
			base += fee
		}
	}

	return addGasBudgetBuffer(base), nil
}

// ResolveGasPayment selects gas payment objects for the given budget.
func (r *Resolver) ResolveGasPayment(ctx context.Context, owner types.Address, budget uint64) ([]types.ObjectRef, error) {
	if r == nil || r.client == nil {
		return nil, fmt.Errorf("nil client")
	}
	if ctx == nil {
		return nil, fmt.Errorf("nil context")
	}
	if budget == 0 {
		return nil, fmt.Errorf("gas budget must be greater than zero")
	}

	coins, err := r.client.SelectCoins(ctx, owner.String(), defaultGasCoinType, budget)
	if err != nil {
		return nil, err
	}

	refs := make([]types.ObjectRef, len(coins))
	for i, coin := range coins {
		ref, err := utils.ParseObjectRef(coin.GetObjectId(), coin.GetVersion(), coin.GetDigest())
		if err != nil {
			return nil, err
		}
		refs[i] = ref
	}

	return refs, nil
}

func addGasBudgetBuffer(base uint64) uint64 {
	buffer := base / gasBudgetBufferPct
	if buffer < minGasBudgetBuffer {
		buffer = minGasBudgetBuffer
	}
	if base > math.MaxUint64-buffer {
		return math.MaxUint64
	}
	return base + buffer
}

func buildGasEstimationTransaction(input transaction.GasBudgetInput) (*v2.Transaction, error) {
	kind, err := buildProtoTransactionKind(input.Kind)
	if err != nil {
		return nil, err
	}

	sender := input.Sender.String()
	expiration, err := buildProtoTransactionExpiration(input.Expiration)
	if err != nil {
		return nil, err
	}

	gasOwner := input.GasOwner
	if gasOwner == (types.Address{}) {
		gasOwner = input.Sender
	}
	owner := gasOwner.String()
	price := input.GasPrice

	var gasPayment *v2.GasPayment
	if owner != "" || price != 0 {
		gasPayment = &v2.GasPayment{}
		if owner != "" {
			gasPayment.Owner = &owner
		}
		if price != 0 {
			gasPayment.Price = &price
		}
	}

	return &v2.Transaction{
		Kind:       kind,
		Sender:     &sender,
		GasPayment: gasPayment,
		Expiration: expiration,
	}, nil
}

func buildProtoTransactionExpiration(expiration transaction.TransactionExpiration) (*v2.TransactionExpiration, error) {
	switch {
	case expiration.None != nil:
		kind := v2.TransactionExpiration_NONE
		return &v2.TransactionExpiration{Kind: kind.Enum()}, nil
	case expiration.Epoch != nil:
		kind := v2.TransactionExpiration_EPOCH
		epoch := *expiration.Epoch
		return &v2.TransactionExpiration{Kind: kind.Enum(), Epoch: &epoch}, nil
	default:
		return nil, fmt.Errorf("unsupported transaction expiration")
	}
}

func buildProtoTransactionKind(kind transaction.TransactionKind) (*v2.TransactionKind, error) {
	if kind.ProgrammableTransaction == nil {
		return nil, fmt.Errorf("unsupported transaction kind")
	}

	programmable, err := buildProtoProgrammableTransaction(kind.ProgrammableTransaction)
	if err != nil {
		return nil, err
	}

	protoKind := v2.TransactionKind_PROGRAMMABLE_TRANSACTION
	return &v2.TransactionKind{
		Kind: protoKind.Enum(),
		Data: &v2.TransactionKind_ProgrammableTransaction{ProgrammableTransaction: programmable},
	}, nil
}

func buildProtoProgrammableTransaction(tx *transaction.ProgrammableTransaction) (*v2.ProgrammableTransaction, error) {
	if tx == nil {
		return nil, fmt.Errorf("nil programmable transaction")
	}
	inputs := make([]*v2.Input, len(tx.Inputs))
	for i, input := range tx.Inputs {
		protoInput, err := buildProtoInput(input)
		if err != nil {
			return nil, err
		}
		inputs[i] = protoInput
	}

	commands := make([]*v2.Command, len(tx.Commands))
	for i, cmd := range tx.Commands {
		protoCmd, err := buildProtoCommand(cmd)
		if err != nil {
			return nil, err
		}
		commands[i] = protoCmd
	}

	return &v2.ProgrammableTransaction{Inputs: inputs, Commands: commands}, nil
}

func buildProtoInput(arg transaction.CallArg) (*v2.Input, error) {
	switch {
	case arg.Pure != nil:
		kind := v2.Input_PURE
		pure := append([]byte(nil), arg.Pure.Bytes...)
		return &v2.Input{Kind: kind.Enum(), Pure: pure}, nil
	case arg.Object != nil:
		obj := arg.Object
		switch {
		case obj.ImmOrOwnedObject != nil:
			kind := v2.Input_IMMUTABLE_OR_OWNED
			ref := obj.ImmOrOwnedObject
			id := ref.ObjectID.String()
			version := ref.Version
			digest := ref.Digest.String()
			return &v2.Input{Kind: kind.Enum(), ObjectId: &id, Version: &version, Digest: &digest}, nil
		case obj.SharedObject != nil:
			kind := v2.Input_SHARED
			ref := obj.SharedObject
			id := ref.ObjectID.String()
			version := ref.InitialSharedVersion
			mutable := ref.Mutable
			return &v2.Input{Kind: kind.Enum(), ObjectId: &id, Version: &version, Mutable: &mutable}, nil
		case obj.Receiving != nil:
			kind := v2.Input_RECEIVING
			ref := obj.Receiving
			id := ref.ObjectID.String()
			version := ref.Version
			digest := ref.Digest.String()
			return &v2.Input{Kind: kind.Enum(), ObjectId: &id, Version: &version, Digest: &digest}, nil
		default:
			return nil, fmt.Errorf("unsupported object argument")
		}
	default:
		return nil, fmt.Errorf("unsupported input argument")
	}
}

func buildProtoCommand(cmd transaction.Command) (*v2.Command, error) {
	switch {
	case cmd.MoveCall != nil:
		protoCall, err := buildProtoMoveCall(*cmd.MoveCall)
		if err != nil {
			return nil, err
		}
		return &v2.Command{Command: &v2.Command_MoveCall{MoveCall: protoCall}}, nil
	case cmd.TransferObjects != nil:
		protoCmd, err := buildProtoTransferObjects(*cmd.TransferObjects)
		if err != nil {
			return nil, err
		}
		return &v2.Command{Command: &v2.Command_TransferObjects{TransferObjects: protoCmd}}, nil
	case cmd.SplitCoins != nil:
		protoCmd, err := buildProtoSplitCoins(*cmd.SplitCoins)
		if err != nil {
			return nil, err
		}
		return &v2.Command{Command: &v2.Command_SplitCoins{SplitCoins: protoCmd}}, nil
	case cmd.MergeCoins != nil:
		protoCmd, err := buildProtoMergeCoins(*cmd.MergeCoins)
		if err != nil {
			return nil, err
		}
		return &v2.Command{Command: &v2.Command_MergeCoins{MergeCoins: protoCmd}}, nil
	case cmd.Publish != nil:
		protoCmd, err := buildProtoPublish(*cmd.Publish)
		if err != nil {
			return nil, err
		}
		return &v2.Command{Command: &v2.Command_Publish{Publish: protoCmd}}, nil
	case cmd.MakeMoveVec != nil:
		protoCmd, err := buildProtoMakeMoveVector(*cmd.MakeMoveVec)
		if err != nil {
			return nil, err
		}
		return &v2.Command{Command: &v2.Command_MakeMoveVector{MakeMoveVector: protoCmd}}, nil
	case cmd.Upgrade != nil:
		protoCmd, err := buildProtoUpgrade(*cmd.Upgrade)
		if err != nil {
			return nil, err
		}
		return &v2.Command{Command: &v2.Command_Upgrade{Upgrade: protoCmd}}, nil
	default:
		return nil, fmt.Errorf("unsupported command")
	}
}

func buildProtoMoveCall(call transaction.ProgrammableMoveCall) (*v2.MoveCall, error) {
	args := make([]*v2.Argument, len(call.Arguments))
	for i, arg := range call.Arguments {
		protoArg, err := buildProtoArgument(arg)
		if err != nil {
			return nil, err
		}
		args[i] = protoArg
	}

	typeArgs := make([]string, len(call.TypeArguments))
	for i, tag := range call.TypeArguments {
		typeArgs[i] = tag.String()
	}

	pkg := call.Package.String()
	module := call.Module
	function := call.Function

	return &v2.MoveCall{
		Package:       &pkg,
		Module:        &module,
		Function:      &function,
		TypeArguments: typeArgs,
		Arguments:     args,
	}, nil
}

func buildProtoTransferObjects(cmd transaction.TransferObjects) (*v2.TransferObjects, error) {
	objects := make([]*v2.Argument, len(cmd.Objects))
	for i, arg := range cmd.Objects {
		protoArg, err := buildProtoArgument(arg)
		if err != nil {
			return nil, err
		}
		objects[i] = protoArg
	}
	address, err := buildProtoArgument(cmd.Address)
	if err != nil {
		return nil, err
	}

	return &v2.TransferObjects{Objects: objects, Address: address}, nil
}

func buildProtoSplitCoins(cmd transaction.SplitCoins) (*v2.SplitCoins, error) {
	coin, err := buildProtoArgument(cmd.Coin)
	if err != nil {
		return nil, err
	}

	amounts := make([]*v2.Argument, len(cmd.Amounts))
	for i, amount := range cmd.Amounts {
		protoArg, err := buildProtoArgument(amount)
		if err != nil {
			return nil, err
		}
		amounts[i] = protoArg
	}

	return &v2.SplitCoins{Coin: coin, Amounts: amounts}, nil
}

func buildProtoMergeCoins(cmd transaction.MergeCoins) (*v2.MergeCoins, error) {
	destination, err := buildProtoArgument(cmd.Destination)
	if err != nil {
		return nil, err
	}

	sources := make([]*v2.Argument, len(cmd.Sources))
	for i, source := range cmd.Sources {
		protoArg, err := buildProtoArgument(source)
		if err != nil {
			return nil, err
		}
		sources[i] = protoArg
	}

	return &v2.MergeCoins{Coin: destination, CoinsToMerge: sources}, nil
}

func buildProtoPublish(cmd transaction.Publish) (*v2.Publish, error) {
	deps := make([]string, len(cmd.Dependencies))
	for i, dep := range cmd.Dependencies {
		deps[i] = dep.String()
	}

	return &v2.Publish{Modules: cmd.Modules, Dependencies: deps}, nil
}

func buildProtoMakeMoveVector(cmd transaction.MakeMoveVec) (*v2.MakeMoveVector, error) {
	elements := make([]*v2.Argument, len(cmd.Elements))
	for i, elem := range cmd.Elements {
		protoArg, err := buildProtoArgument(elem)
		if err != nil {
			return nil, err
		}
		elements[i] = protoArg
	}

	var elementType *string
	if !cmd.Type.None {
		value := cmd.Type.Some.String()
		elementType = &value
	}

	return &v2.MakeMoveVector{ElementType: elementType, Elements: elements}, nil
}

func buildProtoUpgrade(cmd transaction.Upgrade) (*v2.Upgrade, error) {
	deps := make([]string, len(cmd.Dependencies))
	for i, dep := range cmd.Dependencies {
		deps[i] = dep.String()
	}

	pkg := cmd.Package.String()
	ticket, err := buildProtoArgument(cmd.Ticket)
	if err != nil {
		return nil, err
	}

	return &v2.Upgrade{Modules: cmd.Modules, Dependencies: deps, Package: &pkg, Ticket: ticket}, nil
}

func buildProtoArgument(arg transaction.Argument) (*v2.Argument, error) {
	switch {
	case arg.GasCoin != nil:
		kind := v2.Argument_GAS
		return &v2.Argument{Kind: kind.Enum()}, nil
	case arg.Input != nil:
		kind := v2.Argument_INPUT
		idx := uint32(*arg.Input)
		return &v2.Argument{Kind: kind.Enum(), Input: &idx}, nil
	case arg.Result != nil:
		kind := v2.Argument_RESULT
		idx := uint32(*arg.Result)
		return &v2.Argument{Kind: kind.Enum(), Result: &idx}, nil
	case arg.NestedResult != nil:
		kind := v2.Argument_RESULT
		idx := uint32(arg.NestedResult.Index)
		sub := uint32(arg.NestedResult.ResultIndex)
		return &v2.Argument{Kind: kind.Enum(), Result: &idx, Subresult: &sub}, nil
	default:
		return nil, fmt.Errorf("unsupported argument")
	}
}

func objectMetadataFromObject(obj *v2.Object) (transaction.ObjectMetadata, error) {
	if obj == nil {
		return transaction.ObjectMetadata{}, fmt.Errorf("nil object")
	}
	id := obj.GetObjectId()
	if id == "" {
		return transaction.ObjectMetadata{}, fmt.Errorf("object id missing")
	}
	addr, err := utils.ParseAddress(id)
	if err != nil {
		return transaction.ObjectMetadata{}, err
	}
	version := obj.GetVersion()
	digestStr := obj.GetDigest()
	if digestStr == "" {
		return transaction.ObjectMetadata{}, fmt.Errorf("object digest missing")
	}
	digest, err := utils.ParseDigest(digestStr)
	if err != nil {
		return transaction.ObjectMetadata{}, err
	}

	ownerKind := transaction.OwnerUnknown
	var ownerVersion *uint64
	owner := obj.GetOwner()
	if owner != nil {
		switch owner.GetKind() {
		case v2.Owner_ADDRESS:
			ownerKind = transaction.OwnerAddress
		case v2.Owner_OBJECT:
			ownerKind = transaction.OwnerObject
		case v2.Owner_SHARED:
			ownerKind = transaction.OwnerShared
		case v2.Owner_IMMUTABLE:
			ownerKind = transaction.OwnerImmutable
		case v2.Owner_CONSENSUS_ADDRESS:
			ownerKind = transaction.OwnerConsensusAddress
		default:
			ownerKind = transaction.OwnerUnknown
		}
		if owner.GetVersion() != 0 {
			v := owner.GetVersion()
			ownerVersion = &v
		}
	}

	return transaction.ObjectMetadata{
		ID:           types.ObjectID(addr),
		Version:      version,
		Digest:       digest,
		OwnerKind:    ownerKind,
		OwnerVersion: ownerVersion,
	}, nil
}

func convertMoveFunction(fn *v2.FunctionDescriptor) *transaction.MoveFunction {
	params := make([]transaction.MoveParameter, len(fn.GetParameters()))
	for i, param := range fn.GetParameters() {
		params[i] = convertMoveParameter(param)
	}
	return &transaction.MoveFunction{Parameters: params}
}

func convertMoveParameter(param *v2.OpenSignature) transaction.MoveParameter {
	ref := transaction.ReferenceUnknown
	if param != nil {
		switch param.GetReference() {
		case v2.OpenSignature_IMMUTABLE:
			ref = transaction.ReferenceImmutable
		case v2.OpenSignature_MUTABLE:
			ref = transaction.ReferenceMutable
		default:
			ref = transaction.ReferenceUnknown
		}
	}

	var typeName string
	if param != nil && param.Body != nil && param.Body.GetType() == v2.OpenSignatureBody_DATATYPE {
		typeName = param.Body.GetTypeName()
	}

	return transaction.MoveParameter{
		Reference: ref,
		TypeName:  typeName,
	}
}
