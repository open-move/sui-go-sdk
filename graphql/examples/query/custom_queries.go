package query

import (
	"context"
	"fmt"
	"log"

	"github.com/open-move/sui-go-sdk/graphql"
)

// cqTransactionResult defines the expected response structure for our transaction query.
type cqTransactionResult struct {
	Transaction *struct {
		Digest string `json:"digest"`
		Sender *struct {
			Address string `json:"address"`
		} `json:"sender"`
		Effects *struct {
			Status     string `json:"status"`
			Timestamp  string `json:"timestamp"`
			GasEffects *struct {
				GasSummary *struct {
					ComputationCost interface{} `json:"computationCost"`
					StorageCost     interface{} `json:"storageCost"`
					StorageRebate   interface{} `json:"storageRebate"`
				} `json:"gasSummary"`
			} `json:"gasEffects"`
		} `json:"effects"`
	} `json:"transaction"`
}

func exampleBasicCustomQuery(client *graphql.Client, ctx context.Context, digest string) {
	fmt.Println("=== Example 1: Basic Custom Query - Get Transaction ===")

	// Define a custom query using graphql.Query()
	// This is similar to the TypeScript gql`` tagged template
	getTransaction := graphql.Query(`
		query getTransaction($digest: String!) {
			transaction(digest: $digest) {
				digest
				sender { address }
				effects {
					status
					timestamp
					gasEffects {
						gasSummary {
							computationCost
							storageCost
							storageRebate
						}
					}
				}
			}
		}
	`)

	// Execute the query with variables
	var result cqTransactionResult
	err := client.Query(ctx, getTransaction, graphql.Vars{
		"digest": digest,
	}, &result)

	if err != nil {
		log.Printf("Query error: %v", err)
		return
	}

	printJSON("Transaction Result", result)

	if result.Transaction != nil {
		fmt.Printf("\nTransaction Digest: %s\n", result.Transaction.Digest)
		if result.Transaction.Sender != nil {
			fmt.Printf("Sender: %s\n", result.Transaction.Sender.Address)
		}
		if result.Transaction.Effects != nil {
			fmt.Printf("Status: %s\n", result.Transaction.Effects.Status)
		}
	}
	fmt.Println()
}

// cqTransactionWithSenderAssets combines transaction data with the sender's balances.
type cqTransactionWithSenderAssets struct {
	Transaction *struct {
		Digest string `json:"digest"`
		Sender *struct {
			Address  string `json:"address"`
			Balances *struct {
				Nodes []struct {
					CoinType *struct {
						Repr string `json:"repr"`
					} `json:"coinType"`
					TotalBalance string `json:"totalBalance"`
				} `json:"nodes"`
			} `json:"balances"`
		} `json:"sender"`
		Effects *struct {
			Status         string `json:"status"`
			Timestamp      string `json:"timestamp"`
			BalanceChanges *struct {
				Nodes []struct {
					Owner *struct {
						Address string `json:"address"`
					} `json:"owner"`
					CoinType *struct {
						Repr string `json:"repr"`
					} `json:"coinType"`
					Amount string `json:"amount"`
				} `json:"nodes"`
			} `json:"balanceChanges"`
		} `json:"effects"`
	} `json:"transaction"`
}

func exampleTransactionWithSenderAssets(client *graphql.Client, ctx context.Context, digest string) {
	fmt.Println("=== Example 2: Transaction with Sender's Assets ===")

	// This custom query fetches:
	// 1. Transaction details (digest, status, timestamp)
	// 2. Sender's current balances (all coin types)
	// 3. Balance changes from the transaction
	getTransactionWithAssets := graphql.Query(`
		query getTransactionWithSenderAssets($digest: String!) {
			transaction(digest: $digest) {
				digest
				sender {
					address
					balances {
						nodes {
							coinType { repr }
							totalBalance
						}
					}
				}
				effects {
					status
					timestamp
					balanceChanges {
						nodes {
							owner {
								address
							}
							coinType { repr }
							amount
						}
					}
				}
			}
		}
	`)

	var result cqTransactionWithSenderAssets
	err := client.Query(ctx, getTransactionWithAssets, graphql.Vars{
		"digest": digest,
	}, &result)

	if err != nil {
		log.Printf("Query error: %v", err)
		return
	}

	printJSON("Transaction with Sender Assets", result)

	if result.Transaction != nil && result.Transaction.Sender != nil {
		fmt.Printf("\n--- Sender: %s ---\n", result.Transaction.Sender.Address)

		if result.Transaction.Sender.Balances != nil {
			fmt.Println("\nCurrent Balances:")
			for _, balance := range result.Transaction.Sender.Balances.Nodes {
				coinType := "unknown"
				if balance.CoinType != nil {
					coinType = balance.CoinType.Repr
				}
				fmt.Printf("  %s: %s\n", coinType, balance.TotalBalance)
			}
		}

		if result.Transaction.Effects != nil && result.Transaction.Effects.BalanceChanges != nil {
			fmt.Println("\nBalance Changes from Transaction:")
			for _, change := range result.Transaction.Effects.BalanceChanges.Nodes {
				coinType := "unknown"
				if change.CoinType != nil {
					coinType = change.CoinType.Repr
				}
				owner := "unknown"
				if change.Owner != nil {
					owner = change.Owner.Address
				}
				fmt.Printf("  %s: %s (%s)\n", owner, change.Amount, coinType)
			}
		}
	}
	fmt.Println()
}

// cqSenderAssetsResult is the typed result structure.
type cqSenderAssetsResult struct {
	Address *struct {
		Address  string `json:"address"`
		Balances *struct {
			Nodes []struct {
				CoinType *struct {
					Repr string `json:"repr"`
				} `json:"coinType"`
				TotalBalance string `json:"totalBalance"`
			} `json:"nodes"`
		} `json:"balances"`
		Objects *struct {
			Nodes []struct {
				Address  string `json:"address"`
				Version  uint64 `json:"version"`
				Digest   string `json:"digest"`
				Contents *struct {
					Type struct {
						Repr string `json:"repr"`
					} `json:"type"`
				} `json:"contents"`
			} `json:"nodes"`
		} `json:"objects"`
	} `json:"address"`
}

func exampleTypedQuery(client *graphql.Client, ctx context.Context, address string) {
	fmt.Println("=== Example 3: TypedQuery - Get Address Assets ===")

	// Create a typed query with compile-time type safety
	getAddressAssets := graphql.NewTypedQuery[cqSenderAssetsResult](`
		query getAddressAssets($address: SuiAddress!) {
			address(address: $address) {
				address
				balances {
					nodes {
						coinType { repr }
						totalBalance
					}
				}
				objects(first: 5) {
					nodes {
						address
						version
						digest
						contents {
							type { repr }
						}
					}
				}
			}
		}
	`)

	// Execute returns a typed result
	result, err := getAddressAssets.Execute(ctx, client, graphql.Vars{
		"address": address,
	})

	if err != nil {
		log.Printf("Query error: %v", err)
		return
	}

	printJSON("Address Assets", result)

	if result.Address != nil {
		fmt.Printf("\n--- Address: %s ---\n", result.Address.Address)

		if result.Address.Balances != nil {
			fmt.Println("\nBalances:")
			for _, balance := range result.Address.Balances.Nodes {
				coinType := "unknown"
				if balance.CoinType != nil {
					coinType = balance.CoinType.Repr
				}
				fmt.Printf("  %s: %s\n", coinType, balance.TotalBalance)
			}
		}

		if result.Address.Objects != nil {
			fmt.Println("\nOwned Objects (first 5):")
			for _, obj := range result.Address.Objects.Nodes {
				objType := "unknown"
				if obj.Contents != nil {
					objType = obj.Contents.Type.Repr
				}
				fmt.Printf("  %s (v%d): %s\n", obj.Address, obj.Version, objType)
			}
		}
	}
	fmt.Println()
}

func exampleQueryRequestAPI(client *graphql.Client, ctx context.Context, digest string) {
	fmt.Println("=== Example 4: QueryRequest Fluent API ===")

	// Define the query
	getTxQuery := graphql.Query(`
		query getTx($digest: String!) {
			transaction(digest: $digest) {
				digest
				sender { address }
				effects { status }
			}
		}
	`)

	// Use the fluent API to build and execute the query
	var result cqTransactionResult
	err := client.NewQueryRequest(getTxQuery).
		Var("digest", digest).
		Execute(ctx, &result)

	if err != nil {
		log.Printf("Query error: %v", err)
		return
	}

	printJSON("QueryRequest Result", result)
	fmt.Println()
}

func examplePredefinedQueries(client *graphql.Client, ctx context.Context, digest string, address string) {
	fmt.Println("=== Example 5: Using Pre-defined Queries ===")

	// Use the pre-defined GetTransactionQuery
	var txResult struct {
		Transaction *struct {
			Digest string `json:"digest"`
			Sender *struct {
				Address string `json:"address"`
			} `json:"sender"`
			Effects *struct {
				Status    string `json:"status"`
				Timestamp string `json:"timestamp"`
			} `json:"effects"`
		} `json:"transaction"`
	}

	err := client.Query(ctx, graphql.GetTransactionQuery, graphql.Vars{
		"digest": digest,
	}, &txResult)

	if err != nil {
		log.Printf("GetTransactionQuery error: %v", err)
	} else {
		printJSON("Pre-defined GetTransactionQuery Result", txResult)
	}

	// Use the pre-defined GetBalanceQuery
	var balanceResult struct {
		Address *struct {
			Balance *struct {
				CoinType *struct {
					Repr string `json:"repr"`
				} `json:"coinType"`
				TotalBalance string `json:"totalBalance"`
			} `json:"balance"`
		} `json:"address"`
	}

	err = client.Query(ctx, graphql.GetBalanceQuery, graphql.Vars{
		"address":  address,
		"coinType": "0x2::sui::SUI",
	}, &balanceResult)

	if err != nil {
		log.Printf("GetBalanceQuery error: %v", err)
	} else {
		printJSON("Pre-defined GetBalanceQuery Result", balanceResult)
	}

	fmt.Println()
}

// cqMultiTransactionResult represents multiple transactions with their senders.
type cqMultiTransactionResult struct {
	Transactions *struct {
		Nodes []struct {
			Digest string `json:"digest"`
			Sender *struct {
				Address  string `json:"address"`
				Balances *struct {
					Nodes []struct {
						CoinType *struct {
							Repr string `json:"repr"`
						} `json:"coinType"`
						TotalBalance string `json:"totalBalance"`
					} `json:"nodes"`
				} `json:"balances"`
			} `json:"sender"`
			Effects *struct {
				Status    string `json:"status"`
				Timestamp string `json:"timestamp"`
			} `json:"effects"`
		} `json:"nodes"`
		PageInfo struct {
			HasNextPage bool    `json:"hasNextPage"`
			EndCursor   *string `json:"endCursor"`
		} `json:"pageInfo"`
	} `json:"transactions"`
}

func exampleMultipleTransactions(client *graphql.Client, ctx context.Context, senderAddress string) {
	fmt.Println("=== Example 6: Query Multiple Transactions with Sender Assets ===")

	// Query recent transactions from a specific sender with their current balances
	getRecentTransactions := graphql.Query(`
		query getRecentTransactions($sender: SuiAddress!, $first: Int) {
			transactions(filter: { sentAddress: $sender }, first: $first) {
				nodes {
					digest
					sender {
						address
						balances(first: 3) {
							nodes {
								coinType { repr }
								totalBalance
							}
						}
					}
					effects {
						status
						timestamp
					}
				}
				pageInfo {
					hasNextPage
					endCursor
				}
			}
		}
	`)

	var result cqMultiTransactionResult
	err := client.Query(ctx, getRecentTransactions, graphql.Vars{
		"sender": senderAddress,
		"first":  5,
	}, &result)

	if err != nil {
		log.Printf("Query error: %v", err)
		return
	}

	printJSON("Recent Transactions", result)

	if result.Transactions != nil {
		fmt.Printf("\nFound %d transactions\n", len(result.Transactions.Nodes))
		for i, tx := range result.Transactions.Nodes {
			fmt.Printf("\n--- Transaction %d ---\n", i+1)
			fmt.Printf("Digest: %s\n", tx.Digest)
			if tx.Sender != nil {
				fmt.Printf("Sender: %s\n", tx.Sender.Address)
			}
			if tx.Effects != nil {
				fmt.Printf("Status: %s\n", tx.Effects.Status)
				fmt.Printf("Timestamp: %s\n", tx.Effects.Timestamp)
			}
		}
	}
	fmt.Println()
}

func CustomQueries(ctx context.Context, client *graphql.Client, digest string, address string) {
	// Run all examples
	exampleBasicCustomQuery(client, ctx, digest)
	exampleTransactionWithSenderAssets(client, ctx, digest)
	exampleTypedQuery(client, ctx, address)
	exampleQueryRequestAPI(client, ctx, digest)
	examplePredefinedQueries(client, ctx, digest, address)
	exampleMultipleTransactions(client, ctx, address)

	fmt.Println("=== All Custom Query Examples Complete ===")
}
