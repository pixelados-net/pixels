package inventory

import (
	"context"
	"sync/atomic"
	"testing"

	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
)

// inventoryStore counts durable owner inventory reads.
type inventoryStore struct {
	petrecord.Store
	// calls counts durable reads.
	calls atomic.Int32
	// pets stores the fixed durable result.
	pets []petrecord.Pet
}

// Inventory returns the fixed durable result.
func (store *inventoryStore) Inventory(context.Context, int64) ([]petrecord.Pet, error) {
	store.calls.Add(1)
	return store.pets, nil
}

// TestListCachesAndInvalidates verifies warmed reads avoid PostgreSQL.
func TestListCachesAndInvalidates(t *testing.T) {
	store := &inventoryStore{pets: []petrecord.Pet{{ID: 7}}}
	service := New(store, nil)
	for range 2 {
		pets, err := service.List(context.Background(), 3)
		if err != nil || len(pets) != 1 || pets[0].ID != 7 {
			t.Fatalf("pets=%+v err=%v", pets, err)
		}
	}
	if calls := store.calls.Load(); calls != 1 {
		t.Fatalf("expected one durable read, got %d", calls)
	}
	service.Invalidate(3)
	if _, err := service.List(context.Background(), 3); err != nil {
		t.Fatal(err)
	}
	if calls := store.calls.Load(); calls != 2 {
		t.Fatalf("expected invalidated read, got %d calls", calls)
	}
}

// TestWarmedListAllocatesNothing verifies the pre-encoding hot path.
func TestWarmedListAllocatesNothing(t *testing.T) {
	service := New(&inventoryStore{pets: []petrecord.Pet{{ID: 7}}}, nil)
	if _, err := service.List(context.Background(), 3); err != nil {
		t.Fatal(err)
	}
	allocations := testing.AllocsPerRun(1000, func() {
		_, _ = service.List(context.Background(), 3)
	})
	if allocations != 0 {
		t.Fatalf("expected zero allocations, got %.2f", allocations)
	}
}

// BenchmarkInventorySnapshot measures one warmed immutable inventory read.
func BenchmarkInventorySnapshot(b *testing.B) {
	service := New(&inventoryStore{pets: []petrecord.Pet{{ID: 7}}}, nil)
	_, _ = service.List(context.Background(), 3)
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		_, _ = service.List(context.Background(), 3)
	}
}
