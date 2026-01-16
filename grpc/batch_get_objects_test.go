package grpc

import (
	"context"
	"fmt"
	"testing"
)

func TestBatchGetObjects(t *testing.T) {
	ctx := context.Background()
	client, err := NewMainnetClient(ctx)
	requireNoError(t, err, "NewMainnetClient")
	t.Cleanup(func() {
		client.Close()
	})

	objectIDs := []string{
		"0x72f5c6eef73d77de271886219a2543e7c29a33de19a6c69c5cf1899f729c3f17",
		"0x57c9a3d7bdfc965ef4cb402ae0caf4f8535678d009f930910affa599facab39b",
	}
	requests := make([]ObjectRequest, len(objectIDs))
	for i, id := range objectIDs {
		requests[i] = ObjectRequest{ObjectID: id}
	}

	results, err := client.BatchGetObjects(ctx, requests, nil)
	requireNoError(t, err, "BatchGetObjects")
	requireEqual(t, len(results), len(objectIDs), "BatchGetObjects result count")

	for i, res := range results {
		requireNoError(t, res.Err, fmt.Sprintf("result %d error", i))
		requireNotNil(t, res.Object, fmt.Sprintf("result %d object", i))
		requireEqual(t, res.Object.GetObjectId(), objectIDs[i], fmt.Sprintf("result %d object id", i))
	}
}
