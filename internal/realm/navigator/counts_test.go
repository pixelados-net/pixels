package navigator

import (
	"testing"

	roomoccupancy "github.com/niflaot/pixels/internal/realm/room/events/occupancychanged"
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
