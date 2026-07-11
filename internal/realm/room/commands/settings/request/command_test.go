package request

import (
	"testing"

	roommodel "github.com/niflaot/pixels/internal/realm/room/model"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// TestSettingsParamsProjectsCompleteRoom verifies settings response projection.
func TestSettingsParamsProjectsCompleteRoom(t *testing.T) {
	categoryID := int64(4)
	room := roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}}, Name: "Room", Description: "Description",
		DoorMode: roommodel.DoorModePassword, CategoryID: &categoryID, MaxUsers: 25, TradeMode: roommodel.TradeModeAllowed,
		AllowPets: true, AllowPetsEat: true, AllowWalkthrough: true, HideWalls: true, WallThickness: 1,
		FloorThickness: -1, ChatMode: 1, ChatWeight: 2, ChatSpeed: 1, ChatDistance: 50, ChatProtection: 2,
		ModerationMute: 1, ModerationKick: 2, ModerationBan: 0}
	params := settingsParams(room, []roommodel.Tag{{Value: "social"}})
	if params.RoomID != 9 || params.CategoryID != 4 || len(params.Tags) != 1 || params.MaxUsersLimit != 100 || params.ModerationKick != 2 || params.ChatDistance != 50 {
		t.Fatalf("unexpected settings params %#v", params)
	}
}
