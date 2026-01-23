package secp256k1

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/big"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	secp256k1ecdsa "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"github.com/open-move/sui-go-sdk/cryptography/personalmsg"
	"github.com/open-move/sui-go-sdk/cryptography/transaction"
	"github.com/open-move/sui-go-sdk/keychain"
)

type Keypair struct {
	privateKey *secp256k1.PrivateKey
	publicKey  *secp256k1.PublicKey
	chainCode  []byte
	path       keychain.DerivationPath
}

func (k Keypair) Scheme() keychain.Scheme {
	return keychain.SchemeSecp256k1
}

func (k Keypair) PublicKey() []byte {
	if k.publicKey == nil {
		return nil
	}
	return append([]byte(nil), k.publicKey.SerializeCompressed()...)
}

func (k Keypair) SuiAddress() (string, error) {
	return keychain.AddressFromPublicKey(keychain.SchemeSecp256k1, k.PublicKey())
}

func (k Keypair) ExportSecret() ([]byte, error) {
	if k.privateKey == nil {
		return nil, fmt.Errorf("secp256k1: private key is nil")
	}
	return append([]byte(nil), k.privateKey.Serialize()...), nil
}

func (k Keypair) signData(data []byte) ([]byte, error) {
	if k.privateKey == nil {
		return nil, fmt.Errorf("secp256k1: private key is nil")
	}

	hash := sha256.Sum256(data)
	sig := secp256k1ecdsa.Sign(k.privateKey, hash[:])
	if sig == nil {
		return nil, fmt.Errorf("secp256k1: signing failed")
	}

	rScalar := sig.R()
	sScalar := sig.S()
	if (&sScalar).IsOverHalfOrder() {
		(&sScalar).Negate()
	}

	var rBytes, sBytes [32]byte
	(&rScalar).PutBytes(&rBytes)
	(&sScalar).PutBytes(&sBytes)

	out := make([]byte, 64)
	copy(out[:32], rBytes[:])
	copy(out[32:], sBytes[:])
	return out, nil
}

func verifyDigest(publicKey []byte, digest [32]byte, signature []byte) error {
	pub, err := secp256k1.ParsePubKey(publicKey)
	if err != nil {
		return fmt.Errorf("secp256k1: invalid public key: %w", err)
	}

	expectedLen := 1 + 64 + len(publicKey)
	if len(signature) != expectedLen {
		return fmt.Errorf("secp256k1: invalid signature length %d", len(signature))
	}
	if signature[0] != keychain.SchemeSecp256k1.AddressFlag() {
		return fmt.Errorf("secp256k1: unexpected signature flag 0x%02x", signature[0])
	}
	if !bytes.Equal(signature[65:], publicKey) {
		return fmt.Errorf("secp256k1: mismatched public key")
	}

	var rScalar, sScalar secp256k1.ModNScalar
	if overflow := rScalar.SetByteSlice(signature[1:33]); overflow {
		return fmt.Errorf("secp256k1: invalid R component")
	}
	if overflow := sScalar.SetByteSlice(signature[33:65]); overflow {
		return fmt.Errorf("secp256k1: invalid S component")
	}

	sig := secp256k1ecdsa.NewSignature(&rScalar, &sScalar)
	hash := sha256.Sum256(digest[:])
	if !sig.Verify(hash[:], pub) {
		return fmt.Errorf("secp256k1: verification failed")
	}
	return nil
}

func (k Keypair) SignPersonalMessage(message []byte) ([]byte, error) {
	return personalmsg.Sign(
		keychain.SchemeSecp256k1,
		message,
		k.PublicKey(),
		k.signData,
	)
}

func (k Keypair) SignTransaction(txBytes []byte) ([]byte, error) {
	return transaction.Sign(
		keychain.SchemeSecp256k1,
		txBytes,
		k.PublicKey(),
		k.signData,
	)
}

func (k Keypair) VerifyPersonalMessage(message []byte, signature []byte) error {
	return VerifyPersonalMessage(k.PublicKey(), message, signature)
}

func VerifyPersonalMessage(publicKey []byte, message []byte, signature []byte) error {
	return personalmsg.Verify(keychain.SchemeSecp256k1, message, signature, func(digest [32]byte, signature []byte) error {
		return verifyDigest(publicKey, digest, signature)
	})
}

func Generate() (*Keypair, error) {
	priv, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		return nil, fmt.Errorf("secp256k1: generate key: %w", err)
	}

	return &Keypair{
		privateKey: priv,
		publicKey:  priv.PubKey(),
	}, nil
}

func FromSecretKey(secret []byte) (*Keypair, error) {
	if len(secret) != keychain.PrivateKeySize() {
		return nil, keychain.ErrInvalidSecretLength
	}

	order := secp256k1.S256().N
	k := new(big.Int).SetBytes(secret)
	if k.Sign() <= 0 || k.Cmp(order) >= 0 {
		return nil, fmt.Errorf("secp256k1: private key out of range")
	}

	priv := secp256k1.PrivKeyFromBytes(secret)
	if priv == nil || priv.Key.IsZeroBit() == 1 {
		return nil, fmt.Errorf("secp256k1: invalid private key")
	}

	return &Keypair{
		privateKey: priv,
		publicKey:  priv.PubKey(),
	}, nil
}

// Walks the BIP-32 derivation path (allowing both hardened and unhardened
// steps) starting from the provided seed and returns the resulting keypair and
// chain code.
func Derive(seed []byte, path keychain.DerivationPath) (*Keypair, error) {
	if err := path.ValidateForScheme(keychain.SchemeSecp256k1); err != nil {
		return nil, err
	}

	key, chain := keychain.BIP32MasterPrivateKey(seed)
	segments := path.Segments()
	for _, segment := range segments {
		nextKey, nextChain, err := keychain.DeriveChildPrivateKey(key, chain, segment, func(priv []byte) ([]byte, error) {
			privKey := secp256k1.PrivKeyFromBytes(priv)
			if privKey == nil || privKey.Key.IsZeroBit() == 1 {
				return nil, fmt.Errorf("secp256k1: invalid private key")
			}

			return privKey.PubKey().SerializeCompressed(), nil
		}, secp256k1.S256().N)
		if err != nil {
			return nil, err
		}

		key = nextKey
		chain = nextChain
	}

	privKey := secp256k1.PrivKeyFromBytes(key)
	if privKey == nil || privKey.Key.IsZeroBit() == 1 {
		return nil, fmt.Errorf("secp256k1: failed to build private key")
	}

	pubKey := privKey.PubKey()
	return &Keypair{
		privateKey: privKey,
		publicKey:  pubKey,
		chainCode:  append([]byte{}, chain...),
		path:       path,
	}, nil
}
