package service

import (
	"context"
	"errors"
	"testing"
)

// TestUpdateStateValidatesAndPersistsCompareAndSwap verifies durable state mutation behavior.
func TestUpdateStateValidatesAndPersistsCompareAndSwap(t *testing.T) {
	store := newFakeStore()
	_, err := New(store).UpdateState(context.Background(), StateParams{ItemID: 1, RoomID: 9, Expected: "0", Next: "1"})
	if err != nil {
		t.Fatalf("update state: %v", err)
	}
	if store.stateParams.ID != 1 || store.stateParams.RoomID != 9 || store.stateParams.Expected != "0" || store.stateParams.Next != "1" {
		t.Fatalf("unexpected state params %#v", store.stateParams)
	}
	store.stateUpdated = false
	_, err = New(store).UpdateState(context.Background(), StateParams{ItemID: 1, RoomID: 9, Expected: "0", Next: "1"})
	if !errors.Is(err, ErrStateConflict) {
		t.Fatalf("expected state conflict, got %v", err)
	}
}

// TestUpdateStateRejectsInvalidIdentifiers verifies state mutation input validation.
func TestUpdateStateRejectsInvalidIdentifiers(t *testing.T) {
	if _, err := New(newFakeStore()).UpdateState(context.Background(), StateParams{RoomID: 1}); !errors.Is(err, ErrInvalidItemID) {
		t.Fatalf("expected invalid item id, got %v", err)
	}
	if _, err := New(newFakeStore()).UpdateState(context.Background(), StateParams{ItemID: 1}); !errors.Is(err, ErrInvalidRoomID) {
		t.Fatalf("expected invalid room id, got %v", err)
	}
}
