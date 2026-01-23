package ed25519

import (
	"bytes"
	cryptoed25519 "crypto/ed25519"
	"crypto/rand"
	"encoding/binary"
	"fmt"

	"github.com/open-move/sui-go-sdk/cryptography/personalmsg"
	"github.com/open-move/sui-go-sdk/cryptography/transaction"
	"github.com/open-move/sui-go-sdk/keychain"
)

const seedSize = 32

// The master key HMAC salt for Ed25519 according to SLIP-0010.
var slip10Key = []byte("ed25519 seed")

type Keypair struct {
	privateKey cryptoed25519.PrivateKey
	publicKey  cryptoed25519.PublicKey
	chainCode  []byte
	path       keychain.DerivationPath
}

func (k Keypair) Scheme() keychain.Scheme {
	return keychain.SchemeEd25519
}

func (k Keypair) PublicKey() []byte {
	return append(cryptoed25519.PublicKey(nil), k.publicKey...)
}

func (k Keypair) SuiAddress() (string, error) {
	return keychain.AddressFromPublicKey(keychain.SchemeEd25519, k.publicKey)
}

func (k Keypair) ExportSecret() ([]byte, error) {
	seed := k.privateKey.Seed()
	out := make([]byte, len(seed))
	copy(out, seed)
	return out, nil
}

func (k Keypair) signData(data []byte) ([]byte, error) {
	if len(k.privateKey) != cryptoed25519.PrivateKeySize {
		return nil, fmt.Errorf("ed25519: invalid private key length %d", len(k.privateKey))
	}

	signature := cryptoed25519.Sign(k.privateKey, data)
	return append([]byte(nil), signature...), nil
}

func verifyDigest(publicKey []byte, digest [32]byte, signature []byte) error {
	if len(publicKey) != cryptoed25519.PublicKeySize {
		return fmt.Errorf("ed25519: invalid public key length %d", len(publicKey))
	}

	expected := 1 + cryptoed25519.SignatureSize + len(publicKey)
	if len(signature) != expected {
		return fmt.Errorf("ed25519: invalid signature length %d", len(signature))
	}
	if signature[0] != keychain.SchemeEd25519.AddressFlag() {
		return fmt.Errorf("ed25519: unexpected signature flag 0x%02x", signature[0])
	}
	if !bytes.Equal(signature[1+cryptoed25519.SignatureSize:], publicKey) {
		return fmt.Errorf("ed25519: mismatched public key")
	}
	rawSig := signature[1 : 1+cryptoed25519.SignatureSize]
	if !cryptoed25519.Verify(cryptoed25519.PublicKey(publicKey), digest[:], rawSig) {
		return fmt.Errorf("ed25519: verification failed")
	}

	return nil
}

func (k Keypair) SignPersonalMessage(message []byte) ([]byte, error) {
	return personalmsg.Sign(
		keychain.SchemeEd25519,
		message,
		k.PublicKey(),
		k.signData,
	)
}

func (k Keypair) SignTransaction(txBytes []byte) ([]byte, error) {
	return transaction.Sign(
		keychain.SchemeEd25519,
		txBytes,
		k.PublicKey(),
		k.signData,
	)
}

func (k Keypair) VerifyPersonalMessage(message []byte, signature []byte) error {
	return VerifyPersonalMessage(k.PublicKey(), message, signature)
}

func VerifyPersonalMessage(publicKey []byte, message []byte, signature []byte) error {
	return personalmsg.Verify(keychain.SchemeEd25519, message, signature, func(digest [32]byte, signature []byte) error {
		return verifyDigest(publicKey, digest, signature)
	})
}

func Generate() (*Keypair, error) {
	pub, priv, err := cryptoed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("ed25519: generate key: %w", err)
	}

	return &Keypair{
		privateKey: priv,
		publicKey:  pub,
	}, nil
}

// Rebuilds the expanded Ed25519 keypair from a 32-byte
// SLIP-0010 seed.
func FromSecretKey(seed []byte) (*Keypair, error) {
	if len(seed) != keychain.PrivateKeySize() {
		return nil, keychain.ErrInvalidSecretLength
	}

	priv := cryptoed25519.NewKeyFromSeed(seed)
	pub := cryptoed25519.PublicKey(priv[32:])
	return &Keypair{
		privateKey: priv,
		publicKey:  pub,
	}, nil
}

// Traverses the SLIP-0010 hardened derivation path starting from the
// supplied master seed. Ed25519 only supports hardened segments; any
// non-hardened index results in an error.
func Derive(seed []byte, path keychain.DerivationPath) (*Keypair, error) {
	if err := path.ValidateForScheme(keychain.SchemeEd25519); err != nil {
		return nil, err
	}

	key, chain := slip10MasterKey(seed)
	for _, segment := range path.Segments() {
		if !segment.Hardened {
			return nil, fmt.Errorf("ed25519: slip-0010 only supports hardened segments")
		}

		// SLIP-0010 hardened child derivation: HMAC-SHA512 with parent chain code
		// over [0x00 || parent_secret || child_index_with_high_bit]. The left half
		// becomes the child secret scalar; the right half becomes the next chain
		// code.
		data := make([]byte, 1+seedSize+4)
		data[0] = 0x00
		copy(data[1:], key)
		binary.BigEndian.PutUint32(data[1+seedSize:], segment.HardenedIndex())
		digest := keychain.HMACSHA512(chain, data)
		key = digest[:seedSize]
		chain = digest[seedSize:]
	}

	privateKey := cryptoed25519.NewKeyFromSeed(key)
	publicKey := cryptoed25519.PublicKey(privateKey[32:])
	return &Keypair{
		privateKey: privateKey,
		publicKey:  publicKey,
		chainCode:  append([]byte{}, chain...),
		path:       path,
	}, nil
}

// Initializes the root private key and chain code according to
// SLIP-0010's Ed25519 specification.
func slip10MasterKey(seed []byte) ([]byte, []byte) {
	digest := keychain.HMACSHA512(slip10Key, seed)
	return digest[:seedSize], digest[seedSize:]
}
