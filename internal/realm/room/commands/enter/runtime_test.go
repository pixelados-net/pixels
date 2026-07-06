package enter

import (
	"testing"

	roommodel "github.com/niflaot/pixels/internal/realm/room/model"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// TestRoomSnapshotMapsRuntimeFields verifies persistent room to runtime mapping.
func TestRoomSnapshotMapsRuntimeFields(t *testing.T) {
	categoryID := int64(3)
	snapshot := roomSnapshot(roommodel.Room{
		Base:       sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}},
		CategoryID: &categoryID,
		MaxUsers:   25,
	})

	if snapshot.ID != 9 || snapshot.CategoryID == nil || *snapshot.CategoryID != 3 || snapshot.MaxUsers != 25 {
		t.Fatalf("unexpected snapshot %#v", snapshot)
	}
}
