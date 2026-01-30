// Package keypair provides key derivation and management utilities.
package keypair

import (
	"encoding/base64"
	"fmt"

	ed25519keys "github.com/open-move/sui-go-sdk/cryptography/ed25519"
	secp256k1keys "github.com/open-move/sui-go-sdk/cryptography/secp256k1"
	secp256r1keys "github.com/open-move/sui-go-sdk/cryptography/secp256r1"
	"github.com/open-move/sui-go-sdk/keychain"
)

// DeriveFromMnemonic derives a keypair from a BIP-39 mnemonic and derivation path.
func DeriveFromMnemonic(s keychain.Scheme, mnemonic, passphrase, path string) (Keypair, error) {
	parsed, err := keychain.ParseDerivationPath(path)
	if err != nil {
		return nil, err
	}
	seed, err := keychain.SeedFromMnemonic(mnemonic, passphrase)
	if err != nil {
		return nil, err
	}

	switch s {
	case keychain.SchemeEd25519:
		return ed25519keys.Derive(seed, parsed)
	case keychain.SchemeSecp256k1:
		return secp256k1keys.Derive(seed, parsed)
	case keychain.SchemeSecp256r1:
		return secp256r1keys.Derive(seed, parsed)
	default:
		return nil, fmt.Errorf("derive: unsupported scheme %d", s)
	}
}

// Generate generates a new random keypair for the given scheme.
func Generate(s keychain.Scheme) (Keypair, error) {
	switch s {
	case keychain.SchemeEd25519:
		return ed25519keys.Generate()
	case keychain.SchemeSecp256k1:
		return secp256k1keys.Generate()
	case keychain.SchemeSecp256r1:
		return secp256r1keys.Generate()
	default:
		return nil, fmt.Errorf("generate: unsupported scheme %d", s)
	}
}

// FromSecretKey creates a keypair from a raw secret key bytes.
func FromSecretKey(s keychain.Scheme, secret []byte) (Keypair, error) {
	switch s {
	case keychain.SchemeEd25519:
		return ed25519keys.FromSecretKey(secret)
	case keychain.SchemeSecp256k1:
		return secp256k1keys.FromSecretKey(secret)
	case keychain.SchemeSecp256r1:
		return secp256r1keys.FromSecretKey(secret)
	default:
		return nil, fmt.Errorf("from secret: unsupported scheme %d", s)
	}
}

// FromBech32 parses a Bech32-encoded private key (suiprivkey...) into a Keypair.
func FromBech32(encoded string) (Keypair, error) {
	parsed, err := keychain.DecodePrivateKey(encoded)
	if err != nil {
		return nil, err
	}

	kp, err := FromSecretKey(parsed.Scheme, parsed.SecretKey)
	zero(parsed.SecretKey)
	return kp, err
}

func ToBech32(s keychain.Scheme, secret []byte) (string, error) {
	if len(secret) != keychain.PrivateKeySize() {
		return "", fmt.Errorf("export: expected %d secret bytes, got %d", keychain.PrivateKeySize(), len(secret))
	}
	secretCopy := append([]byte(nil), secret...)
	encoded, err := keychain.EncodePrivateKey(s, secretCopy)
	zero(secretCopy)
	return encoded, err
}

// ToBech32FromKeypair encodes a Keypair's secret key into the Sui Bech32 format.
func ToBech32FromKeypair(k SecretExporter) (string, error) {
	if k == nil {
		return "", fmt.Errorf("export: nil keypair")
	}
	secret, err := k.ExportSecret()
	if err != nil {
		return "", err
	}
	encoded, err := ToBech32(k.Scheme(), secret)
	zero(secret)
	return encoded, err
}

func PublicKeyBase64(s keychain.Scheme, publicKey []byte) string {
	payload := append([]byte{s.AddressFlag()}, publicKey...)
	return base64.StdEncoding.EncodeToString(payload)
}

func VerifyPersonalMessage(s keychain.Scheme, publicKey []byte, message []byte, signature []byte) error {
	switch s {
	case keychain.SchemeEd25519:
		return ed25519keys.VerifyPersonalMessage(publicKey, message, signature)
	case keychain.SchemeSecp256k1:
		return secp256k1keys.VerifyPersonalMessage(publicKey, message, signature)
	case keychain.SchemeSecp256r1:
		return secp256r1keys.VerifyPersonalMessage(publicKey, message, signature)
	default:
		return fmt.Errorf("verify personal message: unsupported scheme %d", s)
	}
}

func zero(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
