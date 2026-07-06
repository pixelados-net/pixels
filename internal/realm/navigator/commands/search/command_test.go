package search

import (
	"testing"

	roommodel "github.com/niflaot/pixels/internal/realm/room/model"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// TestCardsMapsRoomRecords verifies search card projection.
func TestCardsMapsRoomRecords(t *testing.T) {
	cards := Handler{}.cards([]roommodel.Room{{
		Base:          sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}},
		OwnerPlayerID: 7,
		OwnerName:     "demo",
		Name:          "Demo Room",
		MaxUsers:      25,
	}})

	if len(cards) != 1 || cards[0].RoomID != 9 || cards[0].Ranking != 1 {
		t.Fatalf("unexpected cards %#v", cards)
	}
}
