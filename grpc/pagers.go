package grpc

import (
	"context"
	"errors"

	v2 "github.com/open-move/sui-go-sdk/proto/sui/rpc/v2"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

// OwnedObjectsPager iterates through objects returned by ListOwnedObjects, handling pagination internally.
type OwnedObjectsPager struct {
	iter *pageIterator[*v2.Object]
}

// OwnedObjectsPager constructs a pager for ListOwnedObjects using the provided request template.
func (c *Client) OwnedObjectsPager(req *v2.ListOwnedObjectsRequest, opts ...grpc.CallOption) (*OwnedObjectsPager, error) {
	if c == nil {
		return nil, errors.New("nil client")
	}
	if req == nil {
		return nil, errors.New("nil request")
	}

	base := proto.Clone(req).(*v2.ListOwnedObjectsRequest)
	fetch := func(ctx context.Context, token []byte) ([]*v2.Object, []byte, error) {
		if len(token) == 0 {
			base.PageToken = nil
		} else {
			base.PageToken = cloneBytes(token)
		}
		resp, err := c.StateClient().ListOwnedObjects(ctx, base, opts...)
		if err != nil {
			return nil, nil, err
		}
		next := cloneBytes(resp.GetNextPageToken())
		base.PageToken = next
		return resp.GetObjects(), next, nil
	}

	iter, err := newPageIterator(base.GetPageToken(), fetch)
	if err != nil {
		return nil, err
	}

	return &OwnedObjectsPager{iter: iter}, nil
}

// Next fetches the next page of owned objects.
func (p *OwnedObjectsPager) Next(ctx context.Context) ([]*v2.Object, error) {
	if p == nil {
		return nil, errors.New("nil pager")
	}
	return p.iter.Next(ctx)
}

// ForEach visits each object by repeatedly invoking Next until exhaustion or callback error.
func (p *OwnedObjectsPager) ForEach(ctx context.Context, fn func(*v2.Object) error) error {
	if p == nil {
		return errors.New("nil pager")
	}
	return p.iter.ForEach(ctx, fn)
}

// Collect drains the pager and returns all remaining objects.
func (p *OwnedObjectsPager) Collect(ctx context.Context) ([]*v2.Object, error) {
	if p == nil {
		return nil, errors.New("nil pager")
	}
	return p.iter.Collect(ctx)
}

// DynamicFieldsPager iterates through dynamic fields returned by ListDynamicFields.
type DynamicFieldsPager struct {
	iter *pageIterator[*v2.DynamicField]
}

// DynamicFieldsPager constructs a pager for ListDynamicFields using the provided request template.
func (c *Client) DynamicFieldsPager(req *v2.ListDynamicFieldsRequest, opts ...grpc.CallOption) (*DynamicFieldsPager, error) {
	if c == nil {
		return nil, errors.New("nil client")
	}
	if req == nil {
		return nil, errors.New("nil request")
	}

	base := proto.Clone(req).(*v2.ListDynamicFieldsRequest)
	fetch := func(ctx context.Context, token []byte) ([]*v2.DynamicField, []byte, error) {
		if len(token) == 0 {
			base.PageToken = nil
		} else {
			base.PageToken = cloneBytes(token)
		}
		resp, err := c.StateClient().ListDynamicFields(ctx, base, opts...)
		if err != nil {
			return nil, nil, err
		}

		next := cloneBytes(resp.GetNextPageToken())
		base.PageToken = next
		return resp.GetDynamicFields(), next, nil
	}

	iter, err := newPageIterator(base.GetPageToken(), fetch)
	if err != nil {
		return nil, err
	}

	return &DynamicFieldsPager{iter: iter}, nil
}

// Next fetches the next page of dynamic fields.
func (p *DynamicFieldsPager) Next(ctx context.Context) ([]*v2.DynamicField, error) {
	if p == nil {
		return nil, errors.New("nil pager")
	}
	return p.iter.Next(ctx)
}

// ForEach visits each dynamic field emitted by the pager.
func (p *DynamicFieldsPager) ForEach(ctx context.Context, fn func(*v2.DynamicField) error) error {
	if p == nil {
		return errors.New("nil pager")
	}
	return p.iter.ForEach(ctx, fn)
}

// Collect drains the pager and returns all remaining dynamic fields.
func (p *DynamicFieldsPager) Collect(ctx context.Context) ([]*v2.DynamicField, error) {
	if p == nil {
		return nil, errors.New("nil pager")
	}
	return p.iter.Collect(ctx)
}

// BalancesPager iterates through balance entries returned by ListBalances.
type BalancesPager struct {
	iter *pageIterator[*v2.Balance]
}

// BalancesPager constructs a pager for ListBalances using the provided request template.
func (c *Client) BalancesPager(req *v2.ListBalancesRequest, opts ...grpc.CallOption) (*BalancesPager, error) {
	if c == nil {
		return nil, errors.New("nil client")
	}
	if req == nil {
		return nil, errors.New("nil request")
	}

	base := proto.Clone(req).(*v2.ListBalancesRequest)
	fetch := func(ctx context.Context, token []byte) ([]*v2.Balance, []byte, error) {
		if len(token) == 0 {
			base.PageToken = nil
		} else {
			base.PageToken = cloneBytes(token)
		}
		resp, err := c.StateClient().ListBalances(ctx, base, opts...)
		if err != nil {
			return nil, nil, err
		}
		next := cloneBytes(resp.GetNextPageToken())
		base.PageToken = next
		return resp.GetBalances(), next, nil
	}

	iter, err := newPageIterator(base.GetPageToken(), fetch)
	if err != nil {
		return nil, err
	}

	return &BalancesPager{iter: iter}, nil
}

// Next fetches the next page of balances.
func (p *BalancesPager) Next(ctx context.Context) ([]*v2.Balance, error) {
	if p == nil {
		return nil, errors.New("nil pager")
	}
	return p.iter.Next(ctx)
}

// ForEach visits each balance emitted by the pager.
func (p *BalancesPager) ForEach(ctx context.Context, fn func(*v2.Balance) error) error {
	if p == nil {
		return errors.New("nil pager")
	}
	return p.iter.ForEach(ctx, fn)
}

// Collect drains the pager and returns all remaining balances.
func (p *BalancesPager) Collect(ctx context.Context) ([]*v2.Balance, error) {
	if p == nil {
		return nil, errors.New("nil pager")
	}
	return p.iter.Collect(ctx)
}

// PackageVersionsPager iterates through package versions returned by ListPackageVersions.
type PackageVersionsPager struct {
	iter *pageIterator[*v2.PackageVersion]
}

// PackageVersionsPager constructs a pager for ListPackageVersions using the provided request template.
func (c *Client) PackageVersionsPager(req *v2.ListPackageVersionsRequest, opts ...grpc.CallOption) (*PackageVersionsPager, error) {
	if c == nil {
		return nil, errors.New("nil client")
	}
	if req == nil {
		return nil, errors.New("nil request")
	}

	base := proto.Clone(req).(*v2.ListPackageVersionsRequest)
	fetch := func(ctx context.Context, token []byte) ([]*v2.PackageVersion, []byte, error) {
		if len(token) == 0 {
			base.PageToken = nil
		} else {
			base.PageToken = cloneBytes(token)
		}
		resp, err := c.MovePackageClient().ListPackageVersions(ctx, base, opts...)
		if err != nil {
			return nil, nil, err
		}
		next := cloneBytes(resp.GetNextPageToken())
		base.PageToken = next
		return resp.GetVersions(), next, nil
	}

	iter, err := newPageIterator(base.GetPageToken(), fetch)
	if err != nil {
		return nil, err
	}

	return &PackageVersionsPager{iter: iter}, nil
}

// Next fetches the next page of package versions.
func (p *PackageVersionsPager) Next(ctx context.Context) ([]*v2.PackageVersion, error) {
	if p == nil {
		return nil, errors.New("nil pager")
	}
	return p.iter.Next(ctx)
}

// ForEach visits each package version emitted by the pager.
func (p *PackageVersionsPager) ForEach(ctx context.Context, fn func(*v2.PackageVersion) error) error {
	if p == nil {
		return errors.New("nil pager")
	}
	return p.iter.ForEach(ctx, fn)
}

// Collect drains the pager and returns all remaining package versions.
func (p *PackageVersionsPager) Collect(ctx context.Context) ([]*v2.PackageVersion, error) {
	if p == nil {
		return nil, errors.New("nil pager")
	}
	return p.iter.Collect(ctx)
}
