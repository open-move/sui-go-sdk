// Package types defines common Sui data types and their BCS serialization rules.
package types

import (
	"encoding/hex"

	bcs "github.com/iotaledger/bcs-go"

	"github.com/btcsuite/btcutil/base58"
)

func init() {
	// Register custom BCS encoder for Digest to serialize as Vec<u8> with length prefix
	// This matches Sui's ObjectDigest serialization format
	bcs.AddCustomEncoder(func(e *bcs.Encoder, d Digest) error {
		// Write length prefix (32)
		e.WriteLen(32)
		// Write the 32 bytes
		_, err := e.Write(d[:])
		return err
	})

	// Register custom BCS decoder for Digest
	bcs.AddCustomDecoder(func(d *bcs.Decoder, dig *Digest) error {
		// Read length prefix
		length := d.ReadLen()
		if length != 32 {
			return d.Err()
		}
		// Read the 32 bytes
		_, err := d.Read(dig[:])
		return err
	})
}

// Address represents a 32-byte Sui address.
type Address [32]byte

// Digest represents a 32-byte transaction or object digest.
type Digest [32]byte

// PersonalMessage represents a personal message to be signed.
type PersonalMessage struct {
	Message []byte
}

// ObjectRef represents a reference to a Sui object, including its ID, version, and digest.
type ObjectRef struct {
	ObjectID Address
	Version  uint64
	Digest   Digest
}

// SharedObjectRef represents a reference to a shared Sui object.
type SharedObjectRef struct {
	ObjectID             Address
	InitialSharedVersion uint64
	Mutable              bool
}

// String returns the hex-encoded string representation of the address, prefixed with "0x".
func (a Address) String() string {
	return "0x" + hex.EncodeToString(a[:])
}

// String returns the Base58-encoded string representation of the digest.
func (d Digest) String() string {
	return base58.Encode(d[:])
}
