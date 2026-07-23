package room

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
	"github.com/niflaot/pixels/internal/realm/room/world/wired/game"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/registry"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
)

// TestViewReadsAuthoritativeRoomFacts verifies condition facts come from the active room world.
func TestViewReadsAuthoritativeRoomFacts(t *testing.T) {
	rooms, active := conditionRoom(t)
	games := game.New()
	provider := New(rooms, games, nil, nil)
	if _, found := provider.View(999); found {
		t.Fatal("missing room returned a condition view")
	}
	facts, found := provider.View(active.ID())
	if !found {
		t.Fatal("active room did not return a condition view")
	}
	view := facts.(*View)
	if view.UserCount() != 1 {
		t.Fatalf("user count=%d, want 1", view.UserCount())
	}
	if _, err := active.TeleportUnit(7, grid.MustPoint(1, 1), worldunit.RotationSouth, false); err != nil {
		t.Fatal(err)
	}
	if pass, err := view.UnitOn(10); err != nil || !pass {
		t.Fatalf("unit on target=%t err=%v", pass, err)
	}
	event := trigger.Event{ActorID: 7, PlayerID: 7}
	if pass, valid, err := view.ActorOn(event, 10); err != nil || !valid || !pass {
		t.Fatalf("actor on target=%t valid=%t err=%v", pass, valid, err)
	}
	if pass, err := view.Stacked(20); err != nil || !pass {
		t.Fatalf("stacked=%t err=%v", pass, err)
	}
	snapshot := record.Snapshot{State: "1", X: 1, Y: 1, Z: 0, Rotation: 0, Present: true}
	if pass, err := view.SnapshotMatches(10, snapshot, []int32{1, 1, 1}); err != nil || !pass {
		t.Fatalf("snapshot match=%t err=%v", pass, err)
	}
	if _, found := active.SetFurnitureExtraData(10, "2"); !found {
		t.Fatal("change snapshot target state")
	}
	if pass, err := view.SnapshotMatches(10, snapshot, []int32{1, 1, 1}); err != nil || pass {
		t.Fatalf("changed snapshot match=%t err=%v", pass, err)
	}
	if _, found := active.SetUnitEffect(7, 8); !found {
		t.Fatal("set effect")
	}
	if _, err := active.SetHandItem(7, 9); err != nil {
		t.Fatal(err)
	}
	if pass, valid, _ := view.WearingEffect(7, 8); !valid || !pass {
		t.Fatal("active effect was not visible")
	}
	if pass, valid, _ := view.HasHanditem(7, 9); !valid || !pass {
		t.Fatal("hand item was not visible")
	}
	if pass, valid, _ := view.ActorGroup(7); pass || valid {
		t.Fatal("missing social group dependency did not fail closed")
	}
	if pass, valid, _ := view.WearingBadge(7, "QA"); pass || valid {
		t.Fatal("missing badge dependency did not fail closed")
	}
	games.Start(active.ID())
	_, err := games.ExecuteGame(context.Background(), effect.JoinTeam, &configuration.Node{Parameters: configuration.Parameters{Values: []int32{2}}}, trigger.Event{RoomID: active.ID(), PlayerID: 7})
	if err != nil {
		t.Fatal(err)
	}
	if pass, valid, _ := view.ActorTeam(7, 2); !valid || !pass {
		t.Fatal("team membership was not visible")
	}
}

// TestValidMovesSimulatesFurnitureDestinations verifies the compatibility condition rejects occupied targets.
func TestValidMovesSimulatesFurnitureDestinations(t *testing.T) {
	_, active := conditionRoom(t)
	view := &View{active: active}
	node := &configuration.Node{
		Descriptor: registry.Descriptor{Key: "wf_act_move_to_dir"},
		Parameters: configuration.Parameters{Values: []int32{2}},
		Targets:    []record.Target{{ItemID: 10}},
	}
	pass, err := view.ValidMoves([]*configuration.Node{node}, trigger.Event{ActorID: 7})
	if err != nil || !pass {
		t.Fatalf("valid move=%t err=%v", pass, err)
	}
	if _, err = active.AddEntity(-1, 1, worldunit.KindBot, worldpath.Position{Point: grid.MustPoint(2, 1)}, worldunit.RotationSouth); err != nil {
		t.Fatal(err)
	}
	pass, err = view.ValidMoves([]*configuration.Node{node}, trigger.Event{ActorID: 7})
	if err != nil || pass {
		t.Fatalf("occupied move=%t err=%v", pass, err)
	}
	missing := &configuration.Node{Descriptor: registry.Descriptor{Key: "wf_act_move_rotate"}, Targets: []record.Target{{ItemID: 999}}}
	if pass, _ = view.ValidMoves([]*configuration.Node{missing}, trigger.Event{}); pass {
		t.Fatal("missing movement target passed")
	}
	if !movementEffect("wf_act_flee") || movementEffect("wf_act_show_message") || firstValue(nil) != 0 {
		t.Fatal("movement helper classification failed")
	}
	if tileDistance(grid.MustPoint(0, 0), grid.MustPoint(2, 1)) != 3 || len(adjacent(grid.MustPoint(1, 1))) != 4 {
		t.Fatal("movement geometry helper failed")
	}
}

// conditionRoom creates one active room with walkable, stacked, and movable furniture fixtures.
func conditionRoom(t *testing.T) (*roomlive.Registry, *roomlive.Room) {
	t.Helper()
	rooms := roomlive.NewRegistry(nil, roomlive.WithTickInterval(time.Hour))
	active, err := rooms.Activate(roomlive.Snapshot{ID: 77, MaxUsers: 10})
	if err != nil {
		t.Fatal(err)
	}
	roomGrid, err := grid.Parse("000\r000\r000", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	items := []worldfurniture.Item{
		{ID: 10, Point: grid.MustPoint(1, 1), ExtraData: "1", Definition: worldfurniture.Definition{Width: 1, Length: 1, AllowWalk: true, AllowStack: true}},
		{ID: 20, Point: grid.MustPoint(2, 2), Definition: worldfurniture.Definition{Width: 1, Length: 1, StackHeight: grid.HeightFromInt(1), AllowStack: true}},
		{ID: 21, Point: grid.MustPoint(2, 2), Z: grid.HeightFromInt(1), Definition: worldfurniture.Definition{Width: 1, Length: 1}},
	}
	if err = active.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Furniture: items, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}}); err != nil {
		t.Fatal(err)
	}
	if _, err = rooms.Join(context.Background(), active.ID(), roomlive.Occupant{PlayerID: 7, Username: "demo", ConnectionID: "test", ConnectionKind: "test"}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _, _ = rooms.Close(context.Background(), active.ID()) })
	return rooms, active
}
