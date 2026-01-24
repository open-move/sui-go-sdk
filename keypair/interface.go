package keypair

import "github.com/open-move/sui-go-sdk/keychain"

type Keypair interface {
	PublicKey() []byte
	Scheme() keychain.Scheme
	SuiAddress() (string, error)
	ExportSecret() ([]byte, error)
	SignTransaction(txBytes []byte) ([]byte, error)
	SignPersonalMessage(message []byte) ([]byte, error)
	VerifyPersonalMessage(message []byte, signature []byte) error
}
