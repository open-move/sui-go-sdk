package grpc

import "testing"

func requireNoError(t *testing.T, err error, context string) {
	t.Helper()
	if err != nil {
		if context != "" {
			t.Fatalf("%s: %v", context, err)
		}
		t.Fatalf("unexpected error: %v", err)
	}
}

func requireNotNil[T comparable](t *testing.T, value T, context string) {
	t.Helper()
	var zero T
	if value == zero {
		if context != "" {
			t.Fatalf("%s: got nil", context)
		}
		t.Fatal("got nil")
	}
}

func requireEqual[T comparable](t *testing.T, got, want T, context string) {
	t.Helper()
	if got != want {
		if context != "" {
			t.Fatalf("%s: got %v, want %v", context, got, want)
		}
		t.Fatalf("got %v, want %v", got, want)
	}
}
