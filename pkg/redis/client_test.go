package redis

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
)

// TestClientOperations verifies basic Redis storage operations.
func TestClientOperations(t *testing.T) {
	server := miniredis.RunT(t)
	client := New(Config{Address: server.Addr()})
	defer func() {
		if err := client.Close(); err != nil {
			t.Fatalf("close client: %v", err)
		}
	}()

	ctx := context.Background()
	key := "pixels:test"
	value := []byte("payload")

	if err := client.Set(ctx, key, value, time.Minute); err != nil {
		t.Fatalf("set value: %v", err)
	}

	found, ok, err := client.Find(ctx, key)
	if err != nil {
		t.Fatalf("find value: %v", err)
	}

	if !ok {
		t.Fatal("expected value to exist")
	}

	if string(found) != string(value) {
		t.Fatalf("expected value %q, got %q", value, found)
	}

	if err := client.Expire(ctx, key, time.Second); err != nil {
		t.Fatalf("expire value: %v", err)
	}

	if err := client.Delete(ctx, key); err != nil {
		t.Fatalf("delete value: %v", err)
	}

	_, ok, err = client.Find(ctx, key)
	if err != nil {
		t.Fatalf("find missing value: %v", err)
	}

	if ok {
		t.Fatal("expected deleted value to be missing")
	}
}

// TestClientTake verifies atomic read and delete.
func TestClientTake(t *testing.T) {
	server := miniredis.RunT(t)
	client := New(Config{Address: server.Addr()})
	defer func() {
		if err := client.Close(); err != nil {
			t.Fatalf("close client: %v", err)
		}
	}()

	ctx := context.Background()
	key := "pixels:take"
	value := []byte("payload")

	if err := client.Set(ctx, key, value, time.Minute); err != nil {
		t.Fatalf("set value: %v", err)
	}

	found, ok, err := client.Take(ctx, key)
	if err != nil {
		t.Fatalf("take value: %v", err)
	}

	if !ok || string(found) != string(value) {
		t.Fatalf("expected taken value %q, got %q", value, found)
	}

	_, ok, err = client.Take(ctx, key)
	if err != nil {
		t.Fatalf("take missing value: %v", err)
	}

	if ok {
		t.Fatal("expected taken value to be missing")
	}
}
