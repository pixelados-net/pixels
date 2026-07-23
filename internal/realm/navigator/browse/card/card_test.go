package projection

import (
	"testing"

	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// TestRoomCardMapsRoomRecord verifies room-card projection fields.
func TestRoomCardMapsRoomRecord(t *testing.T) {
	categoryID := int64(12)
	room := roommodel.Room{
		Base:          sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}},
		OwnerPlayerID: 7,
		OwnerName:     "demo",
		Name:          "Demo Room",
		Description:   "hello",
		DoorMode:      roommodel.DoorModeOpen,
		MaxUsers:      25,
		Score:         3,
		CategoryID:    &categoryID,
		TradeMode:     roommodel.TradeModeAllowed,
		AllowPets:     true,
	}

	card := RoomCard(room, 4, 2, []string{"fun"})
	if card.RoomID != 9 || card.OwnerID != 7 || card.UserCount != 4 || card.CategoryID != 12 {
		t.Fatalf("unexpected card %#v", card)
	}
	if !card.ShowOwner || !card.AllowPets || len(card.Tags) != 1 {
		t.Fatalf("unexpected card flags %#v", card)
	}
}
