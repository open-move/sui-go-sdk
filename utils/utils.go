package utils

import (
	"fmt"
	"math/big"
)

func StringPtr(value string) *string {
	return &value
}

func EncodeU256(value *big.Int) ([]byte, error) {
	if value.Sign() < 0 {
		return nil, fmt.Errorf("u256 value must be positive")
	}

	if value.BitLen() > 256 {
		return nil, fmt.Errorf("u256 value out of range")
	}

	buf := make([]byte, 32)
	bigBytes := value.Bytes()
	copy(buf[32-len(bigBytes):], bigBytes)
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}

	return buf, nil
}

func UniqueValues[T comparable](values []T) []T {
	if len(values) == 0 {
		return nil
	}

	seen := make(map[T]struct{}, len(values))
	unique := make([]T, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		unique = append(unique, value)
	}

	return unique
}
