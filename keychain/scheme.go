package keychain

import (
	"fmt"
	"reflect"
	"strings"
)

// Scheme represents a cryptographic signature scheme supported by Sui.
type Scheme uint8

const (
	// SchemeEd25519 represents the Ed25519 signature scheme.
	SchemeEd25519 Scheme = iota
	// SchemeSecp256k1 represents the Secp256k1 signature scheme.
	SchemeSecp256k1
	// SchemeSecp256r1 represents the Secp256r1 signature scheme.
	SchemeSecp256r1
)

const (
	flagUnspecified byte = 0xff
)

// AddressFlag returns the single-byte flag used for address derivation for the scheme.
func (s Scheme) AddressFlag() byte {
	switch s {
	case SchemeEd25519:
		return 0x00
	case SchemeSecp256k1:
		return 0x01
	case SchemeSecp256r1:
		return 0x02
	default:
		return flagUnspecified
	}
}

func (s Scheme) Purpose() uint32 {
	switch s {
	case SchemeEd25519:
		return 44
	case SchemeSecp256k1:
		return 54
	case SchemeSecp256r1:
		return 74
	default:
		return 0
	}
}

func (scheme Scheme) Label() string {
	switch scheme {
	case SchemeEd25519:
		return "ed25519"
	case SchemeSecp256k1:
		return "secp256k1"
	case SchemeSecp256r1:
		return "secp256r1"
	default:
		return strings.ToLower((reflect.TypeOf(scheme).Name()))
	}
}

// SchemeFromFlag returns the Scheme corresponding to the given flag byte.
func SchemeFromFlag(flag byte) (Scheme, error) {
	switch flag {
	case 0x00:
		return SchemeEd25519, nil
	case 0x01:
		return SchemeSecp256k1, nil
	case 0x02:
		return SchemeSecp256r1, nil
	default:
		return 0, fmt.Errorf("unknown scheme flag 0x%02x", flag)
	}
}
