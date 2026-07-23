package place

import (
	"context"
	"testing"
)

// TestDeferProjectionsWaitsForCompletion verifies transaction-owned effects never run before commit.
func TestDeferProjectionsWaitsForCompletion(t *testing.T) {
	ctx, complete := DeferProjections(context.Background())
	runs := 0
	if err := project(ctx, func(context.Context) error { runs++; return nil }); err != nil {
		t.Fatal(err)
	}
	if runs != 0 {
		t.Fatalf("projection ran before completion: %d", runs)
	}
	if err := complete(context.Background()); err != nil {
		t.Fatal(err)
	}
	if runs != 1 {
		t.Fatalf("projection runs=%d", runs)
	}
}

// TestProjectRunsImmediatelyWithoutQueue verifies ordinary inventory placement behavior is unchanged.
func TestProjectRunsImmediatelyWithoutQueue(t *testing.T) {
	runs := 0
	if err := project(context.Background(), func(context.Context) error { runs++; return nil }); err != nil {
		t.Fatal(err)
	}
	if runs != 1 {
		t.Fatalf("projection runs=%d", runs)
	}
}

// BenchmarkDeferredProjection measures one queued placement completion.
func BenchmarkDeferredProjection(b *testing.B) {
	for b.Loop() {
		ctx, complete := DeferProjections(context.Background())
		_ = project(ctx, func(context.Context) error { return nil })
		_ = complete(context.Background())
	}
}
