package transaction

import (
	"fmt"

	"github.com/open-move/sui-go-sdk/keychain"
	v2 "github.com/open-move/sui-go-sdk/proto/sui/rpc/v2"
)

const (
	signatureLength = 64
)

// UserSignatureFromSerialized parses a serialized signature in the form
// `flag || signature || publicKey` and produces a gRPC UserSignature.
func UserSignatureFromSerialized(serialized []byte) (*v2.UserSignature, error) {
	scheme, sig, pub, err := parseSerializedSignature(serialized)
	if err != nil {
		return nil, err
	}

	grpcScheme, err := grpcSchemeForKeychain(scheme)
	if err != nil {
		return nil, err
	}

	simple := &v2.SimpleSignature{
		Scheme:    grpcScheme.Enum(),
		Signature: sig,
		PublicKey: pub,
	}
	return &v2.UserSignature{
		Scheme:    grpcScheme.Enum(),
		Signature: &v2.UserSignature_Simple{Simple: simple},
	}, nil
}

func parseSerializedSignature(serialized []byte) (keychain.Scheme, []byte, []byte, error) {
	if len(serialized) < 1+signatureLength {
		return 0, nil, nil, ErrInvalidSerializedSig
	}
	flag := serialized[0]
	scheme, err := keychain.SchemeFromFlag(flag)
	if err != nil {
		return 0, nil, nil, ErrInvalidSerializedSig
	}
	pubSize, err := publicKeySizeForScheme(scheme)
	if err != nil {
		return 0, nil, nil, err
	}
	expected := 1 + signatureLength + pubSize
	if len(serialized) != expected {
		return 0, nil, nil, ErrInvalidSerializedSig
	}

	sig := append([]byte(nil), serialized[1:1+signatureLength]...)
	pub := append([]byte(nil), serialized[1+signatureLength:]...)
	return scheme, sig, pub, nil
}

func publicKeySizeForScheme(scheme keychain.Scheme) (int, error) {
	switch scheme {
	case keychain.SchemeEd25519:
		return 32, nil
	case keychain.SchemeSecp256k1, keychain.SchemeSecp256r1:
		return 33, nil
	default:
		return 0, fmt.Errorf("unsupported signature scheme %d", scheme)
	}
}

func grpcSchemeForKeychain(scheme keychain.Scheme) (v2.SignatureScheme, error) {
	switch scheme {
	case keychain.SchemeEd25519:
		return v2.SignatureScheme_ED25519, nil
	case keychain.SchemeSecp256k1:
		return v2.SignatureScheme_SECP256K1, nil
	case keychain.SchemeSecp256r1:
		return v2.SignatureScheme_SECP256R1, nil
	default:
		return v2.SignatureScheme_ED25519, fmt.Errorf("unsupported signature scheme %d", scheme)
	}
}
