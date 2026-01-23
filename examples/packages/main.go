// Package main demonstrates package and module queries using the Sui GraphQL SDK.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/joho/godotenv"
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

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	client := NewClient(
		WithEndpoint(TestnetEndpoint),
		WithTimeout(30*time.Second),
	)

	ctx := context.Background()

	// Example: Get framework package (0x2)
	fmt.Println("=== GetPackage ===")
	pkg, err := client.GetPackage(ctx, "0x2")
	if err != nil {
		log.Fatalf("GetPackage error: %v", err)
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
	fmt.Println()
}
