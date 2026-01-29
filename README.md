# Sui Go SDK

![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)

The official Go SDK for the Sui blockchain, providing gRPC and GraphQL clients, along with cryptography utilities.

## Modules

This SDK includes the following main modules:

*   **[gRPC Client](grpc/README.md)**: A strongly-typed gRPC client for interacting with Sui RPC services.
*   **[GraphQL Client](graphql/README.md)**: A client for interacting with the Sui GraphQL API.
*   **[Cryptography](cryptography/README.md)**: Utilities for key generation, signing, and verification (Ed25519, Secp256k1, Secp256r1).

## Installation

```bash
go get github.com/open-move/sui-go-sdk
```

## Quick Start

### 1. Initialize a Client

**gRPC Client** (Recommended for RPC interactions)
```go
package main

import (
	"context"
	"log"
	"github.com/open-move/sui-go-sdk/grpc"
)

func main() {
	ctx := context.Background()
	// Connect to Mainnet
	client, err := grpc.NewMainnetClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()
}
```

**GraphQL Client** (Recommended for complex data queries)
```go
package main

import (
	"github.com/open-move/sui-go-sdk/graphql"
)

func main() {
	// Connect to Mainnet with custom options
	client := graphql.NewClient(
		graphql.WithEndpoint(graphql.MainnetEndpoint),
	)
}
```

## Core Concepts: Lessons from the SDK Design

As a developer using this SDK, you'll notice several key design patterns that reflect the underlying architecture of Sui.

### 1. The "Black Box" of Cryptography: Intent vs. Authentication

The SDK enforces a clear separation between **building** a transaction (defining intent) and **signing** it (authenticating intent). You don't need to know the math behind Ed25519, just that it acts as the "key" to authorize your actions.

```go
// 1. Build the transaction (The Intent)
// logic...

// 2. Sign the transaction (The Authentication)
// The SDK treats the keypair as an abstract signer.
signature, err := kp.SignTransaction(txBytes)
if err != nil {
    return nil
}
// The signature is just a byte array proving you authorized the 'txBytes'
```

### 2. GraphQL as the Transport Layer

The GraphQL client acts as a bridge. It handles the translation of raw bytes (like signatures) into a format the network understands (Base64), and manages the execution pipeline.

```go
import (
	"encoding/base64"
	"github.com/open-move/sui-go-sdk/graphql"
)

// The SDK handles type conversion automatically for you
func execute(client *graphql.Client, txBytes []byte, sig []byte) {
	// Convert raw bytes to the SDK's Base64 type for transport
	txBase64 := graphql.Base64(base64.StdEncoding.EncodeToString(txBytes))
	sigBase64 := graphql.Base64(base64.StdEncoding.EncodeToString(sig))

	// Execute: The bridge between local state and remote blockchain
	result, err := graphql.ExecuteTransaction(client, ctx, txBase64, []graphql.Base64{sigBase64})
}
```

### 3. Sui's Object-Centric Data Model

Sui's state is a graph of objects. The SDK's query structure reflects this, allowing you to traverse from an address to its owned objects and their specific fields.

```go
// A query that reflects the graph structure: Address -> Objects -> Contents
query := `
    query getObjects($address: SuiAddress!) {
        address(address: $address) {
            objects(first: 5) {
                nodes {
                    version
                    digest
                    contents {
                        type { repr }
                    }
                }
            }
        }
    }
`
```

### 4. Advanced Go Patterns

To make the SDK ergonomic, we use fluent builders and functional options. This allows you to construct complex requests in a readable, type-safe way.

```go
// Fluent Query Builder
qb := graphql.NewQueryBuilder()
qb.Field("epoch").
   Fields("referenceGasPrice").
   Done()

// Typed Queries with Generics
type AssetResult struct { ... }
query := graphql.NewTypedQuery[AssetResult](queryString)
result, err := query.Execute(ctx, client, vars)
```

## Project Structure

```
sui-go-sdk/
├── cryptography/ # Cryptographic primitives (Ed25519, Secp256k1, Secp256r1)
├── graphql/      # GraphQL client and query/mutation builders
├── grpc/         # gRPC client for Sui RPC services
├── keychain/     # Key management and derivation utilities
├── keypair/      # Keypair interfaces and derivation logic
├── proto/        # Generated Protocol Buffer files
├── transaction/  # Transaction building and serialization
├── types/        # Common Sui types
└── typetag/      # Move type tag parsing and handling
```
