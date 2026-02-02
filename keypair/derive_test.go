package keypair

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/open-move/sui-go-sdk/keychain"
)

const testMnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

func TestParseAndValidatePath(t *testing.T) {
	path, err := keychain.ParseDerivationPath("m/44'/784'/0'/0'/0'")
	if err != nil {
		t.Fatalf("parse path: %v", err)
	}
	if err = path.ValidateForScheme(keychain.SchemeEd25519); err != nil {
		t.Fatalf("validate ed25519 path: %v", err)
	}
	if err = path.ValidateForScheme(keychain.SchemeSecp256k1); err == nil {
		t.Fatalf("expected validation failure for secp256k1 hardened change")
	}

	path2, err := keychain.ParseDerivationPath("m/54'/784'/0'/0/0")
	if err != nil {
		t.Fatalf("parse path2: %v", err)
	}
	if err := path2.ValidateForScheme(keychain.SchemeSecp256k1); err != nil {
		t.Fatalf("validate secp256k1 path: %v", err)
	}
	if err := path2.ValidateForScheme(keychain.SchemeEd25519); err == nil {
		t.Fatalf("expected validation failure for ed25519 non-hardened change")
	}
}

func TestDeriveFromMnemonicLegacyVectors(t *testing.T) {
	cases := []struct {
		name     string
		scheme   keychain.Scheme
		path     string
		wantPriv string
		wantPub  string
		wantAddr string
	}{
		{
			name:     "ed25519",
			scheme:   keychain.SchemeEd25519,
			path:     "m/44'/784'/0'/0'/0'",
			wantPriv: "8869cb07178bf67e08d7c4abdf45487dbf379c9a452fcec2836854bf4a3d29b0900b4d81eecea3df2f74b14200c4f4cf3f49afaca7a634ffd2cf6ff82bdaecf2",
			wantPub:  "900b4d81eecea3df2f74b14200c4f4cf3f49afaca7a634ffd2cf6ff82bdaecf2",
			wantAddr: "0x5e93a736d04fbb25737aa40bee40171ef79f65fae833749e3c089fe7cc2161f1",
		},
		{
			name:     "secp256k1",
			scheme:   keychain.SchemeSecp256k1,
			path:     "m/54'/784'/0'/0/0",
			wantPriv: "0eacf0e4e0835692d7cd1a7c2eea8c1cfa10d3000414d31978e7b6ca657d0684",
			wantPub:  "02623d860f46cce9117d3f1ac382b79c59928a004a1986561a99df2a85167cf585",
			wantAddr: "0xc61a7f1161020a717f852dca2e9bfc1ffe235145406dfbdccc16e6907c1f5403",
		},
		{
			name:     "secp256r1",
			scheme:   keychain.SchemeSecp256r1,
			path:     "m/74'/784'/0'/0/0",
			wantPriv: "4877178c47f3c7d7fae6e3cd37a85c1a821f64818e31d41f24a65e6f78446ad1",
			wantPub:  "034019bca8a878458a63e5bf53f30855e31070f7b57cf9dcf265c98bdb17bb17c4",
			wantAddr: "0x0c0f9f53f2ad697e18279dfadefdd070c8e99416309d3ce614086c0860db6bb4",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			kp, err := DeriveFromMnemonic(tc.scheme, testMnemonic, "", tc.path)

			if err != nil {
				t.Fatalf("derive keypair: %v", err)
			}

			secret := fmtBytes(requireSecretBytes(t, kp))
			expectedSecret := tc.wantPriv
			if len(expectedSecret) > keychain.PrivateKeySize()*2 {
				expectedSecret = expectedSecret[:keychain.PrivateKeySize()*2]
			}

			if secret != expectedSecret {
				t.Fatalf("secret mismatch: got %s want %s", secret, expectedSecret)
			}

			if got := fmtBytes(kp.PublicKey()); got != tc.wantPub {
				t.Fatalf("public mismatch: got %s want %s", got, tc.wantPub)
			}

			addr, err := kp.SuiAddress()
			if err != nil {
				t.Fatalf("address: %v", err)
			}

			if addr != tc.wantAddr {
				t.Fatalf("address mismatch: got %s want %s", addr, tc.wantAddr)
			}
		})
	}
}

func fmtBytes(b []byte) string {
	return fmt.Sprintf("%x", b)
}

func requireSecretBytes(t *testing.T, kp Keypair) []byte {
	t.Helper()

	secret, err := kp.ExportSecret()
	if err != nil {
		t.Fatalf("export secret: %v", err)
	}
	return secret
}

func TestGenerateAndBech32RoundTrip(t *testing.T) {
	schemes := []keychain.Scheme{
		keychain.SchemeEd25519,
		keychain.SchemeSecp256k1,
		keychain.SchemeSecp256r1,
	}

	for _, scheme := range schemes {
		t.Run(fmt.Sprintf("scheme_%d", scheme), func(t *testing.T) {
			kp, err := Generate(scheme)
			if err != nil {
				t.Fatalf("generate: %v", err)
			}
			encoded, err := ToBech32FromKeypair(kp)
			if err != nil {
				t.Fatalf("to bech32: %v", err)
			}
			parsed, err := FromBech32(encoded)
			if err != nil {
				t.Fatalf("from bech32: %v", err)
			}
			if parsed.Scheme() != scheme {
				t.Fatalf("scheme mismatch: got %v want %v", parsed.Scheme(), scheme)
			}
			addr1, err := kp.SuiAddress()
			if err != nil {
				t.Fatalf("address original: %v", err)
			}
			addr2, err := parsed.SuiAddress()
			if err != nil {
				t.Fatalf("address parsed: %v", err)
			}
			if addr1 != addr2 {
				t.Fatalf("address mismatch after decode: %s vs %s", addr1, addr2)
			}
		})
	}
}

func TestFromSecretKeyValidation(t *testing.T) {
	zero := make([]byte, keychain.PrivateKeySize())
	if _, err := FromSecretKey(keychain.SchemeSecp256k1, zero); err == nil {
		t.Fatalf("expected secp256k1 zero secret to fail")
	}

	over := make([]byte, keychain.PrivateKeySize())
	for i := range over {
		over[i] = 0xFF
	}
	if _, err := FromSecretKey(keychain.SchemeSecp256r1, over); err == nil {
		t.Fatalf("expected secp256r1 over-range secret to fail")
	}

	mnemonic := testMnemonic
	kp, err := DeriveFromMnemonic(keychain.SchemeSecp256k1, mnemonic, "", "m/54'/784'/0'/0/0")
	if err != nil {
		t.Fatalf("derive: %v", err)
	}
	secret := requireSecretBytes(t, kp)
	imported, err := FromSecretKey(keychain.SchemeSecp256k1, secret)
	if err != nil {
		t.Fatalf("import valid secret: %v", err)
	}
	if fmtBytes(imported.PublicKey()) != fmtBytes(kp.PublicKey()) {
		t.Fatalf("public key mismatch after import")
	}
}

func TestDeriveFromMnemonicKeytoolVectors(t *testing.T) {
	tests := []struct {
		name     string
		scheme   keychain.Scheme
		mnemonic string
		path     string
		wantAddr string
		wantPub  string
	}{
		{
			name:     "ed25519",
			scheme:   keychain.SchemeEd25519,
			mnemonic: "ship host undo vacant also squeeze current alarm shift blush travel supply",
			path:     "m/44'/784'/0'/0'/0'",
			wantAddr: "0xd7486019a92937436adb2fd887f7b649048077ca9b72896a4c37aea054f6b2a9",
			wantPub:  "AJuWUnqJbqR9oTKZasbF8ixevzRFNR5qtVrXmRQz1Lsk",
		},
		{
			name:     "secp256k1",
			scheme:   keychain.SchemeSecp256k1,
			mnemonic: "decline core depend top judge surprise paper vacant caution smoke gospel year",
			path:     "m/54'/784'/0'/0/0",
			wantAddr: "0x3cc218c4e351f0a55f06bc2f0786cb67add024ceac8a511b1e879c118ac87f29",
			wantPub:  "AQLOL/oVGeK0NT62tHKj/ro2IcE7jxi1nW5z5FfpvN5wig==",
		},
		{
			name:     "secp256r1",
			scheme:   keychain.SchemeSecp256r1,
			mnemonic: "neutral cargo public impulse smile lock duck ignore car such remain pattern",
			path:     "m/74'/784'/0'/0/0",
			wantAddr: "0xd531b6148eb0e525d8118bb884964e577768880f9e0b6f194c14738e16550957",
			wantPub:  "AgK9N/mo6OoajCtIysGVMTSCvlakUQiAfmLSL6R42V9GUw==",
		},
		{
			name:     "secp256r1_alt_seed",
			scheme:   keychain.SchemeSecp256r1,
			mnemonic: "gorilla sorry also canal tuition sketch someone gym wonder cave midnight scout",
			path:     "m/74'/784'/0'/0/0",
			wantAddr: "0x15ec026d0681416b4f44ef9e4ccd78f018f3e650cc20d181de1d48a96541327d",
			wantPub:  "AgNoxvjrrvJUobKm3yXVVf2rndPAbcZ/jYH4IgIdM+C39A==",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kp, err := DeriveFromMnemonic(tc.scheme, tc.mnemonic, "", tc.path)
			if err != nil {
				t.Fatalf("derive: %v", err)
			}

			addr, err := kp.SuiAddress()
			if err != nil {
				t.Fatalf("address: %v", err)
			}

			gotPub := PublicKeyBase64(kp.Scheme(), kp.PublicKey())
			if gotPub != tc.wantPub {
				t.Fatalf("pubkey mismatch: got %s want %s", gotPub, tc.wantPub)
			}
			if addr != tc.wantAddr {
				t.Fatalf("address mismatch: got %s want %s", addr, tc.wantAddr)
			}
		})
	}
}

func TestFromBech32(t *testing.T) {
	const encoded = "suiprivkey1qzqgujqx9qh9kapmdlg9nywns9qtxy7my2r575zkpcyzzeu7x5672elhd4v"
	const wantAddress = "0xc888ef48f05c40e3c68940e7fc2c7664b8584c4b175d49fc6b25fe81ceba974c"

	kp, err := FromBech32(encoded)
	if err != nil {
		t.Fatalf("from bech32: %v", err)
	}

	if kp.Scheme() != keychain.SchemeEd25519 {
		t.Fatalf("scheme mismatch: got %v want %v", kp.Scheme(), keychain.SchemeEd25519)
	}

	addr, err := kp.SuiAddress()
	if err != nil {
		t.Fatalf("address: %v", err)
	}

	if addr != wantAddress {
		t.Fatalf("address mismatch: got %s want %s", addr, wantAddress)
	}
}

func TestDeriveFromMnemonicVectors(t *testing.T) {
	tests := []struct {
		name        string
		scheme      keychain.Scheme
		mnemonic    string
		path        string
		wantAddr    string
		wantPubFlag string
		wantPubRaw  string
	}{
		{
			name:        "ed25519_index1",
			scheme:      keychain.SchemeEd25519,
			mnemonic:    "ship host undo vacant also squeeze current alarm shift blush travel supply",
			path:        "m/44'/784'/0'/0'/1'",
			wantAddr:    "0x9c79fa2ae665bba59b1533c2cb378c7d8462d59f6b30658465c765458b15e63e",
			wantPubFlag: "AJCtbjJjXl2IEynolldZeXLqe2wPVPCW/O1rnZheR8Rj",
		},
		{
			name:        "secp256k1_change1",
			scheme:      keychain.SchemeSecp256k1,
			mnemonic:    "decline core depend top judge surprise paper vacant caution smoke gospel year",
			path:        "m/54'/784'/0'/1/0",
			wantAddr:    "0x8b280584e76d1f6783a9051110be6156a876d185f53a758a976edf9d5c17f13b",
			wantPubFlag: "AQKVT35cLt7dcQD6tM023nR2q4e3IqUhQ6rYC2bRyUOxCg==",
		},
		{
			name:        "secp256r1_index1",
			scheme:      keychain.SchemeSecp256r1,
			mnemonic:    "neutral cargo public impulse smile lock duck ignore car such remain pattern",
			path:        "m/74'/784'/0'/0/1",
			wantAddr:    "0x15407b5322ab8a2e24787ecf1703286db4d2951c3f793a4fb89b0ef2b7f6b80e",
			wantPubFlag: "AgIzlaev1QM+UFS4kEYGqMAODU491KAnSD+Eaw0ZcL9fEw==",
		},
		{
			name:       "ts_ed25519_1",
			scheme:     keychain.SchemeEd25519,
			mnemonic:   "film crazy soon outside stand loop subway crumble thrive popular green nuclear struggle pistol arm wife phrase warfare march wheat nephew ask sunny firm",
			path:       "m/44'/784'/0'/0'/0'",
			wantAddr:   "0xa2d14fad60c56049ecf75246a481934691214ce413e6a8ae2fe6834c173a6133",
			wantPubRaw: "ImR/7u82MGC9QgWhZxoV8QoSNnZZGLG19jjYLzPPxGk=",
		},
		{
			name:       "ts_ed25519_2",
			scheme:     keychain.SchemeEd25519,
			mnemonic:   "require decline left thought grid priority false tiny gasp angle royal system attack beef setup reward aunt skill wasp tray vital bounce inflict level",
			path:       "m/44'/784'/0'/0'/0'",
			wantAddr:   "0x1ada6e6f3f3e4055096f606c746690f1108fcc2ca479055cc434a3e1d3f758aa",
			wantPubRaw: "vG6hEnkYNIpdmWa/WaLivd1FWBkxG+HfhXkyWgs9uP4=",
		},
		{
			name:       "ts_ed25519_3",
			scheme:     keychain.SchemeEd25519,
			mnemonic:   "organ crash swim stick traffic remember army arctic mesh slice swear summer police vast chaos cradle squirrel hood useless evidence pet hub soap lake",
			path:       "m/44'/784'/0'/0'/0'",
			wantAddr:   "0xe69e896ca10f5a77732769803cc2b5707f0ab9d4407afb5e4b4464b89769af14",
			wantPubRaw: "arEzeF7Uu90jP4Sd+Or17c+A9kYviJpCEQAbEt0FHbU=",
		},
		{
			name:       "ts_secp256k1_1",
			scheme:     keychain.SchemeSecp256k1,
			mnemonic:   "film crazy soon outside stand loop subway crumble thrive popular green nuclear struggle pistol arm wife phrase warfare march wheat nephew ask sunny firm",
			path:       "m/54'/784'/0'/0/0",
			wantAddr:   "0x9e8f732575cc5386f8df3c784cd3ed1b53ce538da79926b2ad54dcc1197d2532",
			wantPubRaw: "Ar2Vs2ei2HgaCIvcsAVAZ6bKYXhDfRTlF432p8Wn4lsL",
		},
		{
			name:       "ts_secp256k1_2",
			scheme:     keychain.SchemeSecp256k1,
			mnemonic:   "require decline left thought grid priority false tiny gasp angle royal system attack beef setup reward aunt skill wasp tray vital bounce inflict level",
			path:       "m/54'/784'/0'/0/0",
			wantAddr:   "0x9fd5a804ed6b46d36949ff7434247f0fd594673973ece24aede6b86a7b5dae01",
			wantPubRaw: "A5IcrmWDxl0J/4MNkrtE1AvwiLZiqih9tjttcGlafw+m",
		},
		{
			name:       "ts_secp256k1_3",
			scheme:     keychain.SchemeSecp256k1,
			mnemonic:   "organ crash swim stick traffic remember army arctic mesh slice swear summer police vast chaos cradle squirrel hood useless evidence pet hub soap lake",
			path:       "m/54'/784'/0'/0/0",
			wantAddr:   "0x60287d7c38dee783c2ab1077216124011774be6b0764d62bd05f32c88979d5c5",
			wantPubRaw: "AuEiECTZwyHhqStzpO/RNBXO89/Wa8oc4BtoneKnl6h8",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kp, err := DeriveFromMnemonic(tc.scheme, tc.mnemonic, "", tc.path)
			if err != nil {
				t.Fatalf("derive: %v", err)
			}

			addr, err := kp.SuiAddress()
			if err != nil {
				t.Fatalf("address: %v", err)
			}
			if addr != tc.wantAddr {
				t.Fatalf("addr mismatch: got %s want %s", addr, tc.wantAddr)
			}

			if tc.wantPubFlag != "" {
				if got := PublicKeyBase64(kp.Scheme(), kp.PublicKey()); got != tc.wantPubFlag {
					t.Fatalf("flagged pub mismatch: got %s want %s", got, tc.wantPubFlag)
				}
			}
			if tc.wantPubRaw != "" {
				if got := base64.StdEncoding.EncodeToString(kp.PublicKey()); got != tc.wantPubRaw {
					t.Fatalf("raw pub mismatch: got %s want %s", got, tc.wantPubRaw)
				}
			}
		})
	}
}

func TestDeriveFromMnemonicWithPassphrase(t *testing.T) {
	tests := []struct {
		name       string
		scheme     keychain.Scheme
		mnemonic   string
		passphrase string
		path       string
		wantAddr   string
		wantPub    string
	}{
		{
			name:       "ed25519_passphrase",
			scheme:     keychain.SchemeEd25519,
			mnemonic:   "ship host undo vacant also squeeze current alarm shift blush travel supply",
			passphrase: "mypass",
			path:       "m/44'/784'/0'/0'/0'",
			wantAddr:   "0x9ca71a559d588ff6d1680fa73e1bf77ee7d983b12bcb0b4eb72c810d758c4a8e",
			wantPub:    "AN2stTh0iScicZZh1OH2RBaLYh2VgkgYM1rQxIV55tbg",
		},
		{
			name:       "secp256k1_passphrase",
			scheme:     keychain.SchemeSecp256k1,
			mnemonic:   "decline core depend top judge surprise paper vacant caution smoke gospel year",
			passphrase: "mypass",
			path:       "m/54'/784'/0'/0/0",
			wantAddr:   "0xb6edf98a60a4d80d9c4dab1d1f33b3c7e1a15b679fa8011dba0e704b5dfeba89",
			wantPub:    "AQMdDQb8HD+exTyQR6nTqFimnGpYKnNYo9e4SLWK64Bd8A==",
		},
		{
			name:       "secp256r1_passphrase",
			scheme:     keychain.SchemeSecp256r1,
			mnemonic:   "neutral cargo public impulse smile lock duck ignore car such remain pattern",
			passphrase: "mypass",
			path:       "m/74'/784'/0'/0/0",
			wantAddr:   "0x7bc6917b2f37a7d1bf3f917009aa278dcac9f66125a66daddf2c7bee7c07c157",
			wantPub:    "AgPqmeCif5aT0Ou8Ml8gPc6ziXTsprFyGzDeIhwjGWm5/Q==",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kp, err := DeriveFromMnemonic(tc.scheme, tc.mnemonic, tc.passphrase, tc.path)
			if err != nil {
				t.Fatalf("derive: %v", err)
			}
			addr, err := kp.SuiAddress()
			if err != nil {
				t.Fatalf("addr: %v", err)
			}
			if addr != tc.wantAddr {
				t.Fatalf("addr mismatch: got %s want %s", addr, tc.wantAddr)
			}
			if got := PublicKeyBase64(kp.Scheme(), kp.PublicKey()); got != tc.wantPub {
				t.Fatalf("pub mismatch: got %s want %s", got, tc.wantPub)
			}
		})
	}
}

func TestDeriveFromMnemonicInvalidPaths(t *testing.T) {
	invalid := []struct {
		name   string
		scheme keychain.Scheme
		path   string
	}{
		{
			name:   "ed25519_unhardened",
			scheme: keychain.SchemeEd25519,
			path:   "m/44'/784'/0'/0/0",
		},
		{
			name:   "secp256k1_hardened_change",
			scheme: keychain.SchemeSecp256k1,
			path:   "m/54'/784'/0'/0'/0",
		},
		{
			name:   "secp256r1_hardened_change",
			scheme: keychain.SchemeSecp256r1,
			path:   "m/74'/784'/0'/0'/0",
		},
	}

	for _, tc := range invalid {
		t.Run(tc.name, func(t *testing.T) {
			_, err := DeriveFromMnemonic(tc.scheme, testMnemonic, "", tc.path)
			if err == nil {
				t.Fatalf("expected error for %s", tc.path)
			}
		})
	}
}
