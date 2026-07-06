package routes

import (
	"testing"

	roomlive "github.com/niflaot/pixels/internal/realm/room/live"
	roommodel "github.com/niflaot/pixels/internal/realm/room/model"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// TestRoomResponseMapsSafeFields verifies room metadata response mapping.
func TestRoomResponseMapsSafeFields(t *testing.T) {
	room := roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}}, OwnerPlayerID: 7, Name: "Demo", MaxUsers: 25}
	response := roomResponse(room)

	if response.ID != 9 || response.OwnerPlayerID != 7 || response.Name != "Demo" || response.MaxUsers != 25 {
		t.Fatalf("unexpected room response %#v", response)
	}
}

// TestOccupancyResponseCopiesPlayerIDs verifies occupancy response mapping.
func TestOccupancyResponseCopiesPlayerIDs(t *testing.T) {
	response := occupancyResponse(roomlive.Occupancy{RoomID: 9, Count: 1, MaxUsers: 25, PlayerIDs: []int64{7}})
	if response.RoomID != 9 || response.Count != 1 || response.PlayerIDs[0] != 7 {
		t.Fatalf("unexpected occupancy response %#v", response)
	}
}
