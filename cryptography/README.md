# Cryptography Module

This module provides cryptographic primitives and utilities for the Sui blockchain. It supports key generation, signing, and verification using supported elliptic curves.

## Supported Curves

The SDK supports the following signature schemes:

*   **Ed25519**: The default and recommended signature scheme for Sui.
*   **Secp256k1**: widely used in Bitcoin and Ethereum.
*   **Secp256r1** (NIST P-256): widely supported by hardware enclaves (e.g., Apple Secure Enclave, Android Keystore).

## Usage

### Generating Keypairs

```go
package main

import (
	"fmt"
	"log"

	"github.com/open-move/sui-go-sdk/cryptography/ed25519"
	"github.com/open-move/sui-go-sdk/cryptography/secp256k1"
	"github.com/open-move/sui-go-sdk/cryptography/secp256r1"
)

func main() {
	// Ed25519
	edKp, _ := ed25519.Generate()
	fmt.Printf("Ed25519 Address: %s\n", edKp.SuiAddress())

	// Secp256k1
	k1Kp, _ := secp256k1.Generate()
	fmt.Printf("Secp256k1 Address: %s\n", k1Kp.SuiAddress())

	// Secp256r1
	r1Kp, _ := secp256r1.Generate()
	fmt.Printf("Secp256r1 Address: %s\n", r1Kp.SuiAddress())
}
```

### Signing and Verifying

The `keypair` interface provides methods for signing transactions and personal messages.

```go
// Sign a transaction (intent-scoped)
sig, err := kp.SignTransaction(txBytes)

// Sign a personal message (intent-scoped)
sig, err := kp.SignPersonalMessage([]byte("Hello, Sui!"))

// Verify a personal message
err := kp.VerifyPersonalMessage([]byte("Hello, Sui!"), sig)
```

## Intent Signing

Sui uses intent scopes to distinguish between different types of signed messages (e.g., transactions, personal messages). The `intent` package handles the serialization of intent messages.

## Sub-packages

*   `ed25519`: Ed25519 keypair implementation.
*   `secp256k1`: Secp256k1 keypair implementation.
*   `secp256r1`: Secp256r1 keypair implementation.
*   `intent`: Intent scoping for signatures.
*   `personalmsg`: Utilities for encoding personal messages.
*   `transaction`: Utilities for signing transactions.
