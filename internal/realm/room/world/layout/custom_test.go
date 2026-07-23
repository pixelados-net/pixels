package layout

import (
	"context"
	"errors"
	"testing"
)

// TestResolveForRoomPrefersCustomLayout verifies room geometry fallback order.
func TestResolveForRoomPrefersCustomLayout(t *testing.T) {
	store := newFakeStore()
	store.custom = Layout{RoomID: 9, Name: "model_a", Heightmap: "10", DoorX: 0, DoorY: 0, DoorDirection: 4, WallHeight: -1}
	store.customFound = true
	roomLayout, err := NewService(store).ResolveForRoom(context.Background(), 9, "model_a")
	if err != nil || roomLayout.RoomID != 9 || roomLayout.DoorZ != 1 || roomLayout.TileSize != 2 {
		t.Fatalf("unexpected custom layout %#v err=%v", roomLayout, err)
	}
}

// TestResolveForRoomFallsBackToFixedLayout verifies rooms without custom geometry.
func TestResolveForRoomFallsBackToFixedLayout(t *testing.T) {
	roomLayout, err := NewService(newFakeStore()).ResolveForRoom(context.Background(), 9, "model_a")
	if err != nil || roomLayout.ID != 7 {
		t.Fatalf("unexpected fixed layout %#v err=%v", roomLayout, err)
	}
}

// TestSaveCustomNormalizesDerivedFields verifies custom persistence projection.
func TestSaveCustomNormalizesDerivedFields(t *testing.T) {
	store := newFakeStore()
	roomLayout, err := NewService(store).SaveCustom(context.Background(), CustomSaveParams{
		RoomID: 9, Heightmap: "20", DoorX: 0, DoorY: 0, DoorDirection: 2, WallHeight: -1,
	})
	if err != nil || roomLayout.DoorZ != 2 || roomLayout.TileSize != 2 || !store.transactionCalled {
		t.Fatalf("unexpected custom layout %#v transaction=%v err=%v", roomLayout, store.transactionCalled, err)
	}
}

// TestRoomLayoutHelpersCoverFallbackAndTransactions verifies focused room layout contracts.
func TestRoomLayoutHelpersCoverFallbackAndTransactions(t *testing.T) {
	store := newFakeStore()
	service := NewService(store)
	if err := service.WithinTransaction(context.Background(), func(context.Context) error { return nil }); err != nil {
		t.Fatalf("run transaction: %v", err)
	}
	listed, err := service.List(context.Background())
	if err != nil || len(listed) != 1 {
		t.Fatalf("list layouts %#v err=%v", listed, err)
	}
	roomLayout, err := ResolveForRoom(context.Background(), managerOnlyForTest{Manager: service}, 9, "model_a")
	if err != nil || roomLayout.ID != 7 {
		t.Fatalf("resolve fixed fallback %#v err=%v", roomLayout, err)
	}
	store.custom = Layout{RoomID: 9, Name: "model_a", Heightmap: "0", DoorX: 0, DoorY: 0, WallHeight: -1}
	store.customFound = true
	roomLayout, err = ResolveForRoom(context.Background(), service, 9, "model_a")
	if err != nil || roomLayout.RoomID != 9 {
		t.Fatalf("resolve custom helper %#v err=%v", roomLayout, err)
	}
	if _, err = service.SaveCustom(context.Background(), CustomSaveParams{}); !errors.Is(err, ErrInvalidLayoutID) {
		t.Fatalf("expected invalid custom room id, got %v", err)
	}
}

// TestCustomOperationsRejectUnsupportedStore verifies conservative custom layout boundaries.
func TestCustomOperationsRejectUnsupportedStore(t *testing.T) {
	service := NewService(fixedStoreForTest{Store: newFakeStore()})
	if _, err := service.SaveCustom(context.Background(), CustomSaveParams{RoomID: 9}); !errors.Is(err, ErrCustomLayoutsUnsupported) {
		t.Fatalf("expected unsupported save, got %v", err)
	}
	if err := service.WithinTransaction(context.Background(), func(context.Context) error { return nil }); !errors.Is(err, ErrCustomLayoutsUnsupported) {
		t.Fatalf("expected unsupported transaction, got %v", err)
	}
}
