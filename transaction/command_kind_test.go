package transaction

import (
	"bytes"
	"context"
	"encoding/base64"
	"testing"

	"github.com/open-move/sui-go-sdk/types"
)

func TestSplitCoinsKindBCS(t *testing.T) {
	tx := New()
	tx.SplitCoins(SplitCoins{
		Coin:    tx.Gas(),
		Amounts: []Argument{tx.PureU64(1000)},
	})

	result, err := tx.Build(context.Background(), BuildOptions{})
	if err != nil {
		t.Fatalf("build split coins: %v", err)
	}

	assertKindBytes(t, result.KindBytes, "AAEACOgDAAAAAAAAAQIAAQEAAA==")
}

func TestMergeCoinsKindBCS(t *testing.T) {
	digest := types.Digest(bytes.Repeat([]byte{1}, 32))

	tx := New()
	tx.MergeCoins(MergeCoins{
		Destination: tx.ObjectRef(types.ObjectRef{
			ObjectID: mustAddress(t, "0x1"),
			Version:  123,
			Digest:   digest,
		}),
		Sources: []Argument{
			tx.ObjectRef(types.ObjectRef{
				ObjectID: mustAddress(t, "0x2"),
				Version:  123,
				Digest:   digest,
			}),
		},
	})

	result, err := tx.Build(context.Background(), BuildOptions{})
	if err != nil {
		t.Fatalf("build merge coins: %v", err)
	}

	assertKindBytes(t, result.KindBytes, "AAIBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABewAAAAAAAAAgAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACewAAAAAAAAAgAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAwEAAAEBAQA=")
}

func TestTransferObjectsKindBCS(t *testing.T) {
	digest := types.Digest(bytes.Repeat([]byte{1}, 32))

	tx := New()
	tx.TransferObjects(TransferObjects{
		Objects: []Argument{
			tx.ObjectRef(types.ObjectRef{
				ObjectID: mustAddress(t, "0x1"),
				Version:  123,
				Digest:   digest,
			}),
			tx.ObjectRef(types.ObjectRef{
				ObjectID: mustAddress(t, "0x2"),
				Version:  123,
				Digest:   digest,
			}),
		},
		Address: tx.PureAddress("0x2"),
	})

	result, err := tx.Build(context.Background(), BuildOptions{})
	if err != nil {
		t.Fatalf("build transfer objects: %v", err)
	}

	assertKindBytes(t, result.KindBytes, "AAMBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABewAAAAAAAAAgAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACewAAAAAAAAAgAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEAIAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACAQECAQAAAQEAAQIA")
}

func TestMakeMoveVecKindBCS(t *testing.T) {
	tx := New()
	tx.MakeMoveVec(MakeMoveVecInput{
		Elements: []Argument{tx.PureU8(1), tx.PureU8(2)},
	})

	result, err := tx.Build(context.Background(), BuildOptions{})
	if err != nil {
		t.Fatalf("build make move vec: %v", err)
	}

	assertKindBytes(t, result.KindBytes, "AAIAAQEAAQIBBQACAQAAAQEA")
}

func TestMakeMoveVecTypedKindBCS(t *testing.T) {
	typeTag := "u8"
	tx := New()
	tx.MakeMoveVec(MakeMoveVecInput{
		Type:     &typeTag,
		Elements: []Argument{tx.PureU8(1), tx.PureU8(2)},
	})

	result, err := tx.Build(context.Background(), BuildOptions{})
	if err != nil {
		t.Fatalf("build make move vec typed: %v", err)
	}

	assertKindBytes(t, result.KindBytes, "AAIAAQEAAQIBBQEBAgEAAAEBAA==")
}

func assertKindBytes(t *testing.T, bytes []byte, expected string) {
	t.Helper()
	if len(bytes) == 0 {
		t.Fatalf("expected kind bytes to be populated")
	}
	encoded := base64.StdEncoding.EncodeToString(bytes)
	if encoded != expected {
		t.Fatalf("kind bytes mismatch: %s", encoded)
	}
}
