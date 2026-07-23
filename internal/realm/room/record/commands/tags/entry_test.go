package tags

import (
	"testing"

	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
)

// TestTagEntriesMapsRoomTags verifies room tag protocol entries.
func TestTagEntriesMapsRoomTags(t *testing.T) {
	entries := tagEntries([]roommodel.Tag{{Value: "games"}, {Value: "chat"}})
	if len(entries) != 2 || entries[0].Tag != "games" || entries[0].Count != 1 {
		t.Fatalf("unexpected entries %#v", entries)
	}
}
