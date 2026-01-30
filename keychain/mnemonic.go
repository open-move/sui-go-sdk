package keychain

import (
	"crypto/rand"
	"fmt"

	bip39 "github.com/tyler-smith/go-bip39"
)

// NewMnemonic creates a mnemonic with the requested entropy bits (128-256 in steps of 32).
func NewMnemonic(entropyBits int) (string, error) {
	if entropyBits%32 != 0 || entropyBits < 128 || entropyBits > 256 {
		return "", fmt.Errorf("mnemonic: entropy must be 128-256 bits in 32-bit steps")
	}

	entropy := make([]byte, entropyBits/8)
	if _, err := rand.Read(entropy); err != nil {
		return "", fmt.Errorf("mnemonic: entropy: %w", err)
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", fmt.Errorf("mnemonic: generate: %w", err)
	}

	return mnemonic, nil
}

// SeedFromMnemonic returns the BIP-39 seed bytes (64 bytes) using the optional passphrase.
func SeedFromMnemonic(mnemonic, passphrase string) ([]byte, error) {
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, fmt.Errorf("mnemonic: invalid phrase")
	}

	seed := bip39.NewSeed(mnemonic, passphrase)
	if len(seed) == 0 {
		return nil, fmt.Errorf("mnemonic: empty seed derived")
	}

	return seed, nil
}
