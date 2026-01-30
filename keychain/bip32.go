package keychain

import (
	"encoding/binary"
	"fmt"
	"math/big"
)

const bip32MasterKey = "Bitcoin seed"

// BIP32MasterPrivateKey returns the master private key and chain code.
func BIP32MasterPrivateKey(seed []byte) ([]byte, []byte) {
	digest := HMACSHA512([]byte(bip32MasterKey), seed)
	return digest[:privateKeySize], digest[privateKeySize:]
}

// DeriveChildPrivateKey derives the BIP32 child for an ECDSA curve.
func DeriveChildPrivateKey(privKey, chainCode []byte, segment PathSegment, pubKeyFunc func([]byte) ([]byte, error), order *big.Int) ([]byte, []byte, error) {
	var data []byte
	index := segment.Index
	if segment.Hardened {
		data = make([]byte, 1+privateKeySize)
		data[0] = 0x00
		copy(data[1:], privKey)
		index = segment.HardenedIndex()
	} else {
		pub, err := pubKeyFunc(privKey)
		if err != nil {
			return nil, nil, err
		}
		data = make([]byte, len(pub))
		copy(data, pub)
	}
	indexBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(indexBytes, index)
	data = append(data, indexBytes...)

	digest := HMACSHA512(chainCode, data)
	il := new(big.Int).SetBytes(digest[:32])
	if il.Cmp(order) >= 0 {
		return nil, nil, fmt.Errorf("bip32: derived IL >= curve order")
	}
	pkInt := new(big.Int).SetBytes(privKey)
	childInt := new(big.Int).Add(il, pkInt)
	childInt.Mod(childInt, order)
	if childInt.Sign() == 0 {
		return nil, nil, fmt.Errorf("bip32: derived zero key")
	}
	childKey := childInt.Bytes()
	if len(childKey) < privateKeySize {
		padded := make([]byte, privateKeySize)
		copy(padded[privateKeySize-len(childKey):], childKey)
		childKey = padded
	}
	childChain := digest[privateKeySize:]
	return childKey, childChain, nil
}
