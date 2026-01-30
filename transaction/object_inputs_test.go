package transaction

import (
	"bytes"
	"context"
	"encoding/base64"
	"testing"

	"github.com/open-move/sui-go-sdk/types"
	"github.com/open-move/sui-go-sdk/utils"
)

func TestObjectInputKindBCS(t *testing.T) {
	digest := types.Digest(bytes.Repeat([]byte{1}, 32))

	tx := New()
	tx.MoveCall(MoveCall{
		Target: "0x2::foo::bar",
		Arguments: []Argument{
			tx.ReceivingObject(types.ObjectRef{
				ObjectID: mustAddress(t, "0x1"),
				Version:  123,
				Digest:   digest,
			}),
			tx.SharedObject(types.SharedObjectRef{
				ObjectID:             mustAddress(t, "0x2"),
				InitialSharedVersion: 123,
				Mutable:              true,
			}),
			tx.ObjectRef(types.ObjectRef{
				ObjectID: mustAddress(t, "0x3"),
				Version:  123,
				Digest:   digest,
			}),
			tx.PureAddress("0x2"),
		},
	})

	result, err := tx.Build(context.Background(), BuildOptions{})
	if err != nil {
		t.Fatalf("build transaction: %v", err)
	}
	if len(result.KindBytes) == 0 {
		t.Fatalf("expected kind bytes to be populated")
	}

	encoded := base64.StdEncoding.EncodeToString(result.KindBytes)
	const expected = "AAQBAgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABewAAAAAAAAAgAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACewAAAAAAAAABAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA3sAAAAAAAAAIAEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBACAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAgEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAIDZm9vA2JhcgAEAQAAAQEAAQIAAQMA"
	if encoded != expected {
		t.Fatalf("kind bytes mismatch: %s", encoded)
	}
}

func mustAddress(t *testing.T, value string) types.Address {
	t.Helper()
	addr, err := utils.ParseAddress(value)
	if err != nil {
		t.Fatalf("parse address %q: %v", value, err)
	}
	return addr
}
