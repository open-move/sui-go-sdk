package personalmsg

import (
	"errors"
	"fmt"

	"github.com/open-move/sui-go-sdk/cryptography/intent"
	"github.com/open-move/sui-go-sdk/keychain"
	"github.com/open-move/sui-go-sdk/types"
)

var ErrEmptyPersonalMessage = errors.New("personal message: empty message")

// Sign builds the intent-scoped payload for personal messages, hashes it per Sui
// rules, and returns the serialized signature bytes `flag || sig || pubkey`.
func Sign(
	scheme keychain.Scheme,
	message []byte,
	publicKey []byte,
	signFunc func([]byte) ([]byte, error),
) ([]byte, error) {
	digest, err := digest(message)
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

func Verify(
	scheme keychain.Scheme,
	message []byte,
	signature []byte,
	verifyFunc func([32]byte, []byte) error,
) error {
	digest, err := digest(message)
	if err != nil {
		return fmt.Errorf("%s: %w", scheme.Label(), err)
	}

	return verifyFunc(digest, signature)
}

func digest(message []byte) ([32]byte, error) {
	if len(message) == 0 {
		return [32]byte{}, ErrEmptyPersonalMessage
	}

	payload := append([]byte(nil), message...)
	intentMsg := intent.NewIntentMessage(
		intent.DefaultIntent().WithScope(intent.IntentScopePersonalMessage),
		types.PersonalMessage{Message: payload},
	)

	digest, err := intent.HashIntentMessage(intentMsg)
	if err != nil {
		return [32]byte{}, fmt.Errorf("hash intent message: %w", err)
	}

	return digest, nil
}
