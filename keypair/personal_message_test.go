package keypair_test

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/open-move/sui-go-sdk/keychain"
	"github.com/open-move/sui-go-sdk/keypair"
)

type personalSigner interface {
	SignPersonalMessage([]byte) ([]byte, error)
	VerifyPersonalMessage([]byte, []byte) error
}

func TestPersonalMessageSignatures(t *testing.T) {
	message := []byte("hello world")
	encodedMessage := base64.RawStdEncoding.EncodeToString(message)

	cases := []struct {
		name       string
		factory    func() (personalSigner, error)
		flag       byte
		expectBase string
	}{
		{
			name: "ed25519",
			factory: func() (personalSigner, error) {
				return keypair.DeriveFromMnemonic(
					keychain.SchemeEd25519,
					"ship host undo vacant also squeeze current alarm shift blush travel supply",
					"",
					"m/44'/784'/0'/0'/0'",
				)
			},
			flag:       keychain.SchemeEd25519.AddressFlag(),
			expectBase: "AGkilBoAfNafse3UbVFd9vT6NQD5inAAAMDcPL/1OnqZbYQq9YcUgSaRMn37IRooehZJrTMZQZdPN77KHIL7OgabllJ6iW6kfaEymWrGxfIsXr80RTUearVa15kUM9S7JA==",
		},
		{
			name: "secp256k1",
			factory: func() (personalSigner, error) {
				return keypair.DeriveFromMnemonic(
					keychain.SchemeSecp256k1,
					"decline core depend top judge surprise paper vacant caution smoke gospel year",
					"",
					"m/54'/784'/0'/0/0",
				)
			},
			flag:       keychain.SchemeSecp256k1.AddressFlag(),
			expectBase: "Aal8FzuGOxcjh6ngkdjZA2MumWhhwEgsunhflSsWJVfICa2+nc/lfXsRJhGTjX0aE68iol9FnJ+MHRubBW2JcwMCzi/6FRnitDU+trRyo/66NiHBO48YtZ1uc+RX6bzecIo=",
		},
		{
			name: "secp256r1",
			factory: func() (personalSigner, error) {
				return keypair.DeriveFromMnemonic(
					keychain.SchemeSecp256r1,
					"neutral cargo public impulse smile lock duck ignore car such remain pattern",
					"",
					"m/74'/784'/0'/0/0",
				)
			},
			flag:       keychain.SchemeSecp256r1.AddressFlag(),
			expectBase: "Ao1TC+qvzKxP4HMHsIyYL3wUzaB2zXPdwQHLk0UK1ctiHLUlu/nPBHANrAR77oHaQiww9XGltTrjiDZJgaBiSLECvTf5qOjqGowrSMrBlTE0gr5WpFEIgH5i0i+keNlfRlM=",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			signer, err := tc.factory()
			if err != nil {
				t.Fatalf("%s: make keypair: %v", tc.name, err)
			}

			sig, err := signer.SignPersonalMessage(append([]byte(nil), message...))
			if err != nil {
				t.Fatalf("%s: sign personal message: %v", tc.name, err)
			}
			if len(sig) == 0 {
				t.Fatalf("%s: got empty signature", tc.name)
			}
			if sig[0] != tc.flag {
				t.Fatalf("%s: signature flag mismatch: got 0x%x want 0x%x", tc.name, sig[0], tc.flag)
			}

			encodedSig := base64.StdEncoding.EncodeToString(sig)
			fmt.Printf("%s personal message (base64): %s\n", tc.name, encodedMessage)
			fmt.Printf("%s signature (base64): %s\n", tc.name, encodedSig)
			if tc.expectBase != "" && encodedSig != tc.expectBase {
				t.Fatalf("%s: signature mismatch\n\tgot:  %s\n\twant: %s", tc.name, encodedSig, tc.expectBase)
			}

			if err := signer.VerifyPersonalMessage(message, sig); err != nil {
				t.Fatalf("%s: verify personal message: %v", tc.name, err)
			}

			tamperedSig := append([]byte(nil), sig...)
			tamperedSig[len(tamperedSig)-1] ^= 0x01
			if err := signer.VerifyPersonalMessage(message, tamperedSig); err == nil {
				t.Fatalf("%s: verify should fail for tampered signature", tc.name)
			}

			tamperedMsg := append([]byte(nil), message...)
			tamperedMsg[0] ^= 0x01
			if err := signer.VerifyPersonalMessage(tamperedMsg, sig); err == nil {
				t.Fatalf("%s: verify should fail for tampered message", tc.name)
			}
		})
	}
}
