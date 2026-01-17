package utils

import (
	"errors"
	"fmt"
	"strings"

	"github.com/btcsuite/btcutil/base58"
	"github.com/open-move/sui-go-sdk/types"
)

var ErrInvalidDigest = errors.New("invalid object digest")

func ParseDigest(input string) (types.Digest, error) {
	decoded := base58.Decode(input)
	if len(decoded) != len(types.Digest{}) {
		return types.Digest{}, ErrInvalidDigest
	}

	var digest types.Digest
	copy(digest[:], decoded)
	return digest, nil
}

func ParseMoveCallTarget(target string) (string, string, string, error) {
	parts := strings.Split(target, "::")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("move call target must be package::module::function")
	}

	if parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("move call target must be package::module::function")
	}

	return parts[0], parts[1], parts[2], nil
}

func ParseObjectRef(objectID string, version uint64, digest string) (types.ObjectRef, error) {
	addr, err := ParseAddress(objectID)
	if err != nil {
		return types.ObjectRef{}, err
	}

	dig, err := ParseDigest(digest)
	if err != nil {
		return types.ObjectRef{}, err
	}

	return types.ObjectRef{ObjectID: addr, Version: version, Digest: dig}, nil
}
