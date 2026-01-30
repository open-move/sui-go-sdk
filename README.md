# Sui Go SDK

![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)

The official Go SDK for the Sui blockchain, providing gRPC and GraphQL clients, along with cryptography utilities.

## Modules

This SDK includes the following main modules:

*   **gRPC Client**: A strongly-typed gRPC client for interacting with Sui RPC services.
*   **GraphQL Client**: A client for interacting with the Sui GraphQL API.
*   **Cryptography**: Utilities for key generation, signing, and verification (Ed25519, Secp256k1, Secp256r1).
*   **Keychain**: Key derivation (BIP-32), mnemonic handling (BIP-39), and address generation.
*   **Keypair**: Interfaces and helpers for managing different types of keypairs.
*   **Transaction**: A powerful builder for constructing Programmable Transactions.
*   **Types**: Common Sui types (Addresses, ObjectRefs, etc.) and BCS serialization.
*   **Typetag**: Utilities for parsing and manipulating Move type tags.

## Project Structure

```
sui-go-sdk/
├── cryptography/ # Cryptographic primitives (Ed25519, Secp256k1, Secp256r1)
├── graphql/      # GraphQL client and query/mutation builders
├── grpc/         # gRPC client for Sui RPC services
├── keychain/     # Key management, BIP-32/BIP-39, and address derivation
├── keypair/      # Keypair interfaces and high-level derivation logic
├── proto/        # Generated Protocol Buffer files
├── transaction/  # Transaction building and serialization
├── types/        # Common Sui types
└── typetag/      # Move type tag parsing and handling
```

## Installation

```bash
go get github.com/open-move/sui-go-sdk
```

## Usage

### gRPC Client

The gRPC client provides a direct interface to Sui's RPC services.

```go
package main

import (
	"context"
	"log"

	"github.com/open-move/sui-go-sdk/grpc"
)

func main() {
	ctx := context.Background()

	// Create a client connected to the Sui Mainnet
	client, err := grpc.NewMainnetClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Use the client to interact with the blockchain...
}
```

For more details, see the [gRPC README](grpc/README.md).

### GraphQL Client

The GraphQL client allows you to query the Sui blockchain using the flexible GraphQL API.

```go
package main

import (
	"github.com/open-move/sui-go-sdk/graphql"
)

func main() {
	// Create a client connected to the Sui Mainnet
	client := graphql.NewClient(graphql.WithEndpoint(graphql.MainnetEndpoint))

	// Use the client to execute queries and mutations...
}
```

For more details, see the [GraphQL README](graphql/README.md).

### Cryptography

The cryptography module supports key pair generation and signing for transactions and personal messages.

```go
package main

import (
	"fmt"
	"log"

	"github.com/open-move/sui-go-sdk/cryptography/ed25519"
)

func main() {
	// Generate a new Ed25519 keypair
	kp, err := ed25519.Generate()
	if err != nil {
		log.Fatalf("Failed to generate keypair: %v", err)
	}

	fmt.Printf("Sui Address: %s\n", kp.SuiAddress())
}
```

For more details, see the [Cryptography README](cryptography/README.md).

### Transaction Builder

The `transaction` package allows you to build Programmable Transactions easily.

```go
package main

import (
	"context"
	
	"github.com/open-move/sui-go-sdk/transaction"
	"github.com/open-move/sui-go-sdk/types"
)

func main() {
	// Initialize builder
	b := transaction.New()
	
	// Add inputs and commands
	// ...
	
	// Build the transaction
	// result, err := b.Build(context.Background(), transaction.BuildOptions{})
}
```

### Keychain

The `keychain` package handles mnemonics and key derivation.

```go
package main

import (
	"fmt"
	"github.com/open-move/sui-go-sdk/keychain"
)

func main() {
	// Generate a new mnemonic
	mnemonic, _ := keychain.NewMnemonic()
	fmt.Println("Mnemonic:", mnemonic)
}
```

## Contributing

We welcome contributions! Please see [CONTRIBUTION.md](CONTRIBUTION.md) for guidelines on how to contribute to this project.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
