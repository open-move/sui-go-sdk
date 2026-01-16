package intent

import (
	"bytes"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/iotaledger/bcs-go"
	"github.com/open-move/sui-go-sdk/types"
	"golang.org/x/crypto/blake2b"
)

func TestIntentMessageSerializationWithPersonalMessage(t *testing.T) {
	message := types.PersonalMessage{Message: []byte("Hello")}
	bcsMessage, err := bcs.Marshal(&message)
	if err != nil {
		t.Fatalf("marshal personal message: %v", err)
	}

	intent := DefaultIntent().WithScope(IntentScopePersonalMessage)
	intentMessage := NewIntentMessage(intent, message)

	serialized, err := intentMessage.MarshalBCS()
	if err != nil {
		t.Fatalf("marshal intent message: %v", err)
	}

	if len(serialized) != len(bcsMessage)+3 {
		t.Fatalf("unexpected serialized length: got %d want %d", len(serialized), len(bcsMessage)+3)
	}

	expectedPrefix := []byte{
		byte(IntentScopePersonalMessage),
		byte(IntentVersionV0),
		byte(AppIDSui),
	}
	if got := serialized[:3]; got[0] != expectedPrefix[0] || got[1] != expectedPrefix[1] || got[2] != expectedPrefix[2] {
		t.Fatalf("unexpected prefix: got %v want %v", got, expectedPrefix)
	}

	if !bytes.Equal(serialized[3:], bcsMessage) {
		t.Fatalf("unexpected payload: got %x want %x", serialized[3:], bcsMessage)
	}

	digest, err := HashIntentMessage(intentMessage)
	if err != nil {
		t.Fatalf("hash intent message: %v", err)
	}
	expectedDigest := blake2b.Sum256(serialized)
	if digest != expectedDigest {
		t.Fatalf("unexpected digest: got %x want %x", digest, expectedDigest)
	}
}

func TestParseIntentRoundTrip(t *testing.T) {
	intent := DefaultIntent().WithScope(IntentScopeProofOfPossession)
	raw := intent.Bytes()
	hexEncoded := hex.EncodeToString(raw[:])

	parsed, err := ParseIntent(hexEncoded)
	if err != nil {
		t.Fatalf("parse intent: %v", err)
	}

	if parsed != intent {
		t.Fatalf("round-trip mismatch: got %+v want %+v", parsed, intent)
	}
}

func TestIntentValidationErrors(t *testing.T) {
	if _, err := IntentFromBytes([]byte{0x01, 0x02}); !errors.Is(err, errInvalidIntentLength) {
		t.Fatalf("expected errInvalidIntentLength, got %v", err)
	}

	badScope := Intent{Scope: 255, Version: IntentVersionV0, AppID: AppIDSui}
	if _, err := NewIntentMessage(badScope, struct{}{}).MarshalBCS(); !errors.Is(err, errInvalidIntentScope) {
		t.Fatalf("expected errInvalidIntentScope, got %v", err)
	}

	badVersion := Intent{Scope: IntentScopeTransactionData, Version: 99, AppID: AppIDSui}
	if _, err := NewIntentMessage(badVersion, struct{}{}).MarshalBCS(); !errors.Is(err, errInvalidIntentVers) {
		t.Fatalf("expected errInvalidIntentVers, got %v", err)
	}

	badApp := Intent{Scope: IntentScopeTransactionData, Version: IntentVersionV0, AppID: 42}
	if _, err := NewIntentMessage(badApp, struct{}{}).MarshalBCS(); !errors.Is(err, errInvalidIntentAppID) {
		t.Fatalf("expected errInvalidIntentAppID, got %v", err)
	}
}
