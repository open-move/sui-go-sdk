// Package types defines common Sui data types and their BCS serialization rules.
package types

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

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
		buf := make([]byte, length)
		_, err := d.Read(buf)
		if err != nil {
			return err
		}
		*dig = buf
		return nil

	})
}

// Address represents a 32-byte Sui address.
type Address [32]byte

type Digest []byte

type ObjectID = Address

// PersonalMessage represents a personal message to be signed.
type PersonalMessage struct {
	Message []byte
}

// ObjectRef represents a reference to a Sui object, including its ID, version, and digest.
type ObjectRef struct {
	ObjectID ObjectID `json:"objectId"`
	Version  uint64   `json:"version"`
	Digest   Digest   `json:"digest"`
}

// SharedObjectRef represents a reference to a shared Sui object.
type SharedObjectRef struct {
	ObjectID             ObjectID `json:"objectId"`
	InitialSharedVersion uint64   `json:"initialSharedVersion"`
	Mutable              bool     `json:"mutable"`
}

// String returns the hex-encoded string representation of the address, prefixed with "0x".
func (a Address) String() string {
	return "0x" + hex.EncodeToString(a[:])
}

// String returns the Base58-encoded string representation of the digest.
func (d Digest) String() string {
	return base58.Encode(d)
}

const digestLength = 32

func (a Address) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

func (a *Address) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*a = Address{}
		return nil
	}
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	parsed, err := parseAddressString(value)
	if err != nil {
		return err
	}
	*a = parsed
	return nil
}

func (d Digest) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Digest) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*d = nil
		return nil
	}
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	if value == "" {
		*d = nil
		return nil
	}
	decoded := base58.Decode(value)
	if len(decoded) != digestLength {
		return fmt.Errorf("invalid digest")
	}
	*d = append((*d)[:0], decoded...)
	return nil
}

func parseAddressString(input string) (Address, error) {
	trimmed := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(input)), "0x")
	if trimmed == "" {
		return Address{}, fmt.Errorf("invalid address")
	}
	if len(trimmed) > 64 {
		return Address{}, fmt.Errorf("invalid address")
	}
	padded := strings.Repeat("0", 64-len(trimmed)) + trimmed
	decoded, err := hex.DecodeString(padded)
	if err != nil {
		return Address{}, fmt.Errorf("invalid address")
	}
	if len(decoded) != len(Address{}) {
		return Address{}, fmt.Errorf("invalid address")
	}
	var addr Address
	copy(addr[:], decoded)
	return addr, nil
}
