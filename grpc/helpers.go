package grpc

import (
	"context"
	"errors"
	"fmt"
	"strings"

	v2 "github.com/open-move/sui-go-sdk/proto/sui/rpc/v2"
	"google.golang.org/grpc"
	grpcstatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// GetObjectOptions customises the behaviour of GetObject.
type GetObjectOptions struct {
	Version  *uint64
	ReadMask *fieldmaskpb.FieldMask
}

// GetObject fetches a single object by ID, optionally specifying a version and field mask.
func (c *GRPCClient) GetObject(ctx context.Context, objectID string, options *GetObjectOptions, opts ...grpc.CallOption) (*v2.Object, error) {
	if c == nil {
		return nil, errors.New("nil client")
	}
	if ctx == nil {
		return nil, errors.New("nil context")
	}
	if objectID == "" {
		return nil, errors.New("object ID is empty")
	}

	req := &v2.GetObjectRequest{ObjectId: stringPtr(objectID)}
	if options != nil {
		if options.Version != nil {
			version := *options.Version
			req.Version = &version
		}
		if options.ReadMask != nil {
			req.ReadMask = cloneFieldMask(options.ReadMask)
		}
	}

	resp, err := c.LedgerClient().GetObject(ctx, req, opts...)
	if err != nil {
		return nil, err
	}
	obj := resp.GetObject()
	if obj == nil {
		return nil, fmt.Errorf("object %q not found", objectID)
	}
	return obj, nil
}

// GetTransactionOptions customises the behaviour of GetTransaction.
type GetTransactionOptions struct {
	ReadMask *fieldmaskpb.FieldMask
}

// GetTransaction fetches an executed transaction by digest.
func (c *GRPCClient) GetTransaction(ctx context.Context, digest string, options *GetTransactionOptions, opts ...grpc.CallOption) (*v2.ExecutedTransaction, error) {
	if c == nil {
		return nil, errors.New("nil client")
	}
	if ctx == nil {
		return nil, errors.New("nil context")
	}
	if digest == "" {
		return nil, errors.New("transaction digest is empty")
	}

	req := &v2.GetTransactionRequest{Digest: stringPtr(digest)}
	if options != nil && options.ReadMask != nil {
		req.ReadMask = cloneFieldMask(options.ReadMask)
	}

	resp, err := c.LedgerClient().GetTransaction(ctx, req, opts...)
	if err != nil {
		return nil, err
	}
	tx := resp.GetTransaction()
	if tx == nil {
		return nil, fmt.Errorf("transaction %q not found", digest)
	}
	return tx, nil
}

// GetCheckpointBySequence fetches a checkpoint by its sequence number.
func (c *GRPCClient) GetCheckpointBySequence(ctx context.Context, sequence uint64, readMask *fieldmaskpb.FieldMask, opts ...grpc.CallOption) (*v2.Checkpoint, error) {
	if c == nil {
		return nil, errors.New("nil client")
	}
	if ctx == nil {
		return nil, errors.New("nil context")
	}

	req := &v2.GetCheckpointRequest{
		CheckpointId: &v2.GetCheckpointRequest_SequenceNumber{SequenceNumber: sequence},
	}
	if readMask != nil {
		req.ReadMask = cloneFieldMask(readMask)
	}

	resp, err := c.LedgerClient().GetCheckpoint(ctx, req, opts...)
	if err != nil {
		return nil, err
	}
	checkpoint := resp.GetCheckpoint()
	if checkpoint == nil {
		return nil, fmt.Errorf("checkpoint %d not found", sequence)
	}
	return checkpoint, nil
}

// GetCheckpointByDigest fetches a checkpoint by its digest.
func (c *GRPCClient) GetCheckpointByDigest(ctx context.Context, digest string, readMask *fieldmaskpb.FieldMask, opts ...grpc.CallOption) (*v2.Checkpoint, error) {
	if c == nil {
		return nil, errors.New("nil client")
	}
	if ctx == nil {
		return nil, errors.New("nil context")
	}
	if digest == "" {
		return nil, errors.New("checkpoint digest is empty")
	}

	req := &v2.GetCheckpointRequest{
		CheckpointId: &v2.GetCheckpointRequest_Digest{Digest: digest},
	}
	if readMask != nil {
		req.ReadMask = cloneFieldMask(readMask)
	}

	resp, err := c.LedgerClient().GetCheckpoint(ctx, req, opts...)
	if err != nil {
		return nil, err
	}
	checkpoint := resp.GetCheckpoint()
	if checkpoint == nil {
		return nil, fmt.Errorf("checkpoint %q not found", digest)
	}
	return checkpoint, nil
}

// GetCurrentEpoch fetches information about the current epoch, optionally restricting the response with a field mask.
func (c *GRPCClient) GetCurrentEpoch(ctx context.Context, readMask *fieldmaskpb.FieldMask, opts ...grpc.CallOption) (*v2.Epoch, error) {
	if c == nil {
		return nil, errors.New("nil client")
	}
	if ctx == nil {
		return nil, errors.New("nil context")
	}

	req := &v2.GetEpochRequest{}
	if readMask != nil {
		req.ReadMask = cloneFieldMask(readMask)
	}

	resp, err := c.LedgerClient().GetEpoch(ctx, req, opts...)
	if err != nil {
		return nil, err
	}
	epoch := resp.GetEpoch()
	if epoch == nil {
		return nil, errors.New("epoch response missing epoch data")
	}
	return epoch, nil
}

// ReferenceGasPrice returns the reference gas price from the current epoch.
func (c *GRPCClient) ReferenceGasPrice(ctx context.Context, opts ...grpc.CallOption) (uint64, error) {
	mask := &fieldmaskpb.FieldMask{Paths: []string{"reference_gas_price"}}
	epoch, err := c.GetCurrentEpoch(ctx, mask, opts...)
	if err != nil {
		return 0, err
	}
	return epoch.GetReferenceGasPrice(), nil
}

// ObjectRequest describes a single object fetch to include in BatchGetObjects.
type ObjectRequest struct {
	ObjectID string
	Version  *uint64
}

// ObjectResult contains either an object or an error for an entry returned by BatchGetObjects.
type ObjectResult struct {
	Object *v2.Object
	Err    error
}

// BatchGetObjects issues a BatchGetObjects RPC and maps the response to the provided requests.
func (c *GRPCClient) BatchGetObjects(ctx context.Context, requests []ObjectRequest, readMask *fieldmaskpb.FieldMask, opts ...grpc.CallOption) ([]ObjectResult, error) {
	if c == nil {
		return nil, errors.New("nil client")
	}
	if ctx == nil {
		return nil, errors.New("nil context")
	}
	if len(requests) == 0 {
		return nil, errors.New("no object requests provided")
	}

	batch := &v2.BatchGetObjectsRequest{Requests: make([]*v2.GetObjectRequest, 0, len(requests))}
	if readMask != nil {
		batch.ReadMask = cloneFieldMask(readMask)
	}

	for i, req := range requests {
		if strings.TrimSpace(req.ObjectID) == "" {
			return nil, fmt.Errorf("request %d has empty object ID", i)
		}
		objReq := &v2.GetObjectRequest{ObjectId: stringPtr(req.ObjectID)}
		if req.Version != nil {
			version := *req.Version
			objReq.Version = &version
		}
		batch.Requests = append(batch.Requests, objReq)
	}

	resp, err := c.LedgerClient().BatchGetObjects(ctx, batch, opts...)
	if err != nil {
		return nil, err
	}

	objects := resp.GetObjects()
	if len(objects) != len(requests) {
		return nil, fmt.Errorf("unexpected response length %d (expected %d)", len(objects), len(requests))
	}

	results := make([]ObjectResult, len(objects))
	for i, res := range objects {
		if res == nil {
			results[i] = ObjectResult{Err: errors.New("empty object result")}
			continue
		}
		if obj := res.GetObject(); obj != nil {
			results[i] = ObjectResult{Object: obj}
			continue
		}
		if errStatus := res.GetError(); errStatus != nil {
			results[i] = ObjectResult{Err: grpcstatus.ErrorProto(errStatus)}
			continue
		}
		results[i] = ObjectResult{Err: errors.New("object result missing data")}
	}

	return results, nil
}

func stringPtr(s string) *string {
	return &s
}

func cloneFieldMask(mask *fieldmaskpb.FieldMask) *fieldmaskpb.FieldMask {
	if mask == nil {
		return nil
	}
	cloned := proto.Clone(mask)
	if cloned == nil {
		return nil
	}
	return cloned.(*fieldmaskpb.FieldMask)
}

func ensureFieldMaskPaths(mask *fieldmaskpb.FieldMask, paths ...string) *fieldmaskpb.FieldMask {
	var out *fieldmaskpb.FieldMask
	if mask != nil {
		out = cloneFieldMask(mask)
	} else {
		out = &fieldmaskpb.FieldMask{}
	}

	existing := make(map[string]struct{}, len(out.Paths))
	filtered := out.Paths[:0]
	for _, p := range out.Paths {
		if p == "" {
			continue
		}
		if _, ok := existing[p]; ok {
			continue
		}
		existing[p] = struct{}{}
		filtered = append(filtered, p)
	}
	out.Paths = filtered

	for _, p := range paths {
		if p == "" {
			continue
		}
		if _, ok := existing[p]; ok {
			continue
		}
		out.Paths = append(out.Paths, p)
		existing[p] = struct{}{}
	}

	if len(out.Paths) == 0 {
		return nil
	}
	return out
}
