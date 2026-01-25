package query

import (
	"context"
	"fmt"
	"log"

	"github.com/open-move/sui-go-sdk/graphql"
)

// Transactions demonstrates how to fetch transactions.
func Transactions(ctx context.Context, client *graphql.Client, address string) {
	fmt.Println("=== GetTransactionBlock ===")
	digest := "3KmTo5yvbkeg9mrQafUkfFoYVRt45zTg2SoGWBJw615V"
	options := &graphql.TransactionBlockOptions{
		ShowInput:          true,
		ShowRawInput:       true,
		ShowEffects:        true,
		ShowEvents:         true,
		ShowObjectChanges:  true,
		ShowBalanceChanges: true,
	}

	tx, err := client.GetTransactionBlock(ctx, digest, options)
	if err != nil {
		log.Printf("GetTransactionBlock error: %v", err)
	} else if tx != nil {
		printJSON("GetTransactionBlock result", tx)
	}
	fmt.Println()

	fmt.Println("=== GetMultipleTransactionBlocks ===")
	digests := []string{
		"3KmTo5yvbkeg9mrQafUkfFoYVRt45zTg2SoGWBJw615V",
	}

	txs, err := client.GetMultipleTransactionBlocks(ctx, digests, nil)
	if err != nil {
		log.Printf("GetMultipleTransactionBlocks error: %v", err)
	} else {
		printJSON("GetMultipleTransactionBlocks result", txs)
	}
	fmt.Println()

	fmt.Println("=== QueryTransactionBlocks ===")
	filter := &graphql.TransactionFilter{
		SentAddress: graphql.Ptr(graphql.SuiAddress(address)),
	}

	txsConnection, err := client.QueryTransactionBlocks(ctx, filter, &graphql.PaginationArgs{
		First: graphql.Ptr(10),
	})
	if err != nil {
		log.Printf("QueryTransactionBlocks error: %v", err)
	} else if txsConnection != nil {
		printJSON("QueryTransactionBlocks result", txsConnection)
	}
}
