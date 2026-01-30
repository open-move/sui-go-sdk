// Package mutation contains examples of GraphQL mutations.
package mutation

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/open-move/sui-go-sdk/graphql"
	"github.com/open-move/sui-go-sdk/types"
	"github.com/open-move/sui-go-sdk/utils"
)

// GetGasPrice gets the reference gas price from the network
func GetGasPrice(ctx context.Context, client *graphql.Client) (uint64, error) {
	query := `
		query {
			epoch {
				referenceGasPrice
			}
		}
	`

	var result struct {
		Epoch *struct {
			ReferenceGasPrice graphql.BigInt `json:"referenceGasPrice"`
		} `json:"epoch"`
	}

	err := client.Execute(ctx, query, nil, &result)
	if err != nil {
		return 0, err
	}

	if result.Epoch == nil {
		return 0, fmt.Errorf("no epoch data")
	}

	// Convert BigInt (string) to uint64
	gasPrice, err := strconv.ParseUint(string(result.Epoch.ReferenceGasPrice), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse gas price: %w", err)
	}

	return gasPrice, nil
}

// extractUpdatedGasCoin finds the gas coin in the transaction effects and returns
// its updated ObjectRef (with new version and digest after the transaction).
func extractUpdatedGasCoin(result *graphql.ExecuteTransactionResult, gasCoinAddress string) *types.ObjectRef {
	if result == nil || result.Effects == nil || result.Effects.ObjectChanges == nil {
		return nil
	}

	for _, change := range result.Effects.ObjectChanges.Nodes {
		// Check if this is our gas coin by comparing addresses
		if string(change.Address) == gasCoinAddress && change.OutputState != nil {
			// Parse the updated object ref
			ref, err := utils.ParseObjectRef(
				string(change.OutputState.Address),
				uint64(change.OutputState.Version),
				change.OutputState.Digest,
			)
			if err != nil {
				log.Printf("Failed to parse updated gas coin ref: %v", err)
				return nil
			}
			return &ref
		}
	}
	return nil
}

// PrintJSON prints a value as pretty-printed JSON (for debugging)
func PrintJSON(label string, v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Printf("JSON marshal error for %s: %v", label, err)
		return
	}
	fmt.Printf("%s:\n%s\n\n", label, string(data))
}
