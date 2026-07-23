package avatar

import (
	"context"
	"testing"
	"time"

	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/effect"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
)

// TestExpandUsesClosedRuntimeTokenCatalog verifies safe compatibility substitutions.
func TestExpandUsesClosedRuntimeTokenCatalog(t *testing.T) {
	actual := expand("Hello %username%: %online% online, %roomsloaded% rooms, %unknown%", trigger.Event{Username: "demo"}, 17, 4)
	want := "Hello demo: 17 online, 4 rooms, %unknown%"
	if actual != want {
		t.Fatalf("expand=%q, want %q", actual, want)
	}
}

// TestExecuteAvatarUsesRoomScopedPrimitives verifies avatar effects mutate only the resolved live unit.
func TestExecuteAvatarUsesRoomScopedPrimitives(t *testing.T) {
	rooms, active := avatarRoom(t)
	service := New(rooms, nil, nil, nil, nil)
	event := trigger.Event{ID: 5, RoomID: active.ID(), ActorID: 7, PlayerID: 7, Username: "demo"}
	result, err := service.ExecuteAvatar(context.Background(), effect.GiveHanditem, &configuration.Node{Parameters: configuration.Parameters{Number: 9}}, event)
	if err != nil || result.Status != effect.Applied {
		t.Fatalf("hand item result=%+v err=%v", result, err)
	}
	unit, _ := active.Unit(7)
	if unit.HandItem != 9 {
		t.Fatalf("hand item=%d, want 9", unit.HandItem)
	}
	result, err = service.ExecuteAvatar(context.Background(), effect.GiveEffect, &configuration.Node{Parameters: configuration.Parameters{Number: 8}}, event)
	if err != nil || result.Status != effect.Applied {
		t.Fatalf("effect result=%+v err=%v", result, err)
	}
	unit, _ = active.Unit(7)
	if unit.ActiveEffectID != 8 {
		t.Fatalf("effect=%d, want 8", unit.ActiveEffectID)
	}
	node := &configuration.Node{Targets: []record.Target{{ItemID: 10}}}
	result, err = service.ExecuteAvatar(context.Background(), effect.TeleportAvatar, node, event)
	if err != nil || result.Status != effect.Applied {
		t.Fatalf("teleport result=%+v err=%v", result, err)
	}
	unit, _ = active.Unit(7)
	if unit.Position.Point != grid.MustPoint(2, 0) {
		t.Fatalf("safe teleport point=%+v", unit.Position.Point)
	}
	roomPath, err := active.MoveTo(7, grid.MustPoint(1, 0))
	if err != nil || roomPath.Len() == 0 {
		t.Fatalf("walk after teleport path=%+v err=%v", roomPath, err)
	}
}

// TestExecuteAvatarFailsClosed verifies absent dependencies and invalid values never become applied effects.
func TestExecuteAvatarFailsClosed(t *testing.T) {
	rooms, active := avatarRoom(t)
	service := New(rooms, nil, nil, nil, nil)
	event := trigger.Event{RoomID: active.ID(), ActorID: 7, PlayerID: 7}
	cases := []struct {
		operation effect.AvatarOperation
		node      *configuration.Node
		status    effect.Status
	}{
		{operation: effect.ShowMessage, node: &configuration.Node{Parameters: configuration.Parameters{Text: "hello"}}, status: effect.Skipped},
		{operation: effect.AlertAvatar, node: &configuration.Node{Parameters: configuration.Parameters{Text: "hello"}}, status: effect.Skipped},
		{operation: effect.KickAvatar, node: &configuration.Node{}, status: effect.Blocked},
		{operation: effect.MuteAvatar, node: &configuration.Node{Parameters: configuration.Parameters{Values: []int32{1}}}, status: effect.Blocked},
		{operation: effect.GiveRespect, node: &configuration.Node{Parameters: configuration.Parameters{Number: 1}}, status: effect.Blocked},
		{operation: effect.GiveHanditem, node: &configuration.Node{Parameters: configuration.Parameters{Number: 10000}}, status: effect.Blocked},
		{operation: effect.GiveEffect, node: &configuration.Node{Parameters: configuration.Parameters{Number: 10001}}, status: effect.Blocked},
		{operation: effect.AvatarOperation(255), node: &configuration.Node{}, status: effect.Blocked},
	}
	for _, testCase := range cases {
		result, err := service.ExecuteAvatar(context.Background(), testCase.operation, testCase.node, event)
		if err != nil || result.Status != testCase.status {
			t.Errorf("operation=%d result=%+v err=%v", testCase.operation, result, err)
		}
	}
	result, err := service.ExecuteAvatar(context.Background(), effect.ShowMessage, &configuration.Node{}, trigger.Event{RoomID: 999, PlayerID: 7})
	if err != nil || result.Status != effect.Skipped {
		t.Fatalf("missing room result=%+v err=%v", result, err)
	}
}

// avatarRoom creates one active player and one WIRED teleport target.
func avatarRoom(t *testing.T) (*roomlive.Registry, *roomlive.Room) {
	t.Helper()
	rooms := roomlive.NewRegistry(nil, roomlive.WithTickInterval(time.Hour))
	active, err := rooms.Activate(roomlive.Snapshot{ID: 88, MaxUsers: 5})
	if err != nil {
		t.Fatal(err)
	}
	roomGrid, err := grid.Parse("000", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	item := worldfurniture.Item{ID: 10, Point: grid.MustPoint(2, 0), Definition: worldfurniture.Definition{Width: 1, Length: 1, AllowWalk: true}}
	if err = active.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Furniture: []worldfurniture.Item{item}, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}, Body: worldunit.RotationSouth, Head: worldunit.RotationSouth}); err != nil {
		t.Fatal(err)
	}
	if _, err = rooms.Join(context.Background(), active.ID(), roomlive.Occupant{PlayerID: 7, Username: "demo", ConnectionID: "test", ConnectionKind: "test"}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _, _ = rooms.Close(context.Background(), active.ID()) })
	return rooms, active
}
