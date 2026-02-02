package query

import (
	"context"
	"fmt"
	"log"

	"github.com/open-move/sui-go-sdk/graphql"
	"github.com/open-move/sui-go-sdk/types"
	"github.com/open-move/sui-go-sdk/utils"
)

// Objects demonstrates how to fetch objects.
func Objects(ctx context.Context, client *graphql.Client) {
	fmt.Println("=== GetObject ===")
	objectID, err := utils.ParseAddress("0xf41564ce5236f344bc79abb0c6ca22bb31edc4ec64b995824e986b81e71eb031")
	if err != nil {
		log.Printf("invalid object id: %v", err)
		return
	}
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
	objectIDs := make([]types.Address, 0, 2)
	for _, id := range []string{
		"0xf41564ce5236f344bc79abb0c6ca22bb31edc4ec64b995824e986b81e71eb031",
		"0xf31065dcbc46e24bba4c7655eb5ce804067f33a73b30643caa35dc1c20adc2ef",
	} {
		addr, err := utils.ParseAddress(id)
		if err != nil {
			log.Printf("invalid object id: %v", err)
			return
		}
		objectIDs = append(objectIDs, addr)
	}

	objects, err := client.GetMultipleObjects(ctx, objectIDs, nil)
	if err != nil {
		log.Printf("GetMultipleObjects error: %v", err)
	} else {
		printJSON("GetMultipleObjects result", objects)
	}
	fmt.Println()

	fmt.Println("=== GetOwnedObjects ===")
	owner, err := utils.ParseAddress("0x559ef1509af6e837d4153b3b08d9534d3df3f336a5cb6498fa248ce6cb2172e6")
	if err != nil {
		log.Printf("invalid owner address: %v", err)
		return
	}

	owned, err := client.GetOwnedObjects(ctx, owner, nil, &graphql.PaginationArgs{
		First: utils.Ptr(5),
	})
	if err != nil {
		log.Printf("GetOwnedObjects error: %v", err)
	} else {
		printJSON("GetOwnedObjects result", owned)
	}
}
