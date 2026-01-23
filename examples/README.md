# Sui GraphQL SDK Examples

This directory contains runnable examples demonstrating how to use the Sui GraphQL SDK.

## Prerequisites

- Go 1.21 or later
- Network access to Sui testnet

## Running Examples

Each subdirectory contains a standalone example that can be run with:

```bash
go run ./examples/<category>/main.go
```

## Available Examples

| Directory | Description |
|-----------|-------------|
| `balances/` | Balance and coin queries |
| `checkpoints/` | Checkpoint queries |
| `epochs/` | Epoch queries |
| `events/` | Event queries |
| `packages/` | Package and module queries |
| `protocol/` | Protocol and system configuration |
| `query_builder/` | Custom query building |
| `faucet/` | Faucet client usage |

## Notes

- Examples use the **testnet** endpoint by default
- Some examples require valid address/object IDs to work
- Rate limiting may apply to faucet requests
