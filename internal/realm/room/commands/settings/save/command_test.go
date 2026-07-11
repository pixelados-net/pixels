package save

import (
	"testing"

	roommodel "github.com/niflaot/pixels/internal/realm/room/model"
)

// TestUpdateParamsMapsCompleteProtocolInput verifies the packet-to-service boundary.
func TestUpdateParamsMapsCompleteProtocolInput(t *testing.T) {
	input := Command{Name: "Room", Description: "Description", CategoryID: 4, Tags: []string{"social"}, MaxUsers: 25,
		DoorMode: 2, Password: "1234", TradeMode: 2, AllowPets: true, AllowPetsEat: true,
		AllowWalkthrough: true, HideWalls: true, WallThickness: 1, FloorThickness: -1,
		ModerationMute: 1, ModerationKick: 2, ModerationBan: 0, ChatMode: 1, ChatWeight: 2,
		ChatSpeed: 1, ChatDistance: 50, ChatProtection: 2}
	params := updateParams(input, true)
	if params.Name == nil || *params.Name != "Room" || params.CategoryID == nil || *params.CategoryID == nil || **params.CategoryID != 4 || params.Password == nil || *params.Password != "1234" {
		t.Fatalf("unexpected params %#v", params)
	}
	if params.DoorMode == nil || *params.DoorMode != roommodel.DoorModePassword || params.ModerationKick == nil || *params.ModerationKick != 2 || !params.AllowReservedTags {
		t.Fatalf("unexpected typed params %#v", params)
	}
}

// TestUpdateParamsClearsCategoryAndPreservesPassword verifies optional protocol fields.
func TestUpdateParamsClearsCategoryAndPreservesPassword(t *testing.T) {
	params := updateParams(Command{}, false)
	if params.CategoryID == nil || *params.CategoryID != nil || params.Password != nil {
		t.Fatalf("unexpected optional params %#v", params)
	}
}

// TestRestrictedFieldChangesClassifyHCAndModeration verifies focused policy gates.
func TestRestrictedFieldChangesClassifyHCAndModeration(t *testing.T) {
	room := roommodel.Room{WallThickness: 0, FloorThickness: 0, ModerationMute: roommodel.ModerationPolicyOwnerOnly}
	if clubFieldsChanged(room, Command{}) || moderationFieldsChanged(room, Command{ModerationMute: int32(roommodel.ModerationPolicyOwnerOnly)}) {
		t.Fatal("expected unchanged restricted fields")
	}
	if !clubFieldsChanged(room, Command{HideWalls: true}) {
		t.Fatal("expected hide walls to require club")
	}
	if clubFieldsAllowed(room, Command{HideWalls: true}, false, false) {
		t.Fatal("expected non-club owner rejection")
	}
	if !clubFieldsAllowed(room, Command{HideWalls: true}, true, false) || !clubFieldsAllowed(room, Command{HideWalls: true}, false, true) {
		t.Fatal("expected club and global-manager access")
	}
	if !moderationFieldsChanged(room, Command{ModerationMute: int32(roommodel.ModerationPolicyOwnerAndRights)}) {
		t.Fatal("expected moderation policy change")
	}
}

// BenchmarkRestrictedFieldGates measures HC and moderation classification.
func BenchmarkRestrictedFieldGates(b *testing.B) {
	room := roommodel.Room{ModerationMute: roommodel.ModerationPolicyOwnerOnly}
	input := Command{HideWalls: true, ModerationMute: int32(roommodel.ModerationPolicyOwnerAndRights)}
	b.ReportAllocs()
	for b.Loop() {
		if clubFieldsAllowed(room, input, false, false) || !moderationFieldsChanged(room, input) {
			b.Fatal("unexpected restricted-field decision")
		}
	}
}
