package secp256r1

import (
	"bytes"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/open-move/sui-go-sdk/cryptography/personalmsg"
	"github.com/open-move/sui-go-sdk/cryptography/transaction"
	"github.com/open-move/sui-go-sdk/keychain"
)

type Keypair struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey
	ChainCode  []byte
	Path       keychain.DerivationPath
}

func (k Keypair) Scheme() keychain.Scheme {
	return keychain.SchemeSecp256r1
}

func (k Keypair) PrivateKeyBytes() []byte {
	b := k.PrivateKey.D.Bytes()
	if len(b) < 32 {
		padded := make([]byte, 32)
		copy(padded[32-len(b):], b)
		b = padded
	}
	return b
}

func (k Keypair) PublicKeyBytes() []byte {
	curve := elliptic.P256()
	return elliptic.MarshalCompressed(curve, k.PrivateKey.X, k.PrivateKey.Y)
}

func (k Keypair) SuiAddress() (string, error) {
	return keychain.AddressFromPublicKey(keychain.SchemeSecp256r1, k.PublicKeyBytes())
}

func (k Keypair) SecretKeyBytes() []byte {
	return k.PrivateKeyBytes()
}

func (k Keypair) ExportSecret() ([]byte, error) {
	return k.SecretKeyBytes(), nil
}

func (k Keypair) signData(data []byte) ([]byte, error) {
	if k.PrivateKey == nil {
		return nil, fmt.Errorf("secp256r1: private key is nil")
	}

	sig, err := deterministicP256Signature(k.PrivateKey, data)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

func verifyDigest(publicKey []byte, digest [32]byte, signature []byte) error {
	curve := elliptic.P256()
	x, y := elliptic.UnmarshalCompressed(curve, publicKey)
	if x == nil || y == nil {
		return fmt.Errorf("secp256r1: invalid public key")
	}
	pubKey := &ecdsa.PublicKey{Curve: curve, X: x, Y: y}

	expectedLen := 1 + 64 + len(publicKey)
	if len(signature) != expectedLen {
		return fmt.Errorf("secp256r1: invalid signature length %d", len(signature))
	}
	if signature[0] != keychain.SchemeSecp256r1.AddressFlag() {
		return fmt.Errorf("secp256r1: unexpected signature flag 0x%02x", signature[0])
	}
	if !bytes.Equal(signature[65:], publicKey) {
		return fmt.Errorf("secp256r1: mismatched public key")
	}

	r := new(big.Int).SetBytes(signature[1:33])
	s := new(big.Int).SetBytes(signature[33:65])
	msgHash := sha256.Sum256(digest[:])
	if !ecdsa.Verify(pubKey, msgHash[:], r, s) {
		return fmt.Errorf("secp256r1: verification failed")
	}

	return nil
}

func (k Keypair) SignPersonalMessage(message []byte) ([]byte, error) {
	return personalmsg.Sign(
		keychain.SchemeSecp256r1,
		message,
		k.PublicKeyBytes(),
		k.signData,
	)
}

func (k Keypair) SignTransaction(txBytes []byte) ([]byte, error) {
	return transaction.Sign(
		keychain.SchemeSecp256r1,
		txBytes,
		k.PublicKeyBytes(),
		k.signData,
	)
}

func (k Keypair) VerifyPersonalMessage(message []byte, signature []byte) error {
	return VerifyPersonalMessage(k.PublicKeyBytes(), message, signature)
}

func VerifyPersonalMessage(publicKey []byte, message []byte, signature []byte) error {
	return personalmsg.Verify(keychain.SchemeSecp256r1, message, signature, func(digest [32]byte, signature []byte) error {
		return verifyDigest(publicKey, digest, signature)
	})
}

func Generate() (*Keypair, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("secp256r1: generate key: %w", err)
	}
	return &Keypair{
		PrivateKey: priv,
		PublicKey:  &priv.PublicKey,
	}, nil
}

func FromSecretKey(secret []byte) (*Keypair, error) {
	priv, _, err := deriveECDSAKey(secret)
	if err != nil {
		return nil, err
	}

	return &Keypair{PrivateKey: priv, PublicKey: &priv.PublicKey}, nil
}

// Mirrors Mysten's tooling by running BIP-32 derivation with secp256k1 group
// operations before converting the resulting scalar into a P-256 keypair.
func Derive(seed []byte, path keychain.DerivationPath) (*Keypair, error) {
	if err := path.ValidateForScheme(keychain.SchemeSecp256r1); err != nil {
		return nil, err
	}

	key, chain := keychain.BIP32MasterPrivateKey(seed)
	secpOrder := secp256k1.S256().N
	for _, segment := range path.Segments() {
		nextKey, nextChain, err := keychain.DeriveChildPrivateKey(key, chain, segment, func(s []byte) ([]byte, error) {
			privKey := secp256k1.PrivKeyFromBytes(s)
			if privKey == nil || privKey.Key.IsZeroBit() == 1 {
				return nil, fmt.Errorf("secp256r1: invalid intermediate key")
			}
			return privKey.PubKey().SerializeCompressed(), nil
		}, secpOrder)
		if err != nil {
			return nil, err
		}
		key = nextKey
		chain = nextChain
	}

	priv, _, err := deriveECDSAKey(key)
	if err != nil {
		return nil, err
	}

	return &Keypair{
		PrivateKey: priv,
		PublicKey:  &priv.PublicKey,
		ChainCode:  append([]byte{}, chain...),
		Path:       path,
	}, nil
}

// Converts a raw scalar into an ECDSA private key using crypto/ecdh to obtain
// the corresponding public point and compressed SEC1 encoding.
func deriveECDSAKey(secret []byte) (*ecdsa.PrivateKey, []byte, error) {
	if len(secret) != keychain.PrivateKeySize() {
		return nil, nil, keychain.ErrInvalidSecretLength
	}
	secretCopy := make([]byte, keychain.PrivateKeySize())
	copy(secretCopy, secret)
	curve := ecdh.P256()
	priv, err := curve.NewPrivateKey(secretCopy)
	zero(secretCopy)
	if err != nil {
		return nil, nil, fmt.Errorf("secp256r1: invalid private key: %w", err)
	}
	ecCurve := elliptic.P256()
	pubBytes := priv.PublicKey().Bytes()
	if len(pubBytes) != 1+2*keychain.PrivateKeySize() || pubBytes[0] != 0x04 {
		return nil, nil, fmt.Errorf("secp256r1: unexpected public key encoding")
	}
	x := new(big.Int).SetBytes(pubBytes[1 : 1+keychain.PrivateKeySize()])
	y := new(big.Int).SetBytes(pubBytes[1+keychain.PrivateKeySize():])
	zero(pubBytes)
	compressed := elliptic.MarshalCompressed(ecCurve, x, y)
	d := new(big.Int).SetBytes(secret)
	order := ecCurve.Params().N
	if d.Sign() <= 0 || d.Cmp(order) >= 0 {
		zero(compressed)
		return nil, nil, fmt.Errorf("secp256r1: private key out of range")
	}
	ecdsaPriv := &ecdsa.PrivateKey{D: d}
	ecdsaPriv.PublicKey = ecdsa.PublicKey{Curve: ecCurve, X: x, Y: y}
	return ecdsaPriv, compressed, nil
}

func zero(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

func deterministicP256Signature(priv *ecdsa.PrivateKey, digest []byte) ([]byte, error) {
	curve := priv.Curve.Params()
	order := curve.N
	qlen := order.BitLen()
	rlen := (qlen + 7) / 8
	msgHash := sha256.Sum256(digest)

	int2octets := func(x *big.Int) []byte {
		out := make([]byte, rlen)
		x.FillBytes(out)
		return out
	}

	bits2int := func(b []byte) *big.Int {
		i := new(big.Int).SetBytes(b)
		blen := len(b) * 8
		if blen > qlen {
			i.Rsh(i, uint(blen-qlen))
		}
		return i
	}

	bits2octets := func(b []byte) []byte {
		i := bits2int(b)
		i.Mod(i, order)
		return int2octets(i)
	}

	privBytes := int2octets(priv.D)
	hashOctets := bits2octets(msgHash[:])

	v := bytes.Repeat([]byte{0x01}, sha256.Size)
	k := make([]byte, sha256.Size)

	hmacSHA := func(key []byte, data ...[]byte) []byte {
		mac := hmac.New(sha256.New, key)
		for _, d := range data {
			mac.Write(d)
		}
		return mac.Sum(nil)
	}

	mk := func(parts ...[]byte) []byte {
		buf := make([]byte, 0, len(v)+1+len(privBytes)+len(hashOctets))
		for _, part := range parts {
			buf = append(buf, part...)
		}
		return buf
	}

	k = hmacSHA(k, mk(v, []byte{0x00}, privBytes, hashOctets))
	v = hmacSHA(k, v)
	k = hmacSHA(k, mk(v, []byte{0x01}, privBytes, hashOctets))
	v = hmacSHA(k, v)

	for {
		var t []byte
		for len(t) < rlen {
			v = hmacSHA(k, v)
			t = append(t, v...)
		}
		candidate := bits2int(t[:rlen])
		if candidate.Sign() > 0 && candidate.Cmp(order) < 0 {
			kBytes := make([]byte, rlen)
			candidate.FillBytes(kBytes)
			// Use crypto/ecdh to avoid deprecated scalar multiplication APIs.
			ecdhCurve := ecdh.P256()
			ephemeralPriv, err := ecdhCurve.NewPrivateKey(kBytes)
			if err != nil {
				goto retry
			}
			pubBytes := ephemeralPriv.PublicKey().Bytes()
			if len(pubBytes) != 1+2*rlen || pubBytes[0] != 0x04 {
				goto retry
			}
			x := new(big.Int).SetBytes(pubBytes[1 : 1+rlen])
			r := new(big.Int).Mod(x, order)
			if r.Sign() == 0 {
				goto retry
			}
			e := bits2int(msgHash[:])
			e.Mod(e, order)
			kInv := new(big.Int).ModInverse(candidate, order)
			if kInv == nil {
				goto retry
			}
			s := new(big.Int).Mul(priv.D, r)
			s.Add(s, e)
			s.Mul(s, kInv)
			s.Mod(s, order)
			if s.Sign() == 0 {
				goto retry
			}
			halfOrder := new(big.Int).Rsh(new(big.Int).Set(order), 1)
			if s.Cmp(halfOrder) > 0 {
				s.Sub(order, s)
			}
			out := make([]byte, 64)
			r.FillBytes(out[:32])
			s.FillBytes(out[32:])
			return out, nil
		}

	retry:
		k = hmacSHA(k, mk(v, []byte{0x00}))
		v = hmacSHA(k, v)
	}
}
