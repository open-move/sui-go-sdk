package transaction

import (
	"context"
	"fmt"

	suiGrpc "github.com/open-move/sui-go-sdk/grpc"
	v2 "github.com/open-move/sui-go-sdk/proto/sui/rpc/v2"
	"github.com/open-move/sui-go-sdk/types"
	"github.com/open-move/sui-go-sdk/utils"
)

type GRPCResolver struct {
	client *suiGrpc.Client
}

type ObjectResolver interface {
	ResolveObjects(ctx context.Context, objectIDs []string) ([]types.ObjectRef, error)
}

type ObjectResolverFunc func(ctx context.Context, objectIDs []string) ([]types.ObjectRef, error)

func (f ObjectResolverFunc) ResolveObjects(ctx context.Context, objectIDs []string) ([]types.ObjectRef, error) {
	return f(ctx, objectIDs)
}

func ResolverFromGRPC(client *suiGrpc.Client) ObjectResolver {
	return &GRPCResolver{client: client}
}

func (r *GRPCResolver) ResolveObjects(ctx context.Context, objectIDs []string) ([]types.ObjectRef, error) {
	if r == nil || r.client == nil {
		return nil, fmt.Errorf("nil grpc client")
	}

	if len(objectIDs) == 0 {
		return nil, nil
	}

	requests := make([]suiGrpc.ObjectRequest, len(objectIDs))
	for i, id := range objectIDs {
		requests[i] = suiGrpc.ObjectRequest{ObjectID: id}
	}

	results, err := r.client.BatchGetObjects(ctx, requests, nil)
	if err != nil {
		return nil, err
	}

	refs := make([]types.ObjectRef, len(results))
	for i, result := range results {
		if result.Err != nil {
			return nil, fmt.Errorf("resolve object %s: %w", objectIDs[i], result.Err)
		}

		if result.Object == nil {
			return nil, fmt.Errorf("resolve object %s: missing object", objectIDs[i])
		}

		ref, err := ObjectRefFromObject(result.Object)
		if err != nil {
			return nil, fmt.Errorf("resolve object %s: %w", objectIDs[i], err)
		}

		refs[i] = ref
	}

	return refs, nil
}

func ObjectRefFromObject(obj *v2.Object) (types.ObjectRef, error) {
	if obj == nil {
		return types.ObjectRef{}, fmt.Errorf("nil object")
	}

	id := obj.GetObjectId()
	if id == "" {
		return types.ObjectRef{}, ErrInvalidAddress
	}

	version := obj.GetVersion()
	digest := obj.GetDigest()
	if digest == "" {
		return types.ObjectRef{}, ErrInvalidDigest
	}

	return utils.ParseObjectRef(id, version, digest)
}
