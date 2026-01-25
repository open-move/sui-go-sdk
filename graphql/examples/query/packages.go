package query

import (
	"context"
	"fmt"
	"log"

	"github.com/open-move/sui-go-sdk/graphql"
)

// Packages demonstrates how to fetch packages and modules.
func Packages(ctx context.Context, client *graphql.Client) {
	// Example: Get framework package (0x2)
	fmt.Println("=== GetPackage ===")
	pkg, err := client.GetPackage(ctx, "0x2")
	if err != nil {
		log.Printf("GetPackage error: %v", err)
		return // Changed from log.Fatal to return to avoid stopping other examples
	}

	printJSON("GetPackage result", pkg)
	fmt.Printf("Package Address: %s\n", pkg.Address)
	fmt.Printf("Version: %d\n", pkg.Version)
	fmt.Printf("Digest: %s\n", pkg.Digest)
	fmt.Println()

	// Example: Get a specific module
	fmt.Println("=== GetModule (coin) ===")
	module, err := client.GetModule(ctx, "0x2", "coin")
	if err != nil {
		log.Printf("GetModule error: %v", err)
	} else if module != nil {
		printJSON("GetModule result", module)
		fmt.Printf("Module Name: %s\n", module.Name)
		fmt.Printf("File Format Version: %d\n", module.FileFormatVersion)
		if module.Functions != nil {
			fmt.Printf("Functions: %d\n", len(module.Functions.Nodes))
		}
		if module.Structs != nil {
			fmt.Printf("Structs: %d\n", len(module.Structs.Nodes))
		}
	}
	fmt.Println()

	// Example: Get normalized Move function
	fmt.Println("=== GetNormalizedMoveFunction (coin::value) ===")
	fn, err := client.GetNormalizedMoveFunction(ctx, "0x2", "coin", "value")
	if err != nil {
		log.Printf("GetNormalizedMoveFunction error: %v", err)
	} else if fn != nil {
		printJSON("GetNormalizedMoveFunction result", fn)
		fmt.Printf("Function: %s\n", fn.Name)
		fmt.Printf("Visibility: %s\n", fn.Visibility)
		fmt.Printf("Is Entry: %v\n", fn.IsEntry)
		fmt.Printf("Type Parameters: %d\n", len(fn.TypeParameters))
		fmt.Printf("Parameters: %d\n", len(fn.Parameters))
		fmt.Printf("Return Types: %d\n", len(fn.Return))
	}
	fmt.Println()

	// Example: Get normalized Move struct
	fmt.Println("=== GetNormalizedMoveStruct (coin::Coin) ===")
	s, err := client.GetNormalizedMoveStruct(ctx, "0x2", "coin", "Coin")
	if err != nil {
		log.Printf("GetNormalizedMoveStruct error: %v", err)
	} else if s != nil {
		printJSON("GetNormalizedMoveStruct result", s)
		fmt.Printf("Struct: %s\n", s.Name)
		fmt.Printf("Abilities: %v\n", s.Abilities)
		fmt.Printf("Type Parameters: %d\n", len(s.TypeParameters))
		fmt.Printf("Fields:\n")
		for _, f := range s.Fields {
			if f.Type != nil {
				fmt.Printf("  - %s: %s\n", f.Name, f.Type.Repr)
			}
		}
	}
}
