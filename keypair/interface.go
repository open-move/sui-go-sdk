// Package keypair defines the interface for Sui keypairs and provides utilities for key generation and management.
package keypair

import "github.com/open-move/sui-go-sdk/keychain"

// Keypair defines the interface for Sui keypairs and provides utilities for key generation and management.
type Keypair interface {
	Scheme() keychain.Scheme
	SuiAddress() (string, error)
	SignTransaction(txBytes []byte) ([]byte, error)
	SignPersonalMessage(message []byte) ([]byte, error)
	VerifyPersonalMessage(message []byte, signature []byte) error
}

// PublicKeyer exposes raw public key bytes for a keypair.
// It is intentionally separate from the base Keypair interface.
type PublicKeyer interface {
	PublicKey() []byte
}

// SecretExporter opts into exporting the 32-byte secret material for encoding.
type SecretExporter interface {
	Scheme() keychain.Scheme
	ExportSecret() ([]byte, error)
}
