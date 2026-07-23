package leave

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomleft "github.com/niflaot/pixels/internal/realm/room/access/events/left"
	roomsession "github.com/niflaot/pixels/internal/realm/room/runtime/commands/session"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
)

// TestCommandName verifies the stable command name.
func TestCommandName(t *testing.T) {
	if (Command{}).CommandName() != Name {
		t.Fatalf("unexpected command name %s", (Command{}).CommandName())
	}
}

// TestHandleLeavesRoom verifies room leave orchestration.
func TestHandleLeavesRoom(t *testing.T) {
	player := playerForTest(t)
	connections := netconn.NewRegistry()
	sent := registerConnectionForTest(t, connections, "other")
	runtime := roomlive.NewRegistry(nil)
	active := activeRoomForTest(t, runtime)
	if _, err := active.Join(roomlive.Occupant{PlayerID: 8, Username: "other", ConnectionID: "other", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("join other: %v", err)
	}
	publisher := &publisherForTest{}

	err := (Handler{
		Players: playerRegistryForTest(t, player), Bindings: bindingRegistryForTest(t),
		Runtime: runtime, Connections: connections, Events: publisher,
	}).Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: connectionForTest()},
	})
	if err != nil {
		t.Fatalf("leave room: %v", err)
	}
	if _, found := player.CurrentRoom(); found {
		t.Fatalf("expected player room cleared")
	}
	if active.Occupancy().Count != 1 {
		t.Fatalf("unexpected occupancy %#v", active.Occupancy())
	}
	if len(*sent) != 1 || (*sent)[0].Header != 2661 {
		t.Fatalf("expected remove packet, got %#v", *sent)
	}
	if len(publisher.events) != 1 || publisher.events[0].Name != roomleft.Name {
		t.Fatalf("unexpected events %#v", publisher.events)
	}
}

// TestHandleLeavesDirectPlayer verifies player id based exits.
func TestHandleLeavesDirectPlayer(t *testing.T) {
	player := playerForTest(t)
	runtime := roomlive.NewRegistry(nil)
	activeRoomForTest(t, runtime)

	err := (Handler{
		Players: playerRegistryForTest(t, player),
		Runtime: runtime,
	}).Handle(context.Background(), command.Envelope[Command]{
		Command: Command{PlayerID: 7},
	})
	if err != nil {
		t.Fatalf("leave direct player: %v", err)
	}
	if _, found := player.CurrentRoom(); found {
		t.Fatalf("expected player room cleared")
	}
}

// TestHandleIgnoresMissingRoom verifies idempotent exits.
func TestHandleIgnoresMissingRoom(t *testing.T) {
	err := (Handler{Runtime: roomlive.NewRegistry(nil)}).Handle(context.Background(), command.Envelope[Command]{
		Command: Command{PlayerID: 7},
	})
	if err != nil {
		t.Fatalf("leave missing room: %v", err)
	}
}

// TestHandleReturnsSessionError verifies missing session bindings.
func TestHandleReturnsSessionError(t *testing.T) {
	err := (Handler{Runtime: roomlive.NewRegistry(nil)}).Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: connectionForTest()},
	})
	if !errors.Is(err, roomsession.ErrBindingNotFound) {
		t.Fatalf("expected binding error, got %v", err)
	}
}

// playerForTest creates a live player.
func playerForTest(t *testing.T) *playerlive.Player {
	t.Helper()

	peer, err := playerlive.NewSessionPeer("conn", "websocket", time.Now())
	if err != nil {
		t.Fatalf("create peer: %v", err)
	}
	player, err := playerlive.NewPlayer(playerlive.Snapshot{ID: 7, Username: "demo"}, peer)
	if err != nil {
		t.Fatalf("create player: %v", err)
	}
	if err := player.EnterRoom(9); err != nil {
		t.Fatalf("enter room: %v", err)
	}

	return player
}

// playerRegistryForTest creates a registry with one player.
func playerRegistryForTest(t *testing.T, player *playerlive.Player) *playerlive.Registry {
	t.Helper()

	registry := playerlive.NewRegistry()
	if err := registry.Add(player); err != nil {
		t.Fatalf("add player: %v", err)
	}

	return registry
}

// bindingRegistryForTest creates a connection binding registry.
func bindingRegistryForTest(t *testing.T) *binding.Registry {
	t.Helper()

	registry := binding.NewRegistry()
	err := registry.Add(binding.Binding{PlayerID: 7, ConnectionID: "conn", ConnectionKind: "websocket"})
	if err != nil {
		t.Fatalf("add binding: %v", err)
	}

	return registry
}

// activeRoomForTest creates an active room with one player.
func activeRoomForTest(t *testing.T, runtime *roomlive.Registry) *roomlive.Room {
	t.Helper()

	active, err := runtime.Activate(roomlive.Snapshot{ID: 9, MaxUsers: 5})
	if err != nil {
		t.Fatalf("activate room: %v", err)
	}
	roomGrid, err := grid.Parse("0", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatalf("parse grid: %v", err)
	}
	err = active.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}})
	if err != nil {
		t.Fatalf("load world: %v", err)
	}
	if _, err := runtime.Join(context.Background(), 9, roomlive.Occupant{PlayerID: 7, Username: "demo", ConnectionID: "conn", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("join room: %v", err)
	}

	return active
}

// connectionForTest creates a handler connection context.
func connectionForTest() netconn.Context {
	return netconn.Context{ConnectionID: "conn", ConnectionKind: "websocket"}
}

// registerConnectionForTest registers a captured test connection.
func registerConnectionForTest(t *testing.T, connections *netconn.Registry, id netconn.ID) *[]codec.Packet {
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

// publisherForTest captures published events.
type publisherForTest struct {
	// events stores published events.
	events []bus.Event
}

// Publish captures an event.
func (publisher *publisherForTest) Publish(_ context.Context, event bus.Event) error {
	publisher.events = append(publisher.events, event)

	return nil
}
