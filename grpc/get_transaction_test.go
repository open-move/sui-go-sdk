package grpc

import (
	"context"
	"testing"
)

func TestGetTransaction(t *testing.T) {
	ctx := context.Background()
	client, err := NewMainnetClient(ctx)
	requireNoError(t, err, "NewMainnetClient")
	t.Cleanup(func() {
		client.Close()
	})

	const expectedDigest = "3HZq1gEnF4sr5MTkRCirAapw3YaqgiwhWbjJdcqXmPra"
	tx, err := client.GetTransaction(ctx, expectedDigest, nil)
	requireNoError(t, err, "GetTransaction")
	requireNotNil(t, tx, "GetTransaction")
	requireEqual(t, tx.GetDigest(), expectedDigest, "GetTransaction digest")
}
