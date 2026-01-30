// Package keychain provides key derivation, management, and address generation utilities.
package keychain

import (
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/blake2b"
)

const addressLength = 32

// AddressFromPublicKey derives the Sui address for the given signature scheme and public key bytes.
func AddressFromPublicKey(s Scheme, publicKey []byte) (string, error) {
	if len(publicKey) == 0 {
		return "", fmt.Errorf("address: public key must not be empty")
	}

	flag := s.AddressFlag()
	if flag == 0xff {
		return "", fmt.Errorf("address: unsupported scheme %d", s)
	}

	hasher, err := blake2b.New256(nil)
	if err != nil {
		return "", fmt.Errorf("address: blake2b init: %w", err)
	}

	if _, err := hasher.Write([]byte{flag}); err != nil {
		return "", fmt.Errorf("address: write flag: %w", err)
	}

	if _, err := hasher.Write(publicKey); err != nil {
		return "", fmt.Errorf("address: write public key: %w", err)
	}

	digest := hasher.Sum(nil)
	if len(digest) != addressLength {
		return "", fmt.Errorf("address: unexpected digest length %d", len(digest))
	}

	return "0x" + hex.EncodeToString(digest), nil
}
