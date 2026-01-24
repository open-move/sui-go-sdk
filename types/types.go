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

type Address [32]byte

type Digest [32]byte

type PersonalMessage struct {
	Message []byte
}

type ObjectRef struct {
	ObjectID Address
	Version  uint64
	Digest   Digest
}

type SharedObjectRef struct {
	ObjectID             Address
	InitialSharedVersion uint64
	Mutable              bool
}

func (a Address) String() string {
	return "0x" + hex.EncodeToString(a[:])
}

func (d Digest) String() string {
	return base58.Encode(d[:])
}
