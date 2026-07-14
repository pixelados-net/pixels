package action

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	actionservice "github.com/niflaot/pixels/internal/realm/room/world/action"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// TestHandleCoordinatesSupportedActions verifies every avatar action family.
func TestHandleCoordinatesSupportedActions(t *testing.T) {
	handler, player, active := commandFixture(t)
	if err := player.EnterRoom(9); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	connection := commandConnection()
	commands := []Command{
		{Handler: connection, Kind: KindDance, Value: 3},
		{Handler: connection, Kind: KindGesture, Value: 2},
		{Handler: connection, Kind: KindGesture, Value: 5},
		{Handler: connection, Kind: KindSign, Value: 18},
		{Handler: connection, Kind: KindPosture, Value: 1},
	}
	for _, current := range commands {
		if err := handler.Handle(ctx, command.Envelope[Command]{Command: current}); err != nil {
			t.Fatalf("handle %#v: %v", current, err)
		}
	}
	unit, _ := active.Unit(7)
	if !unit.Idle || commandStatus(unit, worldunit.StatusSign) != "18" || commandStatus(unit, worldunit.StatusSit) == "" {
		t.Fatalf("unexpected final unit %#v", unit)
	}
	if err := handler.Handle(ctx, command.Envelope[Command]{Command: Command{Handler: connection, Kind: KindGesture, Value: 5}}); err != nil {
		t.Fatal(err)
	}
	unit, _ = active.Unit(7)
	if unit.Idle {
		t.Fatal("expected second manual idle action to resume")
	}
}

// TestHandleIgnoresInvalidActions verifies malformed action ids are soft failures.
func TestHandleIgnoresInvalidActions(t *testing.T) {
	handler, player, active := commandFixture(t)
	if err := player.EnterRoom(9); err != nil {
		t.Fatal(err)
	}
	connection := commandConnection()
	tests := []Command{
		{Handler: connection, Kind: KindDance, Value: 6},
		{Handler: connection, Kind: KindGesture, Value: 4},
		{Handler: connection, Kind: KindSign, Value: 19},
		{Handler: connection, Kind: KindPosture, Value: 3},
		{Handler: connection, Kind: Kind(99), Value: 1},
	}
	for _, current := range tests {
		if err := handler.Handle(context.Background(), command.Envelope[Command]{Command: current}); err != nil {
			t.Fatalf("expected soft failure for %#v: %v", current, err)
		}
	}
	unit, _ := active.Unit(7)
	if len(unit.Statuses) != 0 {
		t.Fatalf("invalid actions changed unit %#v", unit)
	}
}

// TestHandleValidatesSessionRoomAndUnit verifies stale command boundaries.
func TestHandleValidatesSessionRoomAndUnit(t *testing.T) {
	handler, player, active := commandFixture(t)
	envelope := command.Envelope[Command]{Command: Command{Handler: commandConnection(), Kind: KindDance, Value: 1}}
	if err := handler.Handle(context.Background(), envelope); !errors.Is(err, ErrPlayerNotInRoom) {
		t.Fatalf("expected no room, got %v", err)
	}
	if err := player.EnterRoom(10); err != nil {
		t.Fatal(err)
	}
	if err := handler.Handle(context.Background(), envelope); !errors.Is(err, roomlive.ErrRoomNotFound) {
		t.Fatalf("expected missing room, got %v", err)
	}
	player.LeaveRoom()
	if err := player.EnterRoom(9); err != nil {
		t.Fatal(err)
	}
	active.Leave(7)
	if err := handler.Handle(context.Background(), envelope); !errors.Is(err, roomlive.ErrUnitNotFound) {
		t.Fatalf("expected missing unit, got %v", err)
	}
}

// TestCommandHelpers verifies command identity and compact sign formatting.
func TestCommandHelpers(t *testing.T) {
	if (Command{}).CommandName() != Name || !validGesture(7) || validGesture(8) {
		t.Fatal("unexpected command helpers")
	}
	if stringValue(7) != "7" || stringValue(18) != "18" {
		t.Fatal("unexpected compact values")
	}
}

// commandFixture creates one authenticated player and active room.
func commandFixture(t testing.TB) (Handler, *playerlive.Player, *roomlive.Room) {
	t.Helper()
	players := playerlive.NewRegistry()
	peer, err := playerlive.NewSessionPeer("conn", "websocket", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	player, err := playerlive.NewPlayer(playerlive.Snapshot{ID: 7, Username: "demo"}, peer)
	if err != nil {
		t.Fatal(err)
	}
	if err = players.Add(player); err != nil {
		t.Fatal(err)
	}
	bindings := binding.NewRegistry()
	if err = bindings.Add(binding.Binding{PlayerID: 7, ConnectionID: "conn", ConnectionKind: "websocket"}); err != nil {
		t.Fatal(err)
	}
	runtime := roomlive.NewRegistry(nil)
	active, err := runtime.Activate(roomlive.Snapshot{ID: 9, MaxUsers: 5})
	if err != nil {
		t.Fatal(err)
	}
	roomGrid, err := grid.Parse("00", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if err = active.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}, Body: worldunit.RotationSouth, Head: worldunit.RotationSouth}); err != nil {
		t.Fatal(err)
	}
	if _, err = runtime.Join(context.Background(), 9, roomlive.Occupant{PlayerID: 7, Username: "demo", ConnectionID: "conn", ConnectionKind: "websocket"}); err != nil {
		t.Fatal(err)
	}
	return Handler{Players: players, Bindings: bindings, Runtime: runtime, Actions: actionservice.New(nil, nil)}, player, active
}

// commandConnection creates one matching packet context.
func commandConnection() netconn.Context {
	return netconn.Context{ConnectionID: "conn", ConnectionKind: "websocket"}
}

// commandStatus returns one status from a stable unit snapshot.
func commandStatus(unit roomlive.UnitSnapshot, key string) string {
	for _, status := range unit.Statuses {
		if status.Key == key {
			return status.Value
		}
	}
	return ""
}
