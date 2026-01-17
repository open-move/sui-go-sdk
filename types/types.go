package types

import (
	"encoding/hex"

	"github.com/btcsuite/btcutil/base58"
)

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
