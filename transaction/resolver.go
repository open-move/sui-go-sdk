package transaction

import (
	"context"

	"github.com/open-move/sui-go-sdk/types"
)

type OwnerKind int

type ReferenceKind int

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
