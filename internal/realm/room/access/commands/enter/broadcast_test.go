package enter

import (
	"context"
	"testing"

	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	"github.com/niflaot/pixels/internal/realm/room/world/layout"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// TestBroadcastJoinedSkipsMissingConnections verifies optional broadcasts.
func TestBroadcastJoinedSkipsMissingConnections(t *testing.T) {
	handler := Handler{}
	err := handler.broadcastJoined(context.Background(), activeRoomForBroadcastTest(t), 7)
	if err != nil {
		t.Fatalf("broadcast joined: %v", err)
	}
}

// TestSendRoomStateWithEmptyRoom verifies empty state sends nothing.
func TestSendRoomStateWithEmptyRoom(t *testing.T) {
	connection, sent := sessionConnectionForTest(t)
	room, err := roomlive.NewRoom(roomlive.Snapshot{ID: 9, MaxUsers: 5})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}

	err = (Handler{}).sendRoomState(context.Background(), connection, room, 0)
	if err != nil {
		t.Fatalf("send room state: %v", err)
	}
	if len(*sent) != 0 {
		t.Fatalf("expected no packets, got %#v", *sent)
	}
}

// TestLoadWorldRejectsInvalidLayout verifies layout validation during loading.
func TestLoadWorldRejectsInvalidLayout(t *testing.T) {
	room, err := roomlive.NewRoom(roomlive.Snapshot{ID: 9, MaxUsers: 5})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	err = (Handler{}).loadWorld(context.Background(), room, roommodel.Room{}, layout.Layout{Heightmap: "x", DoorX: 0, DoorY: 0})
	if err == nil {
		t.Fatal("expected invalid layout world")
	}
}

// TestPlayerFilter verifies optional room state filtering.
func TestPlayerFilter(t *testing.T) {
	if playerFilter(0) != nil {
		t.Fatal("expected no filter")
	}
	filter := playerFilter(7)
	if len(filter) != 1 || filter[0] != 7 {
		t.Fatalf("unexpected filter %#v", filter)
	}
}

// TestBroadcastJoinedSendsToOtherOccupants verifies joined broadcasts.
func TestBroadcastJoinedSendsToOtherOccupants(t *testing.T) {
	connections := netconn.NewRegistry()
	sent := registeredConnectionForBroadcastTest(t, connections, "other")
	room := activeRoomForBroadcastTest(t)
	if _, err := room.Join(roomlive.Occupant{PlayerID: 8, Username: "other", ConnectionID: "other", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("join other: %v", err)
	}

	err := (Handler{Connections: connections}).broadcastJoined(context.Background(), room, 7)
	if err != nil {
		t.Fatalf("broadcast joined: %v", err)
	}
	if len(*sent) != 2 {
		t.Fatalf("expected unit and status packets, got %#v", *sent)
	}
}

// TestJoinBroadcastsPreviousRoomRemoval verifies room switches remove old units.
func TestJoinBroadcastsPreviousRoomRemoval(t *testing.T) {
	player := playerForTest(t)
	connections := netconn.NewRegistry()
	sent := registeredConnectionForBroadcastTest(t, connections, "other")
	runtime := roomlive.NewRegistry(nil)
	previous, err := runtime.Activate(roomlive.Snapshot{ID: 3, MaxUsers: 5})
	if err != nil {
		t.Fatalf("activate previous room: %v", err)
	}
	if err := (Handler{}).loadWorld(context.Background(), previous, roommodel.Room{}, layoutForTest()); err != nil {
		t.Fatalf("load previous world: %v", err)
	}
	if _, err := runtime.Join(context.Background(), 3, occupantForTest(7)); err != nil {
		t.Fatalf("join previous self: %v", err)
	}
	if _, err := runtime.Join(context.Background(), 3, roomlive.Occupant{PlayerID: 8, Username: "other", ConnectionID: "other", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("join previous other: %v", err)
	}
	if err := player.EnterRoom(3); err != nil {
		t.Fatalf("enter previous room: %v", err)
	}

	_, err = (Handler{Runtime: runtime, Connections: connections}).join(context.Background(), player, connectionForTest(), roomForTest(), layoutForTest())
	if err != nil {
		t.Fatalf("join target room: %v", err)
	}
	if len(*sent) != 1 || (*sent)[0].Header != 2661 {
		t.Fatalf("expected remove packet, got %#v", *sent)
	}
}

// activeRoomForBroadcastTest creates a loaded room with one occupant.
func activeRoomForBroadcastTest(t *testing.T) *roomlive.Room {
	t.Helper()

	room, err := roomlive.NewRoom(roomlive.Snapshot{ID: 9, MaxUsers: 5})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	roomGrid, err := grid.Parse("0", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatalf("parse grid: %v", err)
	}
	err = room.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}})
	if err != nil {
		t.Fatalf("load world: %v", err)
	}
	if _, err := room.Join(roomlive.Occupant{PlayerID: 7, Username: "demo", ConnectionID: "self", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("join room: %v", err)
	}

	return room
}

// registeredConnectionForBroadcastTest creates a registered test connection.
func registeredConnectionForBroadcastTest(t *testing.T, connections *netconn.Registry, id netconn.ID) *[]codec.Packet {
	t.Helper()

	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error {
		return nil
	}, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	sent := make([]codec.Packet, 0, 2)
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
