package grpc

import (
	"context"
	"testing"
)

func TestGetObject(t *testing.T) {
	ctx := context.Background()
	client, err := NewMainnetClient(ctx)
	requireNoError(t, err, "NewMainnetClient")
	t.Cleanup(func() {
		client.Close()
	})

	const objectID = "0x72f5c6eef73d77de271886219a2543e7c29a33de19a6c69c5cf1899f729c3f17"
	obj, err := client.GetObject(ctx, objectID, nil)
	requireNoError(t, err, "GetObject")
	requireNotNil(t, obj, "GetObject")
	requireEqual(t, obj.GetObjectId(), objectID, "GetObject id")
}
