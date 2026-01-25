package grpc

import (
	"context"
	"fmt"
	"sync"

	v2 "github.com/open-move/sui-go-sdk/proto/sui/rpc/v2"
	"github.com/open-move/sui-go-sdk/transaction"
	"github.com/open-move/sui-go-sdk/types"
	"github.com/open-move/sui-go-sdk/utils"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

type Resolver struct {
	client *Client
	mu     sync.Mutex

	objectCache   map[string]transaction.ObjectMetadata
	functionCache map[string]*transaction.MoveFunction
	packageCache  map[string]*transaction.PackageMetadata
}

func NewResolver(client *Client) *Resolver {
	return &Resolver{
		client:        client,
		objectCache:   make(map[string]transaction.ObjectMetadata),
		functionCache: make(map[string]*transaction.MoveFunction),
		packageCache:  make(map[string]*transaction.PackageMetadata),
	}
}

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
	resp, err := r.client.MovePackageClient().GetFunction(ctx, req)
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
	resp, err := r.client.MovePackageClient().GetPackage(ctx, req)
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
