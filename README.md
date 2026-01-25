# Go SDK for Sui

# Go SDK for Sui

A Go SDK for building, signing, and submitting Sui transactions. Supports gRPC today, with GraphQL coming soon.

## Installation

```
go get github.com/open-move/sui-go-sdk
```

## Features

- gRPC client for Sui RPC services
- Transaction builder (`transaction.Transaction`)
- Keypairs and signing utilities
- Type tags and BCS helpers
- Coin selection and gas estimation

## Quick Start

```go
ctx := context.Background()
client, err := grpc.NewTestnetClient(ctx)
if err != nil {
    log.Fatal(err)
}
defer client.Close()

kp, err := keypair.FromBech32(bech32Key)
if err != nil {
    log.Fatal(err)
}

tx := transaction.New()
tx.MoveCall(transaction.MoveCall{
    Target:    "0x2::clock::timestamp_ms",
    Arguments: []transaction.Argument{tx.Object("0x6")},
})

executed, err := client.SignAndExecuteTransaction(ctx, tx, kp, nil)
if err != nil {
    log.Fatal(err)
}
fmt.Println("digest:", executed.GetDigest())
```

## Examples

- `examples/transactions`
- `examples/keypairs`

## Roadmap

- GraphQL client (coming soon)

## Testing

This repo includes some tests that run against Sui mainnet via gRPC. To run them:

```
go test ./tests
```

## Regenerating Protos

```
make proto
```

This expects `protoc` and `protoc-gen-go` / `protoc-gen-go-grpc` to be installed.

## Contributing

1. Fork and clone this repository.
2. Install Go 1.24+ and the protobuf toolchain (`protoc`, `protoc-gen-go`, `protoc-gen-go-grpc`).
3. Run `go test ./...` before submitting a PR. If you add helpers around live RPCs, include unit tests that use fakes alongside any integration tests.
4. Regenerate gRPC code with `make proto` when proto files change, and run `gofmt` on modified Go files.
5. Create a pull request describing the change.

Issues and feature requests are welcome via GitHub Issues.

## License

Apache-2.0.
