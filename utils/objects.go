package utils

import (
	"encoding/hex"
	"fmt"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/open-move/sui-go-sdk/types"
	"golang.org/x/crypto/blake2b"
)

const dynamicFieldHashPrefix = 0xf0

func DeriveDynamicFieldID(parentID string, typeTag string, key []byte) (string, error) {
	addr, err := ParseAddress(parentID)
	if err != nil {
		return "", err
	}
	parsedTag, err := types.ParseTypeTag(typeTag)
	if err != nil {
		return "", err
	}

	addressBytes, err := bcs.Marshal(&addr)
	if err != nil {
		return "", fmt.Errorf("bcs address: %w", err)
	}

	tagBytes, err := bcs.Marshal(&parsedTag)
	if err != nil {
		return "", fmt.Errorf("bcs type tag: %w", err)
	}

	keyLength := uint64(len(key))
	keyLengthBytes, err := bcs.Marshal(&keyLength)
	if err != nil {
		return "", fmt.Errorf("bcs key length: %w", err)
	}

	payload := make([]byte, 0, 1+len(addressBytes)+len(keyLengthBytes)+len(key)+len(tagBytes))
	payload = append(payload, dynamicFieldHashPrefix)
	payload = append(payload, addressBytes...)
	payload = append(payload, keyLengthBytes...)
	payload = append(payload, key...)
	payload = append(payload, tagBytes...)
	result := blake2b.Sum256(payload)
	return "0x" + hex.EncodeToString(result[:]), nil
}

func DeriveObjectID(parentID string, typeTag string, key []byte) (string, error) {
	parsedTag, err := types.ParseTypeTag(typeTag)
	if err != nil {
		return "", err
	}
	normalized, err := types.TypeTagString(parsedTag)
	if err != nil {
		return "", err
	}

	derivedType := fmt.Sprintf("0x2::derived_object::DerivedObjectKey<%s>", normalized)
	return DeriveDynamicFieldID(parentID, derivedType, key)
}
