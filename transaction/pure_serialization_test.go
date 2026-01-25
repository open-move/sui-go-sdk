package transaction

import (
	"context"
	"encoding/base64"
	"math/big"
	"testing"

	bcs "github.com/iotaledger/bcs-go"
)

func TestPureSerializationBCS(t *testing.T) {
	tx := New()

	tx.PureU8(1)
	tx.PureU16(1)
	tx.PureU32(1)
	tx.PureU64(1)
	tx.PureU128(big.NewInt(1))
	tx.PureU256(big.NewInt(1))
	tx.PureBool(true)
	tx.PureString("foo")
	tx.PureAddress("0x2")
	tx.PureAddress("0x2")

	vector := []byte{1, 2, 3}
	bytes, err := bcs.Marshal(&vector)
	if err != nil {
		t.Fatalf("marshal vector<u8>: %v", err)
	}
	tx.PureBytes(bytes)

	optSome := bcs.Option[uint8]{Some: 1}
	bytes, err = bcs.Marshal(&optSome)
	if err != nil {
		t.Fatalf("marshal option<u8> some: %v", err)
	}
	tx.PureBytes(bytes)

	optNone := bcs.Option[uint8]{None: true}
	bytes, err = bcs.Marshal(&optNone)
	if err != nil {
		t.Fatalf("marshal option<u8> none: %v", err)
	}
	tx.PureBytes(bytes)

	nested := bcs.Option[[][]bcs.Option[uint8]]{
		Some: [][]bcs.Option[uint8]{
			{{Some: 1}, {None: true}, {Some: 3}},
			{{Some: 4}, {None: true}, {Some: 6}},
		},
	}
	bytes, err = bcs.Marshal(&nested)
	if err != nil {
		t.Fatalf("marshal option<vector<vector<option<u8>>>>: %v", err)
	}
	tx.PureBytes(bytes)

	result, err := tx.Build(context.Background(), BuildOptions{})
	if err != nil {
		t.Fatalf("build transaction: %v", err)
	}

	expected := []string{
		"AQ==",
		"AQA=",
		"AQAAAA==",
		"AQAAAAAAAAA=",
		"AQAAAAAAAAAAAAAAAAAAAA==",
		"AQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		"AQ==",
		"A2Zvbw==",
		"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAI=",
		"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAI=",
		"AwECAw==",
		"AQE=",
		"AA==",
		"AQIDAQEAAQMDAQQAAQY=",
	}

	if len(result.ResolvedInputArgs) != len(expected) {
		t.Fatalf("expected %d inputs, got %d", len(expected), len(result.ResolvedInputArgs))
	}

	for i, arg := range result.ResolvedInputArgs {
		if arg.Pure == nil {
			t.Fatalf("input %d missing pure bytes", i)
		}
		encoded := base64.StdEncoding.EncodeToString(arg.Pure.Bytes)
		if encoded != expected[i] {
			t.Fatalf("input %d bytes mismatch: %s", i, encoded)
		}
	}
}
