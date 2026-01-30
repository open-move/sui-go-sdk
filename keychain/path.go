package keychain

import (
	"fmt"
	"strconv"
	"strings"
)

const suiCoinType uint32 = 784

// PathSegment represents a single segment in a BIP-32 derivation path.
type PathSegment struct {
	Index    uint32
	Hardened bool
}

// HardenedIndex returns the index with the hardened bit set if the segment is hardened.
func (s PathSegment) HardenedIndex() uint32 {
	if s.Hardened {
		return s.Index | 0x80000000
	}

	return s.Index
}

// DerivationPath represents a full BIP-32 derivation path (e.g., m/44'/784'/0'/0'/0').
type DerivationPath struct {
	segments []PathSegment
}

// Segments returns a copy of the path segments.
func (p DerivationPath) Segments() []PathSegment {
	out := make([]PathSegment, len(p.segments))
	copy(out, p.segments)
	return out
}

// String returns the string representation of the derivation path (e.g., "m/44'/784'/0'/0'/0'").
func (p DerivationPath) String() string {
	parts := make([]string, len(p.segments))
	for i, seg := range p.segments {
		suffix := ""
		if seg.Hardened {
			suffix = "'"
		}
		parts[i] = fmt.Sprintf("%d%s", seg.Index, suffix)
	}
	return "m/" + strings.Join(parts, "/")
}

func ParseDerivationPath(raw string) (DerivationPath, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return DerivationPath{}, fmt.Errorf("path: empty path")
	}
	if !strings.HasPrefix(raw, "m/") && raw != "m" {
		return DerivationPath{}, fmt.Errorf("path: must start with m/")
	}
	if raw == "m" {
		return DerivationPath{segments: nil}, nil
	}
	body := strings.TrimPrefix(raw, "m/")
	parts := strings.Split(body, "/")
	segments := make([]PathSegment, len(parts))
	for i, part := range parts {
		if part == "" {
			return DerivationPath{}, fmt.Errorf("path: empty segment at position %d", i)
		}
		hardened := strings.HasSuffix(part, "'")
		if hardened {
			part = strings.TrimSuffix(part, "'")
		}
		idx, err := strconv.ParseUint(part, 10, 32)
		if err != nil {
			return DerivationPath{}, fmt.Errorf("path: invalid segment %q: %w", part, err)
		}
		segments[i] = PathSegment{Index: uint32(idx), Hardened: hardened}
	}
	return DerivationPath{segments: segments}, nil
}

// ValidateForScheme checks if the path structure adheres to the requirements of the given signature scheme.
func (p DerivationPath) ValidateForScheme(s Scheme) error {
	if len(p.segments) < 5 {
		return fmt.Errorf("path: expected at least 5 segments, got %d", len(p.segments))
	}

	purpose := p.segments[0]
	if !purpose.Hardened || purpose.Index != s.Purpose() {
		return fmt.Errorf("path: invalid purpose segment %v for scheme %d", purpose, s)
	}

	coin := p.segments[1]
	if !coin.Hardened || coin.Index != suiCoinType {
		return fmt.Errorf("path: invalid coin type %v", coin)
	}

	account := p.segments[2]
	if !account.Hardened {
		return fmt.Errorf("path: account segment must be hardened")
	}
	change := p.segments[3]
	address := p.segments[4]

	switch s {
	case SchemeEd25519:
		if !change.Hardened || !address.Hardened {
			return fmt.Errorf("path: ed25519 requires hardened change and address")
		}
	case SchemeSecp256k1, SchemeSecp256r1:
		if change.Hardened || address.Hardened {
			return fmt.Errorf("path: ecdsa schemes require non-hardened change and address")
		}
	default:
		return fmt.Errorf("path: unsupported scheme %d", s)
	}
	return nil
}
