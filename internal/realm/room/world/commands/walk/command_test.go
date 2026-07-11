package walk

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// TestHandleMovesPlayer verifies walk command handling.
func TestHandleMovesPlayer(t *testing.T) {
	handler, player := handlerForTest(t)
	if err := player.EnterRoom(9); err != nil {
		t.Fatalf("enter room: %v", err)
	}

	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: connectionForTest(), X: 1, Y: 0},
	})
	if err != nil {
		t.Fatalf("handle walk: %v", err)
	}
	room, _ := handler.Runtime.Find(9)
	if units := room.Units(); len(units) != 1 || !units[0].Moving {
		t.Fatalf("expected moving unit %#v", units)
	}
}

// TestHandleFacesOccupiedTarget verifies occupied targets do not disconnect clients.
func TestHandleFacesOccupiedTarget(t *testing.T) {
	handler, player := handlerForTest(t)
	if err := player.EnterRoom(9); err != nil {
		t.Fatalf("enter room: %v", err)
	}
	connections := netconn.NewRegistry()
	sent := registeredConnectionForWalkTest(t, connections, "conn")
	handler.Connections = connections
	room, _ := handler.Runtime.Find(9)
	if _, err := room.Join(roomlive.Occupant{PlayerID: 8, Username: "other", ConnectionID: "other", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("join other: %v", err)
	}
	if _, err := room.MoveTo(8, grid.MustPoint(1, 0)); err != nil {
		t.Fatalf("move other: %v", err)
	}
	if movements := room.Tick(); len(movements) != 1 {
		t.Fatalf("expected other movement %#v", movements)
	}

	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: connectionForTest(), X: 1, Y: 0},
	})
	if err != nil {
		t.Fatalf("handle occupied walk: %v", err)
	}
	units := room.Units()
	if len(units) != 2 || units[0].BodyRotation != worldunit.RotationEast {
		t.Fatalf("expected player facing target %#v", units)
	}
	if len(*sent) != 1 || (*sent)[0].Header != 1640 {
		t.Fatalf("expected status packet, got %#v", *sent)
	}
}

// TestHandleRejectsMissingRoomPresence verifies room presence validation.
func TestHandleRejectsMissingRoomPresence(t *testing.T) {
	handler, _ := handlerForTest(t)
	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: connectionForTest(), X: 1, Y: 0},
	})
	if !errors.Is(err, ErrPlayerNotInRoom) {
		t.Fatalf("expected player not in room, got %v", err)
	}
}

// TestHandleRejectsInvalidTarget verifies target validation.
func TestHandleRejectsInvalidTarget(t *testing.T) {
	handler, player := handlerForTest(t)
	if err := player.EnterRoom(9); err != nil {
		t.Fatalf("enter room: %v", err)
	}

	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: connectionForTest(), X: -1, Y: 0},
	})
	if !errors.Is(err, ErrInvalidTarget) {
		t.Fatalf("expected invalid target, got %v", err)
	}
}

// TestCommandNameAndSoftMoveErrors verifies command identity and soft failures.
func TestCommandNameAndSoftMoveErrors(t *testing.T) {
	handler, player := handlerForTest(t)
	if (Command{}).CommandName() != Name {
		t.Fatalf("unexpected command name %s", (Command{}).CommandName())
	}
	if err := player.EnterRoom(9); err != nil {
		t.Fatalf("enter room: %v", err)
	}
	room, _ := handler.Runtime.Find(9)

	err := handler.handleMoveError(context.Background(), room, 7, grid.MustPoint(1, 0), worldpath.ErrNoPath)
	if err != nil {
		t.Fatalf("handle soft move: %v", err)
	}
	units := room.Units()
	if len(units) != 1 || units[0].BodyRotation != worldunit.RotationEast {
		t.Fatalf("expected facing unit %#v", units)
	}

	err = handler.handleMoveError(context.Background(), room, 99, grid.MustPoint(1, 0), worldpath.ErrNoPath)
	if err != nil {
		t.Fatalf("handle missing unit soft move: %v", err)
	}
	hardErr := errors.New("hard failure")
	err = handler.handleMoveError(context.Background(), room, 7, grid.MustPoint(1, 0), hardErr)
	if !errors.Is(err, hardErr) {
		t.Fatalf("expected hard error, got %v", err)
	}

	// ErrInvalidStart can surface when the surface under a unit changed height after it last moved
	// (e.g. furniture it stood on was moved away); it must never disconnect the client.
	err = handler.handleMoveError(context.Background(), room, 7, grid.MustPoint(1, 0), worldpath.ErrInvalidStart)
	if err != nil {
		t.Fatalf("expected invalid start to be treated as a soft move miss, got %v", err)
	}
}

// handlerForTest creates a walk command handler.
func handlerForTest(t *testing.T) (Handler, *playerlive.Player) {
	t.Helper()

	players := playerlive.NewRegistry()
	peer, err := playerlive.NewSessionPeer(netconn.ID("conn"), netconn.Kind("websocket"), time.Now())
	if err != nil {
		t.Fatalf("create peer: %v", err)
	}
	player, err := playerlive.NewPlayer(playerlive.Snapshot{ID: 7, Username: "demo"}, peer)
	if err != nil {
		t.Fatalf("create player: %v", err)
	}
	if err := players.Add(player); err != nil {
		t.Fatalf("add player: %v", err)
	}

	bindings := binding.NewRegistry()
	if err := bindings.Add(binding.Binding{PlayerID: 7, ConnectionID: netconn.ID("conn"), ConnectionKind: netconn.Kind("websocket")}); err != nil {
		t.Fatalf("add binding: %v", err)
	}

	runtime := roomlive.NewRegistry(nil)
	room, err := runtime.Activate(roomlive.Snapshot{ID: 9, MaxUsers: 10})
	if err != nil {
		t.Fatalf("activate room: %v", err)
	}
	roomGrid, err := grid.Parse("00", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatalf("parse grid: %v", err)
	}
	err = room.LoadWorld(roomlive.WorldConfig{
		Grid: roomGrid,
		Door: worldpath.Position{Point: grid.MustPoint(0, 0)},
		Body: worldunit.RotationSouth,
		Head: worldunit.RotationSouth,
	})
	if err != nil {
		t.Fatalf("load world: %v", err)
	}
	if _, err := runtime.Join(context.Background(), 9, occupantForTest(7)); err != nil {
		t.Fatalf("join runtime: %v", err)
	}

	return Handler{Players: players, Bindings: bindings, Runtime: runtime}, player
}

// connectionForTest creates a connection context.
func connectionForTest() netconn.Context {
	return netconn.Context{ConnectionID: netconn.ID("conn"), ConnectionKind: netconn.Kind("websocket")}
}

// occupantForTest creates a room occupant.
func occupantForTest(playerID int64) roomlive.Occupant {
	return roomlive.Occupant{PlayerID: playerID, Username: "demo", ConnectionID: netconn.ID("conn"), ConnectionKind: netconn.Kind("websocket")}
}

// registeredConnectionForWalkTest creates a registered test connection.
func registeredConnectionForWalkTest(t *testing.T, connections *netconn.Registry, id netconn.ID) *[]codec.Packet {
	t.Helper()

	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error {
		return nil
	}, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	sent := make([]codec.Packet, 0, 1)
	session, err := netconn.NewSession(netconn.SessionConfig{
		ID:       id,
		Kind:     "websocket",
		Outbound: outbound,
		Sender: func(_ context.Context, packet codec.Packet) error {
			sent = append(sent, packet)
			return nil
		},
		Disposer: func(context.Context, netconn.Reason) error {
			return nil
		},
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if err := connections.Register(session); err != nil {
		t.Fatalf("register session: %v", err)
	}

	return &sent
}
