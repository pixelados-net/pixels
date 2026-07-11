package broadcast

import (
	"context"
	"testing"

	"github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// worldRoomForBroadcastTest creates an active room with a loaded world for height map broadcasting.
func worldRoomForBroadcastTest(t *testing.T, heightmap string) *live.Room {
	t.Helper()

	room, err := live.NewRoom(live.Snapshot{ID: 9, MaxUsers: 5})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	roomGrid, err := grid.Parse(heightmap, grid.WithDoor(0, 0))
	if err != nil {
		t.Fatalf("parse grid: %v", err)
	}
	if err := room.LoadWorld(live.WorldConfig{
		Grid: roomGrid,
		Door: worldpath.Position{Point: grid.Point{X: 0, Y: 0}},
		Body: worldunit.RotationSouth,
		Head: worldunit.RotationSouth,
	}); err != nil {
		t.Fatalf("load world: %v", err)
	}

	return room
}

// registerConnectionForTest registers a captured test connection.
func registerConnectionForTest(t *testing.T, connections *netconn.Registry, id netconn.ID) *[]codec.Packet {
	t.Helper()

	sent := make([]codec.Packet, 0, 1)
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error {
		return nil
	}, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
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

// registerFailingConnectionForTest registers a failing test connection.
func registerFailingConnectionForTest(t *testing.T, connections *netconn.Registry, id netconn.ID, sendErr error) {
	t.Helper()

	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error {
		return nil
	}, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	session, err := netconn.NewSession(netconn.SessionConfig{
		ID:       id,
		Kind:     "websocket",
		Outbound: outbound,
		Sender: func(context.Context, codec.Packet) error {
			return sendErr
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
}
