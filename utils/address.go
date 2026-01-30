package utils

import (
	"encoding/hex"
	"errors"
	"strings"

	"github.com/open-move/sui-go-sdk/types"
)

var ErrInvalidAddress = errors.New("invalid sui address")

func NormalizeAddress(input string) (string, error) {
	trimmed := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(input)), "0x")

	if trimmed == "" {
		return "", ErrInvalidAddress
	}

	if len(trimmed) > 64 {
		return "", ErrInvalidAddress
	}

	padded := strings.Repeat("0", 64-len(trimmed)) + trimmed
	if _, err := hex.DecodeString(padded); err != nil {
		return "", ErrInvalidAddress
	}

	return "0x" + padded, nil
}

// ParseAddress parses a hex string into a Sui Address type.
func ParseAddress(input string) (types.Address, error) {
	normalized, err := NormalizeAddress(input)
	if err != nil {
		return types.Address{}, err
	}

	decoded, err := hex.DecodeString(normalized[2:])
	if err != nil || len(decoded) != len(types.Address{}) {
		return types.Address{}, ErrInvalidAddress
	}

	var addr types.Address
	copy(addr[:], decoded)
	return addr, nil
}
