package main

import (
	"context"
	"fmt"
	"log"

	"github.com/open-move/sui-go-sdk/grpc"
	"github.com/open-move/sui-go-sdk/keypair"
	"github.com/open-move/sui-go-sdk/transaction"
)

const bech32Key = "suiprivkey1qz6qzxye624vk8epr7c9j4flnxm5lze2e7y2pmxzm4qarny03lt8xavx8zj"

func main() {
	ctx := context.Background()
	client, err := grpc.NewClient(ctx, grpc.TestnetFullnodeURL)
	if err != nil {
		log.Fatalf("dial grpc: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	kp, err := keypair.FromBech32(bech32Key)
	if err != nil {
		log.Fatalf("from bech32: %v", err)
	}

	tx := transaction.New()
	tx.MoveCall(transaction.MoveCall{
		Target:    "0x2::clock::timestamp_ms",
		Arguments: []transaction.Argument{tx.Object("0x6")},
	})

	executed, err := client.SignAndExecuteTransaction(ctx, tx, kp, nil)
	if err != nil {
		log.Fatalf("execute: %v", err)
	}
	fmt.Printf("digest: %s\n", executed.GetDigest())
}
