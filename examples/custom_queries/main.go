// Package main demonstrates custom GraphQL queries using the Sui GraphQL SDK.
// This example shows how to write and execute custom queries for fetching
// transaction details and the sender's assets in a single query.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/open-move/sui-go-sdk/graphql"
)

func printJSON(label string, v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Printf("JSON marshal error for %s: %v", label, err)
		return
	}
	fmt.Printf("%s:\n%s\n", label, string(data))
}

// =============================================================================
// Example 1: Basic Custom Query - Get Transaction Details
// =============================================================================

// TransactionResult defines the expected response structure for our transaction query.
type TransactionResult struct {
	TransactionBlock *struct {
		Digest string `json:"digest"`
		Sender *struct {
			Address string `json:"address"`
		} `json:"sender"`
		Effects *struct {
			Status     string `json:"status"`
			Timestamp  string `json:"timestamp"`
			GasEffects *struct {
				GasSummary *struct {
					ComputationCost string `json:"computationCost"`
					StorageCost     string `json:"storageCost"`
					StorageRebate   string `json:"storageRebate"`
				} `json:"gasSummary"`
			} `json:"gasEffects"`
		} `json:"effects"`
	} `json:"transactionBlock"`
}

func exampleBasicCustomQuery(client *Client, ctx context.Context, digest string) {
	fmt.Println("=== Example 1: Basic Custom Query - Get Transaction ===")

	// Define a custom query using Query()
	// This is similar to the TypeScript gql`` tagged template
	getTransaction := Query(`
		query getTransaction($digest: String!) {
			transactionBlock(digest: $digest) {
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
	var result TransactionResult
	err := client.Query(ctx, getTransaction, Vars{
		"digest": digest,
	}, &result)

	if err != nil {
		log.Printf("Query error: %v", err)
		return
	}

	printJSON("Transaction Result", result)

	if result.TransactionBlock != nil {
		fmt.Printf("\nTransaction Digest: %s\n", result.TransactionBlock.Digest)
		if result.TransactionBlock.Sender != nil {
			fmt.Printf("Sender: %s\n", result.TransactionBlock.Sender.Address)
		}
		if result.TransactionBlock.Effects != nil {
			fmt.Printf("Status: %s\n", result.TransactionBlock.Effects.Status)
		}
	}
	fmt.Println()
}

// =============================================================================
// Example 2: Query Transaction and Sender's Assets
// =============================================================================

// TransactionWithSenderAssets combines transaction data with the sender's balances.
type TransactionWithSenderAssets struct {
	TransactionBlock *struct {
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
						Typename string `json:"__typename"`
						Address  *struct {
							Address string `json:"address"`
						} `json:"address"`
					} `json:"owner"`
					CoinType *struct {
						Repr string `json:"repr"`
					} `json:"coinType"`
					Amount string `json:"amount"`
				} `json:"nodes"`
			} `json:"balanceChanges"`
		} `json:"effects"`
	} `json:"transactionBlock"`
}

func exampleTransactionWithSenderAssets(client *Client, ctx context.Context, digest string) {
	fmt.Println("=== Example 2: Transaction with Sender's Assets ===")

	// This custom query fetches:
	// 1. Transaction details (digest, status, timestamp)
	// 2. Sender's current balances (all coin types)
	// 3. Balance changes from the transaction
	getTransactionWithAssets := Query(`
		query getTransactionWithSenderAssets($digest: String!) {
			transactionBlock(digest: $digest) {
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
								__typename
								... on AddressOwner {
									address { address }
								}
							}
							coinType { repr }
							amount
						}
					}
				}
			}
		}
	`)

	var result TransactionWithSenderAssets
	err := client.Query(ctx, getTransactionWithAssets, Vars{
		"digest": digest,
	}, &result)

	if err != nil {
		log.Printf("Query error: %v", err)
		return
	}

	printJSON("Transaction with Sender Assets", result)

	if result.TransactionBlock != nil && result.TransactionBlock.Sender != nil {
		fmt.Printf("\n--- Sender: %s ---\n", result.TransactionBlock.Sender.Address)

		if result.TransactionBlock.Sender.Balances != nil {
			fmt.Println("\nCurrent Balances:")
			for _, balance := range result.TransactionBlock.Sender.Balances.Nodes {
				coinType := "unknown"
				if balance.CoinType != nil {
					coinType = balance.CoinType.Repr
				}
				fmt.Printf("  %s: %s\n", coinType, balance.TotalBalance)
			}
		}

		if result.TransactionBlock.Effects != nil && result.TransactionBlock.Effects.BalanceChanges != nil {
			fmt.Println("\nBalance Changes from Transaction:")
			for _, change := range result.TransactionBlock.Effects.BalanceChanges.Nodes {
				coinType := "unknown"
				if change.CoinType != nil {
					coinType = change.CoinType.Repr
				}
				owner := "unknown"
				if change.Owner != nil && change.Owner.Address != nil {
					owner = change.Owner.Address.Address
				}
				fmt.Printf("  %s: %s (%s)\n", owner, change.Amount, coinType)
			}
		}
	}
	fmt.Println()
}

// =============================================================================
// Example 3: Using TypedQuery for Type-Safe Queries
// =============================================================================

// SenderAssetsResult is the typed result structure.
type SenderAssetsResult struct {
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

func exampleTypedQuery(client *Client, ctx context.Context, address string) {
	fmt.Println("=== Example 3: TypedQuery - Get Address Assets ===")

	// Create a typed query with compile-time type safety
	getAddressAssets := NewTypedQuery[SenderAssetsResult](`
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
	result, err := getAddressAssets.Execute(ctx, client, Vars{
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

// =============================================================================
// Example 4: Using QueryRequest Fluent API
// =============================================================================

func exampleQueryRequestAPI(client *Client, ctx context.Context, digest string) {
	fmt.Println("=== Example 4: QueryRequest Fluent API ===")

	// Define the query
	getTxQuery := Query(`
		query getTx($digest: String!) {
			transactionBlock(digest: $digest) {
				digest
				sender { address }
				effects { status }
			}
		}
	`)

	// Use the fluent API to build and execute the query
	var result TransactionResult
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

// =============================================================================
// Example 5: Using Pre-defined Queries
// =============================================================================

func examplePredefinedQueries(client *Client, ctx context.Context, digest string, address string) {
	fmt.Println("=== Example 5: Using Pre-defined Queries ===")

	// Use the pre-defined GetTransactionQuery
	var txResult struct {
		TransactionBlock *struct {
			Digest string `json:"digest"`
			Sender *struct {
				Address string `json:"address"`
			} `json:"sender"`
			Effects *struct {
				Status    string `json:"status"`
				Timestamp string `json:"timestamp"`
			} `json:"effects"`
		} `json:"transactionBlock"`
	}

	err := client.Query(ctx, GetTransactionQuery, Vars{
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

	err = client.Query(ctx, GetBalanceQuery, Vars{
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

// =============================================================================
// Example 6: Complex Query - Multiple Transactions and Sender Info
// =============================================================================

// MultiTransactionResult represents multiple transactions with their senders.
type MultiTransactionResult struct {
	TransactionBlocks *struct {
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
	} `json:"transactionBlocks"`
}

func exampleMultipleTransactions(client *Client, ctx context.Context, senderAddress string) {
	fmt.Println("=== Example 6: Query Multiple Transactions with Sender Assets ===")

	// Query recent transactions from a specific sender with their current balances
	getRecentTransactions := Query(`
		query getRecentTransactions($sender: SuiAddress!, $first: Int) {
			transactionBlocks(filter: { sentAddress: $sender }, first: $first) {
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

	var result MultiTransactionResult
	err := client.Query(ctx, getRecentTransactions, Vars{
		"sender": senderAddress,
		"first":  5,
	}, &result)

	if err != nil {
		log.Printf("Query error: %v", err)
		return
	}

	printJSON("Recent Transactions", result)

	if result.TransactionBlocks != nil {
		fmt.Printf("\nFound %d transactions\n", len(result.TransactionBlocks.Nodes))
		for i, tx := range result.TransactionBlocks.Nodes {
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

func main() {
	// Create a client connected to testnet
	client := NewClient(
		WithEndpoint(TestnetEndpoint),
		WithTimeout(30*time.Second),
	)

	ctx := context.Background()

	// Test data - using a known testnet transaction and address
	testDigest := "3KmTo5yvbkeg9mrQafUkfFoYVRt45zTg2SoGWBJw615V"
	testAddress := "0x559ef1509af6e837d4153b3b08d9534d3df3f336a5cb6498fa248ce6cb2172e6"

	// Run all examples
	exampleBasicCustomQuery(client, ctx, testDigest)
	exampleTransactionWithSenderAssets(client, ctx, testDigest)
	exampleTypedQuery(client, ctx, testAddress)
	exampleQueryRequestAPI(client, ctx, testDigest)
	examplePredefinedQueries(client, ctx, testDigest, testAddress)
	exampleMultipleTransactions(client, ctx, testAddress)

	fmt.Println("=== All Custom Query Examples Complete ===")
}
