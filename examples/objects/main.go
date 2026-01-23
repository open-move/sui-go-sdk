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

func main() {
	client := NewClient(
		WithEndpoint(TestnetEndpoint),
		WithTimeout(30*time.Second),
	)

	ctx := context.Background()

	fmt.Println("=== GetObject ===")
	objectID := SuiAddress("0xf41564ce5236f344bc79abb0c6ca22bb31edc4ec64b995824e986b81e71eb031")
	options := &ObjectDataOptions{
		ShowType:                true,
		ShowContent:             true,
		ShowOwner:               true,
		ShowPreviousTransaction: true,
		ShowStorageRebate:       true,
		ShowDisplay:             true,
	}

	obj, err := client.GetObject(ctx, objectID, options)
	if err != nil {
		log.Printf("GetObject error: %v", err)
	} else if obj != nil {
		printJSON("GetObject result", obj)
	}
	fmt.Println()

	fmt.Println("=== GetMultipleObjects ===")
	objectIDs := []SuiAddress{
		"0xf41564ce5236f344bc79abb0c6ca22bb31edc4ec64b995824e986b81e71eb031",
		"0xf31065dcbc46e24bba4c7655eb5ce804067f33a73b30643caa35dc1c20adc2ef",
	}

	objects, err := client.GetMultipleObjects(ctx, objectIDs, nil)
	if err != nil {
		log.Printf("GetMultipleObjects error: %v", err)
	} else {
		printJSON("GetMultipleObjects result", objects)
	}
	fmt.Println()

	fmt.Println("=== GetOwnedObjects ===")
	owner := SuiAddress("0x559ef1509af6e837d4153b3b08d9534d3df3f336a5cb6498fa248ce6cb2172e6")

	owned, err := client.GetOwnedObjects(ctx, owner, nil, &PaginationArgs{
		First: Ptr(5),
	})
	if err != nil {
		log.Printf("GetOwnedObjects error: %v", err)
	} else {
		printJSON("GetOwnedObjects result", owned)
	}
}
