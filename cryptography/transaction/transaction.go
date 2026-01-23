package transaction

import (
	"fmt"

	"github.com/open-move/sui-go-sdk/cryptography/intent"
	"github.com/open-move/sui-go-sdk/keychain"
)

// Sign hashes transaction bytes with intent scope and serializes the signature
// as `flag || signature || publicKey`.
func Sign(
	scheme keychain.Scheme,
	transactionBytes []byte,
	publicKey []byte,
	signFunc func([]byte) ([]byte, error),
) ([]byte, error) {
	digest, err := intent.HashIntentBytes(intent.IntentScopeTransactionData, transactionBytes)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", scheme.Label(), err)
	}

	sig, err := signFunc(digest[:])
	if err != nil {
		return nil, err
	}
	if len(sig) != 64 {
		return nil, fmt.Errorf("%s: unexpected signature length %d", scheme.Label(), len(sig))
	}

	serialized := make([]byte, 0, 1+len(sig)+len(publicKey))
	serialized = append(serialized, scheme.AddressFlag())
	serialized = append(serialized, sig...)
	serialized = append(serialized, publicKey...)
	return serialized, nil
}
