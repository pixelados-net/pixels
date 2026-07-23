package model

import "testing"

// TestItemPlacementState verifies inventory versus room placement detection.
func TestItemPlacementState(t *testing.T) {
	inventoryItem := Item{}
	if !inventoryItem.InInventory() || inventoryItem.InRoom() {
		t.Fatalf("expected inventory item, got %#v", inventoryItem)
	}

	roomID := int64(1)
	placedItem := Item{RoomID: &roomID}
	if !placedItem.InRoom() || placedItem.InInventory() {
		t.Fatalf("expected placed item, got %#v", placedItem)
	}
}
