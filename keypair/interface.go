// Package keypair defines the interface for Sui keypairs and provides utilities for key generation and management.
package keypair

import "github.com/open-move/sui-go-sdk/keychain"

// Keypair defines the interface for Sui keypairs and provides utilities for key generation and management.
type Keypair interface {
	PublicKey() []byte
	Scheme() keychain.Scheme
	SuiAddress() (string, error)
	ExportSecret() ([]byte, error)
	SignTransaction(txBytes []byte) ([]byte, error)
	SignPersonalMessage(message []byte) ([]byte, error)
	VerifyPersonalMessage(message []byte, signature []byte) error
}
