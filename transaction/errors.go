package transaction

import "errors"

var (
	ErrNilBuilder              = errors.New("transaction builder is nil")
	ErrInvalidAddress          = errors.New("invalid sui address")
	ErrInvalidDigest           = errors.New("invalid object digest")
	ErrInvalidSerializedSig    = errors.New("invalid serialized signature")
	ErrMissingMoveCallTarget   = errors.New("move call target or package/module/function required")
	ErrUnresolvedInput         = errors.New("transaction input unresolved")
	ErrResolverRequired        = errors.New("resolver required to resolve object inputs")
	ErrMissingProgrammableKind = errors.New("programmable transaction required")
)
