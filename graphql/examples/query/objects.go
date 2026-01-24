package query

import (
	"context"
	"fmt"
	"log"

	"github.com/open-move/sui-go-sdk/graphql"
)

func Objects(ctx context.Context, client *graphql.Client) {
	fmt.Println("=== GetObject ===")
	objectID := graphql.SuiAddress("0xf41564ce5236f344bc79abb0c6ca22bb31edc4ec64b995824e986b81e71eb031")
	options := &graphql.ObjectDataOptions{
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
	objectIDs := []graphql.SuiAddress{
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
	owner := graphql.SuiAddress("0x559ef1509af6e837d4153b3b08d9534d3df3f336a5cb6498fa248ce6cb2172e6")

	owned, err := client.GetOwnedObjects(ctx, owner, nil, &graphql.PaginationArgs{
		First: graphql.Ptr(5),
	})
	if err != nil {
		log.Printf("GetOwnedObjects error: %v", err)
	} else {
		printJSON("GetOwnedObjects result", owned)
	}
}
