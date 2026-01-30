package main

import (
	"encoding/hex"
	"fmt"
	"log"

	"github.com/open-move/sui-go-sdk/keychain"
	"github.com/open-move/sui-go-sdk/keypair"
)

const (
	ed25519Bech32 = "suiprivkey1qz6qzxye624vk8epr7c9j4flnxm5lze2e7y2pmxzm4qarny03lt8xavx8zj"
	ed25519Hex    = "b4011899d2aacb1f211fb059553f99b74f8b2acf88a0ecc2dd41d1cc8f8fd673"
	ed25519Addr   = "0x55d07fa035a02cf143f2bea4aa34bbac516560c9386a077af3fd30f169fc2cc2"

	secp256k1Bech32 = "suiprivkey1qyqr6yvxdqkh32ep4pk9caqvphmk9epn6rhkczcrhaeermsyvwsg783y9am"
	secp256k1Addr   = "0x9e8f732575cc5386f8df3c784cd3ed1b53ce538da79926b2ad54dcc1197d2532"

	secp256r1Bech32 = "suiprivkey1qgj6vet4rstf2p00j860xctkg4fyqqq5hxgu4mm0eg60fq787ujnqs5wc8q"
	secp256r1Addr   = "0x4a822457f1970468d38dae8e63fb60eefdaa497d74d781f581ea2d137ec36f3a"
)

const (
	edMnemonic = "film crazy soon outside stand loop subway crumble thrive popular green nuclear struggle pistol arm wife phrase warfare march wheat nephew ask sunny firm"
	edPath     = "m/44'/784'/0'/0'/0'"

	secp256k1Path = "m/54'/784'/0'/0/0"

	secp256r1Mnemonic = "act wing dilemma glory episode region allow mad tourist humble muffin oblige"
	secp256r1Path     = "m/74'/784'/0'/0/0"
)

func main() {
	kp, err := keypair.FromBech32(ed25519Bech32)
	if err != nil {
		log.Fatalf("from bech32: %v", err)
	}
	addr, err := kp.SuiAddress()
	if err != nil {
		log.Fatalf("ed25519 bech32 address: %v", err)
	}
	fmt.Printf("ed25519 bech32 address: %s (expected %s)\n", addr, ed25519Addr)
	fmt.Printf("ed25519 bech32 public key: %x\n", kp.PublicKey())
	fmt.Print("====================\n\n")

	seed, err := hex.DecodeString(ed25519Hex)
	if err != nil {
		log.Fatalf("decode hex seed: %v", err)
	}
	kp, err = keypair.FromSecretKey(keychain.SchemeEd25519, seed)
	if err != nil {
		log.Fatalf("from secret key: %v", err)
	}
	addr, err = kp.SuiAddress()
	if err != nil {
		log.Fatalf("ed25519 hex address: %v", err)
	}
	fmt.Printf("ed25519 hex address: %s (expected %s)\n", addr, ed25519Addr)
	fmt.Printf("ed25519 hex public key: %x\n", kp.PublicKey())
	fmt.Print("====================\n\n")

	kp, err = keypair.FromBech32(secp256k1Bech32)
	if err != nil {
		log.Fatalf("from bech32: %v", err)
	}
	addr, err = kp.SuiAddress()
	if err != nil {
		log.Fatalf("secp256k1 bech32 address: %v", err)
	}
	fmt.Printf("secp256k1 bech32 address: %s (expected %s)\n", addr, secp256k1Addr)
	fmt.Printf("secp256k1 bech32 public key: %x\n", kp.PublicKey())
	fmt.Print("====================\n\n")

	kp, err = keypair.FromBech32(secp256r1Bech32)
	if err != nil {
		log.Fatalf("from bech32: %v", err)
	}
	addr, err = kp.SuiAddress()
	if err != nil {
		log.Fatalf("secp256r1 bech32 address: %v", err)
	}
	fmt.Printf("secp256r1 bech32 address: %s (expected %s)\n", addr, secp256r1Addr)
	fmt.Printf("secp256r1 bech32 public key: %x\n", kp.PublicKey())
	fmt.Print("====================\n\n")

	kp, err = keypair.DeriveFromMnemonic(keychain.SchemeEd25519, edMnemonic, "", edPath)
	if err != nil {
		log.Fatalf("derive ed25519 mnemonic: %v", err)
	}
	addr, err = kp.SuiAddress()
	if err != nil {
		log.Fatalf("ed25519 mnemonic address: %v", err)
	}
	fmt.Printf("ed25519 mnemonic address: %s\n", addr)
	fmt.Printf("ed25519 mnemonic public key: %x\n", kp.PublicKey())
	fmt.Print("====================\n\n")

	kp, err = keypair.DeriveFromMnemonic(keychain.SchemeSecp256k1, edMnemonic, "", secp256k1Path)
	if err != nil {
		log.Fatalf("derive secp256k1 mnemonic: %v", err)
	}
	addr, err = kp.SuiAddress()
	if err != nil {
		log.Fatalf("secp256k1 mnemonic address: %v", err)
	}
	fmt.Printf("secp256k1 mnemonic address: %s\n", addr)
	fmt.Printf("secp256k1 mnemonic public key: %x\n", kp.PublicKey())
	fmt.Print("====================\n\n")

	kp, err = keypair.DeriveFromMnemonic(keychain.SchemeSecp256r1, secp256r1Mnemonic, "", secp256r1Path)
	if err != nil {
		log.Fatalf("derive secp256r1 mnemonic: %v", err)
	}
	addr, err = kp.SuiAddress()
	if err != nil {
		log.Fatalf("secp256r1 mnemonic address: %v", err)
	}
	fmt.Printf("secp256r1 mnemonic address: %s\n", addr)
	fmt.Printf("secp256r1 mnemonic public key: %x\n", kp.PublicKey())
}
