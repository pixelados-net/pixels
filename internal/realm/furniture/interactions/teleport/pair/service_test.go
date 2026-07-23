package pair

import (
	"context"
	"errors"
	"testing"

	"github.com/niflaot/pixels/internal/command"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// TestPairCanonicalizesAndReplaces verifies validated symmetric pairing.
func TestPairCanonicalizesAndReplaces(t *testing.T) {
	store := &storeForTest{}
	furniture := furnitureForTest()
	service := NewService(store, furniture)
	paired, err := service.Pair(context.Background(), 7, 2, 1)
	if err != nil {
		t.Fatalf("pair teleports: %v", err)
	}
	if paired.ItemOneID != 1 || paired.ItemTwoID != 2 || store.pair != paired {
		t.Fatalf("unexpected pair %#v stored=%#v", paired, store.pair)
	}
	target, definition, found, err := service.FindTarget(context.Background(), 1)
	if err != nil || !found || target.ID != 2 || definition.InteractionType != "teleport" {
		t.Fatalf("unexpected target item=%#v definition=%#v found=%v err=%v", target, definition, found, err)
	}
}

// TestPairGrantedAllowsOwnedInventoryItems verifies catalog pairing before placement.
func TestPairGrantedAllowsOwnedInventoryItems(t *testing.T) {
	store := &storeForTest{}
	furniture := furnitureForTest()
	for id, item := range furniture.items {
		item.RoomID, item.X, item.Y, item.Z = nil, nil, nil, nil
		furniture.items[id] = item
	}
	service := NewService(store, furniture)
	if _, err := service.Pair(context.Background(), 7, 1, 2); !errors.Is(err, ErrItemNotFound) {
		t.Fatalf("expected manual pairing to require placement, got %v", err)
	}
	paired, err := service.PairGranted(context.Background(), 7, 1, 2)
	if err != nil || paired != (Pair{ItemOneID: 1, ItemTwoID: 2}) {
		t.Fatalf("pair granted inventory items=%#v err=%v", paired, err)
	}
}

// TestCommandAndUnpair verify command dispatch and removal behavior.
func TestCommandAndUnpair(t *testing.T) {
	store := &storeForTest{}
	service := NewService(store, furnitureForTest())
	cmd := Command{ActorPlayerID: 7, FirstItemID: 1, SecondItemID: 2}
	if cmd.CommandName() != CommandName {
		t.Fatalf("unexpected command name %q", cmd.CommandName())
	}
	if err := (Handler{Service: service}).Handle(context.Background(), command.Envelope[Command]{Command: cmd}); err != nil {
		t.Fatalf("handle pair command: %v", err)
	}
	removed, err := service.Unpair(context.Background(), 1)
	if err != nil || !removed {
		t.Fatalf("unpair removed=%v err=%v", removed, err)
	}
	if _, err := service.Unpair(context.Background(), 0); !errors.Is(err, ErrInvalidPair) {
		t.Fatalf("expected invalid unpair, got %v", err)
	}
}

// TestPairModelValidation verifies canonical model edge cases.
func TestPairModelValidation(t *testing.T) {
	if _, err := New(1, 1); !errors.Is(err, ErrInvalidPair) {
		t.Fatalf("expected self-pair rejection, got %v", err)
	}
	paired, err := New(1, 2)
	if err != nil {
		t.Fatalf("create pair: %v", err)
	}
	if _, found := paired.Other(3); found {
		t.Fatal("expected unrelated item miss")
	}
}

// TestPairRejectsOwnershipAndKinds verifies pairing policy.
func TestPairRejectsOwnershipAndKinds(t *testing.T) {
	furniture := furnitureForTest()
	service := NewService(&storeForTest{}, furniture)
	if _, err := service.Pair(context.Background(), 8, 1, 2); !errors.Is(err, ErrNotOwner) {
		t.Fatalf("expected ownership error, got %v", err)
	}
	furniture.definitions[2] = furnituremodel.Definition{InteractionType: "default"}
	if _, err := service.Pair(context.Background(), 0, 1, 2); !errors.Is(err, ErrNotTeleport) {
		t.Fatalf("expected type error, got %v", err)
	}
}

// TestFindTargetCleansMissingPair verifies broken pairs are removed.
func TestFindTargetCleansMissingPair(t *testing.T) {
	store := &storeForTest{pair: Pair{ItemOneID: 1, ItemTwoID: 3}, found: true}
	service := NewService(store, furnitureForTest())
	_, _, found, err := service.FindTarget(context.Background(), 1)
	if err != nil || found || !store.deleted {
		t.Fatalf("expected cleaned missing target found=%v deleted=%v err=%v", found, store.deleted, err)
	}
}

// BenchmarkPairOther measures the animation-path pair lookup primitive.
func BenchmarkPairOther(b *testing.B) {
	paired := Pair{ItemOneID: 7, ItemTwoID: 9}
	b.ReportAllocs()
	for b.Loop() {
		_, _ = paired.Other(7)
	}
}

// storeForTest stores one pair in memory.
type storeForTest struct {
	// pair stores the relationship fixture.
	pair Pair
	// found reports whether the relationship exists.
	found bool
	// deleted reports cleanup execution.
	deleted bool
}

// FindByItem finds the fixture pair.
func (store *storeForTest) FindByItem(_ context.Context, _ int64) (Pair, bool, error) {
	return store.pair, store.found, nil
}

// Replace stores a relationship.
func (store *storeForTest) Replace(_ context.Context, paired Pair) error {
	store.pair, store.found = paired, true
	return nil
}

// DeleteByItem removes the fixture relationship.
func (store *storeForTest) DeleteByItem(_ context.Context, _ int64) (bool, error) {
	store.found, store.deleted = false, true
	return true, nil
}

// furnitureStoreForTest stores item and definition fixtures.
type furnitureStoreForTest struct {
	// items stores fixtures by id.
	items map[int64]furnituremodel.Item
	// definitions stores fixtures by id.
	definitions map[int64]furnituremodel.Definition
}

// furnitureForTest creates paired teleport fixtures.
func furnitureForTest() *furnitureStoreForTest {
	roomID, x, y, z := int64(9), 1, 1, 0.0
	return &furnitureStoreForTest{
		items: map[int64]furnituremodel.Item{
			1: {Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 1}}, DefinitionID: 1, OwnerPlayerID: 7, RoomID: &roomID, X: &x, Y: &y, Z: &z},
			2: {Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 2}}, DefinitionID: 2, OwnerPlayerID: 7, RoomID: &roomID, X: &x, Y: &y, Z: &z},
		},
		definitions: map[int64]furnituremodel.Definition{
			1: {InteractionType: "teleport"}, 2: {InteractionType: "teleport"},
		},
	}
}

// FindItemByID finds an item fixture.
func (store *furnitureStoreForTest) FindItemByID(_ context.Context, id int64) (furnituremodel.Item, bool, error) {
	item, found := store.items[id]
	return item, found, nil
}

// FindDefinitionByID finds a definition fixture.
func (store *furnitureStoreForTest) FindDefinitionByID(_ context.Context, id int64) (furnituremodel.Definition, bool, error) {
	definition, found := store.definitions[id]
	return definition, found, nil
}
