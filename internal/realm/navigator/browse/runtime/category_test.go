package runtime

import (
	"testing"

	roomoccupancy "github.com/niflaot/pixels/internal/realm/room/runtime/events/occupancychanged"
)

// TestCategoryEntryAggregatesRooms verifies category counts aggregate active rooms.
func TestCategoryEntryAggregatesRooms(t *testing.T) {
	categoryID := int64(4)
	broadcaster := NewCategoryCountBroadcaster(nil, nil)
	broadcaster.rooms[1] = roomoccupancy.Payload{RoomID: 1, CategoryID: &categoryID, Count: 2, MaxUsers: 25}
	broadcaster.rooms[2] = roomoccupancy.Payload{RoomID: 2, CategoryID: &categoryID, Count: 3, MaxUsers: 50}

	entry := broadcaster.categoryEntryLocked(4)
	if entry.CurrentVisitorCount != 5 || entry.MaxVisitorCount != 75 {
		t.Fatalf("unexpected category count %#v", entry)
	}
}

// TestSnapshotReturnsSortedCategoryCounts verifies current count snapshots.
func TestSnapshotReturnsSortedCategoryCounts(t *testing.T) {
	firstCategoryID := int64(4)
	secondCategoryID := int64(2)
	broadcaster := NewCategoryCountBroadcaster(nil, nil)
	broadcaster.rooms[1] = roomoccupancy.Payload{RoomID: 1, CategoryID: &firstCategoryID, Count: 2, MaxUsers: 25}
	broadcaster.rooms[2] = roomoccupancy.Payload{RoomID: 2, CategoryID: &secondCategoryID, Count: 3, MaxUsers: 50}

	entries := broadcaster.Snapshot()
	if len(entries) != 2 {
		t.Fatalf("expected two entries, got %#v", entries)
	}
	if entries[0].CategoryID != 2 || entries[0].CurrentVisitorCount != 3 || entries[0].MaxVisitorCount != 50 {
		t.Fatalf("unexpected first entry %#v", entries[0])
	}
	if entries[1].CategoryID != 4 || entries[1].CurrentVisitorCount != 2 || entries[1].MaxVisitorCount != 25 {
		t.Fatalf("unexpected second entry %#v", entries[1])
	}
}

// TestQueueClearsCategoryWhenRoomEmpties verifies leave updates include zeroes.
func TestQueueClearsCategoryWhenRoomEmpties(t *testing.T) {
	categoryID := int64(4)
	broadcaster := NewCategoryCountBroadcaster(nil, nil)
	broadcaster.queue(roomoccupancy.Payload{RoomID: 1, CategoryID: &categoryID, Count: 2, MaxUsers: 25})
	broadcaster.queue(roomoccupancy.Payload{RoomID: 1, CategoryID: &categoryID, Count: 0, MaxUsers: 25})
	broadcaster.Close()

	entries := broadcaster.takePending()
	if len(entries) != 1 {
		t.Fatalf("expected one pending entry, got %#v", entries)
	}
	if entries[0].CurrentVisitorCount != 0 || entries[0].MaxVisitorCount != 0 {
		t.Fatalf("expected zeroed category entry, got %#v", entries[0])
	}
}
