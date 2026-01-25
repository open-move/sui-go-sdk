package utils

import (
	"errors"
	"fmt"
	"strings"

	"github.com/btcsuite/btcutil/base58"
	"github.com/open-move/sui-go-sdk/types"
	"github.com/open-move/sui-go-sdk/typetag"
)

var ErrInvalidDigest = errors.New("invalid object digest")

const digestLength = 32

func ParseDigest(input string) (types.Digest, error) {
	decoded := base58.Decode(input)
	if len(decoded) != digestLength {
		return nil, ErrInvalidDigest
	}
	return append(types.Digest(nil), decoded...), nil
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

func ParseStructTag(input string) (typetag.StructTag, error) {
	tag, err := ParseTypeTag(input)
	if err != nil {
		return typetag.StructTag{}, err
	}
	if tag.Struct == nil {
		return typetag.StructTag{}, fmt.Errorf("type tag %q is not a struct", input)
	}
	return *tag.Struct, nil
}

func ParseTypeTag(input string) (typetag.TypeTag, error) {
	trimmed := strings.TrimSpace(input)
	switch trimmed {
	case "bool":
		return typetag.TypeTagBool(), nil
	case "u8":
		return typetag.TypeTagU8(), nil
	case "u16":
		return typetag.TypeTagU16(), nil
	case "u32":
		return typetag.TypeTagU32(), nil
	case "u64":
		return typetag.TypeTagU64(), nil
	case "u128":
		return typetag.TypeTagU128(), nil
	case "u256":
		return typetag.TypeTagU256(), nil
	case "address":
		return typetag.TypeTagAddress(), nil
	case "signer":
		return typetag.TypeTagSigner(), nil
	}

	if strings.HasPrefix(trimmed, "vector<") && strings.HasSuffix(trimmed, ">") {
		inner := strings.TrimSuffix(strings.TrimPrefix(trimmed, "vector<"), ">")
		innerTag, err := ParseTypeTag(inner)
		if err != nil {
			return typetag.TypeTag{}, err
		}

		return typetag.TypeTagVector(innerTag), nil
	}

	parts := strings.SplitN(trimmed, "::", 3)
	if len(parts) != 3 {
		return typetag.TypeTag{}, fmt.Errorf("invalid type tag %q", input)
	}

	address, module, namePart := parts[0], parts[1], parts[2]
	if address == "" || module == "" || namePart == "" {
		return typetag.TypeTag{}, fmt.Errorf("invalid type tag %q", input)
	}

	name, typeArgsStr, hasArgs, err := splitStructType(namePart)
	if err != nil {
		return typetag.TypeTag{}, err
	}

	normalized, err := NormalizeAddress(address)
	if err != nil {
		return typetag.TypeTag{}, err
	}
	addr, err := ParseAddress(normalized)
	if err != nil {
		return typetag.TypeTag{}, err
	}

	var typeParams []typetag.TypeTag
	if hasArgs {
		parts, err := splitGenericParams(typeArgsStr)
		if err != nil {
			return typetag.TypeTag{}, err
		}
		typeParams = make([]typetag.TypeTag, len(parts))
		for i, part := range parts {
			tag, err := ParseTypeTag(part)
			if err != nil {
				return typetag.TypeTag{}, err
			}
			typeParams[i] = tag
		}
	}

	return typetag.TypeTagStruct(typetag.NewStructTag(addr, module, name, typeParams)), nil
}

func splitStructType(input string) (string, string, bool, error) {
	openIdx := strings.Index(input, "<")
	if openIdx == -1 {
		return input, "", false, nil
	}
	if !strings.HasSuffix(input, ">") {
		return "", "", false, fmt.Errorf("invalid struct type %q", input)
	}
	name := input[:openIdx]
	if name == "" {
		return "", "", false, fmt.Errorf("invalid struct type %q", input)
	}
	return name, input[openIdx+1 : len(input)-1], true, nil
}

func splitGenericParams(input string) ([]string, error) {
	var parts []string
	depth := 0
	start := 0
	for i, r := range input {
		switch r {
		case '<':
			depth++
		case '>':
			if depth == 0 {
				return nil, fmt.Errorf("invalid type parameters %q", input)
			}
			depth--
		case ',':
			if depth == 0 {
				part := strings.TrimSpace(input[start:i])
				if part == "" {
					return nil, fmt.Errorf("invalid type parameters %q", input)
				}
				parts = append(parts, part)
				start = i + 1
			}
		}
	}
	if depth != 0 {
		return nil, fmt.Errorf("invalid type parameters %q", input)
	}
	last := strings.TrimSpace(input[start:])
	if last == "" {
		return nil, fmt.Errorf("invalid type parameters %q", input)
	}
	parts = append(parts, last)
	return parts, nil
}
