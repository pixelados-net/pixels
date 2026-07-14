package effect

import (
	"context"
	"testing"
	"time"
)

// memoryStore is a focused effect service fixture.
type memoryStore struct {
	// effects stores stacks by player and effect id.
	effects map[int64]map[int32]Effect
	// active stores selected effects.
	active map[int64]*int32
	// expired stores the next sweep result.
	expired []Expiration
}

// newMemoryStore creates an empty effect fixture.
func newMemoryStore() *memoryStore {
	return &memoryStore{effects: make(map[int64]map[int32]Effect), active: make(map[int64]*int32)}
}

// WithinTransaction executes fixture work immediately.
func (store *memoryStore) WithinTransaction(ctx context.Context, work func(context.Context) error) error {
	return work(ctx)
}

// List returns fixture effects.
func (store *memoryStore) List(_ context.Context, playerID int64) ([]Effect, error) {
	result := make([]Effect, 0, len(store.effects[playerID]))
	for _, item := range store.effects[playerID] {
		result = append(result, item)
	}
	return result, nil
}

// Grant creates or increments one fixture stack.
func (store *memoryStore) Grant(_ context.Context, playerID int64, effectID int32, duration int32) (Effect, error) {
	if store.effects[playerID] == nil {
		store.effects[playerID] = make(map[int32]Effect)
	}
	item := store.effects[playerID][effectID]
	item.PlayerID, item.ID, item.DurationSeconds = playerID, effectID, duration
	if item.RemainingCharges < MaximumCharges {
		item.RemainingCharges++
	}
	store.effects[playerID][effectID] = item
	return item, nil
}

// Activate starts one fixture charge.
func (store *memoryStore) Activate(_ context.Context, playerID int64, effectID int32, now time.Time) (Effect, bool, error) {
	item, found := store.effects[playerID][effectID]
	if !found {
		return Effect{}, false, nil
	}
	if item.ActivatedAt == nil {
		item.ActivatedAt = &now
	}
	store.effects[playerID][effectID] = item
	return item, true, nil
}

// SetActive replaces one fixture selection.
func (store *memoryStore) SetActive(_ context.Context, playerID int64, effectID *int32) error {
	store.active[playerID] = effectID
	return nil
}

// Active returns one fixture selection.
func (store *memoryStore) Active(_ context.Context, playerID int64) (*int32, error) {
	return store.active[playerID], nil
}

// Revoke deletes one fixture stack.
func (store *memoryStore) Revoke(_ context.Context, playerID int64, effectID int32) (bool, error) {
	_, found := store.effects[playerID][effectID]
	delete(store.effects[playerID], effectID)
	return found, nil
}

// Expire returns one prepared fixture batch.
func (store *memoryStore) Expire(context.Context, time.Time, int32) ([]Expiration, error) {
	return store.expired, nil
}

// TestGrantEnableAndRevoke verifies the complete durable effect lifecycle.
func TestGrantEnableAndRevoke(t *testing.T) {
	store := newMemoryStore()
	service := New(store, nil, nil, nil, nil, nil)
	granted, err := service.GrantEnabled(context.Background(), 7, 101, 60, SourceCatalog)
	if err != nil || granted.RemainingCharges != 1 {
		t.Fatalf("granted=%#v err=%v", granted, err)
	}
	if store.active[7] == nil || *store.active[7] != 101 || store.effects[7][101].ActivatedAt == nil {
		t.Fatalf("active=%v err=%v", store.active[7], err)
	}
	if err = service.Revoke(context.Background(), 7, 101); err != nil || store.active[7] != nil {
		t.Fatalf("active=%v err=%v", store.active[7], err)
	}
}

// TestGrantRejectsInvalidValues verifies administrative validation.
func TestGrantRejectsInvalidValues(t *testing.T) {
	service := New(newMemoryStore(), nil, nil, nil, nil, nil)
	if _, err := service.Grant(context.Background(), 0, 1, 1, SourceAdmin); err != ErrInvalidEffect {
		t.Fatalf("expected invalid effect, got %v", err)
	}
}

// TestGrantCapsCharges verifies repeated grants cannot overflow Nitro's stack bound.
func TestGrantCapsCharges(t *testing.T) {
	service := New(newMemoryStore(), nil, nil, nil, nil, nil)
	var effect Effect
	var err error
	for range MaximumCharges + 10 {
		effect, err = service.Grant(context.Background(), 7, 101, 60, SourceAdmin)
		if err != nil {
			t.Fatal(err)
		}
	}
	if effect.RemainingCharges != MaximumCharges {
		t.Fatalf("expected %d charges, got %d", MaximumCharges, effect.RemainingCharges)
	}
}

// TestEnableZeroDisablesSelection verifies Nitro's non-positive selection wire.
func TestEnableZeroDisablesSelection(t *testing.T) {
	store := newMemoryStore()
	service := New(store, nil, nil, nil, nil, nil)
	if _, err := service.Grant(context.Background(), 7, 101, 60, SourceAdmin); err != nil {
		t.Fatal(err)
	}
	if err := service.Enable(context.Background(), 7, 101); err != nil {
		t.Fatal(err)
	}
	if err := service.Enable(context.Background(), 7, 0); err != nil {
		t.Fatal(err)
	}
	if store.active[7] != nil {
		t.Fatalf("expected disabled selection, got %v", store.active[7])
	}
}

// TestActivationDoesNotRestartCharge verifies repeated activation preserves its deadline.
func TestActivationDoesNotRestartCharge(t *testing.T) {
	store := newMemoryStore()
	service := New(store, nil, nil, nil, nil, nil)
	service.now = func() time.Time { return time.Unix(100, 0) }
	if _, err := service.Grant(context.Background(), 7, 101, 60, SourceAdmin); err != nil {
		t.Fatal(err)
	}
	first, err := service.Activate(context.Background(), 7, 101)
	if err != nil {
		t.Fatal(err)
	}
	service.now = func() time.Time { return time.Unix(120, 0) }
	second, err := service.Activate(context.Background(), 7, 101)
	if err != nil || first.ActivatedAt == nil || second.ActivatedAt == nil || !first.ActivatedAt.Equal(*second.ActivatedAt) {
		t.Fatalf("first=%#v second=%#v err=%v", first, second, err)
	}
}

// TestRevokeMissingEffectReportsNotFound verifies administrative deletion semantics.
func TestRevokeMissingEffectReportsNotFound(t *testing.T) {
	service := New(newMemoryStore(), nil, nil, nil, nil, nil)
	if err := service.Revoke(context.Background(), 7, 101); err != ErrEffectNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}

// BenchmarkListEffect measures the protocol hot projection.
func BenchmarkListEffect(b *testing.B) {
	now := time.Unix(100, 0)
	activated := now.Add(-time.Second)
	item := Effect{ID: 101, DurationSeconds: 60, ActivatedAt: &activated, RemainingCharges: 2}
	b.ReportAllocs()
	for b.Loop() {
		_ = listEffect(item, now)
	}
}
