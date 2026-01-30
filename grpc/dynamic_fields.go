package grpc

import (
	"context"
	"fmt"

	v2 "github.com/open-move/sui-go-sdk/proto/sui/rpc/v2"
	"github.com/open-move/sui-go-sdk/utils"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// ListDynamicFieldsOptions configures the ListDynamicFields request.
type ListDynamicFieldsOptions struct {
	PageSize  uint32
	PageToken []byte
	ReadMask  *fieldmaskpb.FieldMask
}

// ListDynamicFields returns the dynamic fields associated with the parent object.
func (c *GRPCClient) ListDynamicFields(
	ctx context.Context,
	parentID string,
	options *ListDynamicFieldsOptions,
	opts ...grpc.CallOption,
) (*v2.ListDynamicFieldsResponse, error) {
	if c == nil {
		return nil, fmt.Errorf("nil client")
	}
	if ctx == nil {
		return nil, fmt.Errorf("nil context")
	}

	normalized, err := utils.NormalizeAddress(parentID)
	if err != nil {
		return nil, err
	}

	req := &v2.ListDynamicFieldsRequest{Parent: utils.StringPtr(normalized)}
	if options != nil {
		if options.PageSize != 0 {
			size := options.PageSize
			req.PageSize = &size
		}
		if len(options.PageToken) != 0 {
			req.PageToken = append([]byte(nil), options.PageToken...)
		}
		if options.ReadMask != nil {
			req.ReadMask = options.ReadMask
		}
	}

	return c.StateClient().ListDynamicFields(ctx, req, opts...)
}

func (c *GRPCClient) GetDynamicFieldObject(
	ctx context.Context,
	parentID string,
	nameType string,
	nameBcs []byte,
	options *GetObjectOptions,
	opts ...grpc.CallOption,
) (*v2.Object, error) {
	if c == nil {
		return nil, fmt.Errorf("nil client")
	}
	fieldID, err := utils.DeriveDynamicFieldID(parentID, nameType, nameBcs)
	if err != nil {
		return nil, err
	}
	return c.GetObject(ctx, fieldID, options, opts...)
}

// GetDerivedObject returns the object derived from the parent object and key.
func (c *GRPCClient) GetDerivedObject(
	ctx context.Context,
	parentID string,
	typeTag string,
	key []byte,
	options *GetObjectOptions,
	opts ...grpc.CallOption,
) (*v2.Object, error) {
	if c == nil {
		return nil, fmt.Errorf("nil client")
	}
	derivedID, err := utils.DeriveObjectID(parentID, typeTag, key)
	if err != nil {
		return nil, err
	}
	return c.GetObject(ctx, derivedID, options, opts...)
}
