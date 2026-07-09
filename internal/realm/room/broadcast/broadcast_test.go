package broadcast

import (
	"context"
	"errors"
	"testing"

	"github.com/niflaot/pixels/internal/realm/room/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outheightmapupdate "github.com/niflaot/pixels/networking/outbound/room/heightmapupdate"
)

// TestRoomPacketSendsToOccupants verifies room broadcast delivery.
func TestRoomPacketSendsToOccupants(t *testing.T) {
	connections := netconn.NewRegistry()
	sent := registerConnectionForTest(t, connections, "conn")
	room, err := live.NewRoom(live.Snapshot{ID: 9, MaxUsers: 5})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	if _, err := room.Join(live.Occupant{PlayerID: 7, Username: "demo", ConnectionID: "conn", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("join room: %v", err)
	}

	err = RoomPacket(context.Background(), connections, room, codec.Packet{Header: 9}, 0)
	if err != nil {
		t.Fatalf("broadcast packet: %v", err)
	}
	if len(*sent) != 1 || (*sent)[0].Header != 9 {
		t.Fatalf("unexpected sent packets %#v", *sent)
	}

	err = RoomPacket(context.Background(), connections, room, codec.Packet{Header: 10}, 7)
	if err != nil {
		t.Fatalf("broadcast excluded packet: %v", err)
	}
	if len(*sent) != 1 {
		t.Fatalf("expected excluded packet to be skipped %#v", *sent)
	}
}

// TestRoomRemoveEncodesPacket verifies remove broadcasting.
func TestRoomRemoveEncodesPacket(t *testing.T) {
	connections := netconn.NewRegistry()
	sent := registerConnectionForTest(t, connections, "conn")
	room, err := live.NewRoom(live.Snapshot{ID: 9, MaxUsers: 5})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	if _, err := room.Join(live.Occupant{PlayerID: 7, Username: "demo", ConnectionID: "conn", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("join room: %v", err)
	}

	err = RoomRemove(context.Background(), connections, room, 7, 0)
	if err != nil {
		t.Fatalf("remove packet: %v", err)
	}
	if len(*sent) != 1 || (*sent)[0].Header != 2661 {
		t.Fatalf("unexpected sent packets %#v", *sent)
	}
}

// TestMovementPublisherSendsStatus verifies movement publisher wiring.
func TestMovementPublisherSendsStatus(t *testing.T) {
	connections := netconn.NewRegistry()
	sent := registerConnectionForTest(t, connections, "conn")
	room, err := live.NewRoom(live.Snapshot{ID: 9, MaxUsers: 5})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	if _, err := room.Join(live.Occupant{PlayerID: 7, Username: "demo", ConnectionID: "conn", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("join room: %v", err)
	}

	publisher := NewMovementPublisher(connections)
	err = publisher(context.Background(), room, []live.Movement{{Unit: live.UnitSnapshot{UnitID: 1}}})
	if err != nil {
		t.Fatalf("publish movement: %v", err)
	}
	if len(*sent) != 1 || (*sent)[0].Header != 1640 {
		t.Fatalf("unexpected sent packets %#v", *sent)
	}
}

// TestMovementPublisherSkipsMissingState verifies movement no-op guards.
func TestMovementPublisherSkipsMissingState(t *testing.T) {
	publisher := NewMovementPublisher(nil)
	if err := publisher(context.Background(), nil, nil); err != nil {
		t.Fatalf("publish empty movement: %v", err)
	}
}

// TestRoomPacketHandlesMissingConnection verifies stale occupant connections.
func TestRoomPacketHandlesMissingConnection(t *testing.T) {
	room, err := live.NewRoom(live.Snapshot{ID: 9, MaxUsers: 5})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	if _, err := room.Join(live.Occupant{PlayerID: 7, Username: "demo", ConnectionID: "missing", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("join room: %v", err)
	}

	err = RoomPacket(context.Background(), netconn.NewRegistry(), room, codec.Packet{Header: 9}, 0)
	if err != nil {
		t.Fatalf("broadcast missing connection: %v", err)
	}
}

// TestRoomPacketSwallowsSendError verifies a failing recipient never fails the caller.
func TestRoomPacketSwallowsSendError(t *testing.T) {
	sendErr := errors.New("send failed")
	connections := netconn.NewRegistry()
	registerFailingConnectionForTest(t, connections, "bad", sendErr)
	room, err := live.NewRoom(live.Snapshot{ID: 9, MaxUsers: 5})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	if _, err := room.Join(live.Occupant{PlayerID: 7, Username: "demo", ConnectionID: "bad", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("join room: %v", err)
	}

	err = RoomPacket(context.Background(), connections, room, codec.Packet{Header: 9}, 0)
	if err != nil {
		t.Fatalf("expected best-effort delivery to swallow send error, got %v", err)
	}
}

// TestRoomPacketContinuesAfterSendError verifies best-effort delivery reaches other occupants.
func TestRoomPacketContinuesAfterSendError(t *testing.T) {
	sendErr := errors.New("send failed")
	connections := netconn.NewRegistry()
	registerFailingConnectionForTest(t, connections, "bad", sendErr)
	sent := registerConnectionForTest(t, connections, "good")
	room, err := live.NewRoom(live.Snapshot{ID: 9, MaxUsers: 5})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	if _, err := room.Join(live.Occupant{PlayerID: 7, Username: "bad", ConnectionID: "bad", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("join bad room: %v", err)
	}
	if _, err := room.Join(live.Occupant{PlayerID: 8, Username: "good", ConnectionID: "good", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("join good room: %v", err)
	}

	err = RoomPacket(context.Background(), connections, room, codec.Packet{Header: 9}, 0)
	if err != nil {
		t.Fatalf("expected best-effort delivery to swallow send error, got %v", err)
	}
	if len(*sent) != 1 || (*sent)[0].Header != 9 {
		t.Fatalf("expected good connection delivery, got %#v", *sent)
	}
}

// TestRoomHeightMapUpdateSendsRequestedTiles verifies a furniture change's affected points encode
// into a ROOM_HEIGHT_MAP_UPDATE packet sent to every occupant.
func TestRoomHeightMapUpdateSendsRequestedTiles(t *testing.T) {
	connections := netconn.NewRegistry()
	sent := registerConnectionForTest(t, connections, "conn")
	room := worldRoomForBroadcastTest(t, "00\r00")
	if _, err := room.Join(live.Occupant{PlayerID: 7, Username: "demo", ConnectionID: "conn", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("join room: %v", err)
	}

	err := RoomHeightMapUpdate(context.Background(), connections, room, []grid.Point{{X: 0, Y: 0}}, 0)
	if err != nil {
		t.Fatalf("broadcast height map update: %v", err)
	}
	if len(*sent) != 1 || (*sent)[0].Header != outheightmapupdate.Header {
		t.Fatalf("unexpected sent packets %#v", *sent)
	}
}

// TestRoomHeightMapUpdateSkipsEmptyPoints verifies the no-op guard for an empty point set.
func TestRoomHeightMapUpdateSkipsEmptyPoints(t *testing.T) {
	room := worldRoomForBroadcastTest(t, "00\r00")

	if err := RoomHeightMapUpdate(context.Background(), netconn.NewRegistry(), room, nil, 0); err != nil {
		t.Fatalf("broadcast empty height map update: %v", err)
	}
}

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
