package teleport

import (
	"context"
	"testing"
	"time"

	teleportpair "github.com/niflaot/pixels/internal/realm/furniture/interactions/teleport/pair"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outstatus "github.com/niflaot/pixels/networking/outbound/room/entities/status"
	outforward "github.com/niflaot/pixels/networking/outbound/room/forward"
	outupdate "github.com/niflaot/pixels/networking/outbound/room/furniture/update"
	"github.com/niflaot/pixels/pkg/bus"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// crossRoomServiceForTest creates two loaded rooms and a connected session.
func crossRoomServiceForTest(t *testing.T) (*Service, *roomlive.Room, *roomlive.Room, *[]codec.Packet, time.Time) {
	t.Helper()
	now := time.Date(2026, 7, 11, 13, 0, 0, 0, time.UTC)
	runtime := roomlive.NewRegistry(nil)
	definition := worldfurniture.Definition{SpriteID: 202, InteractionType: "teleport", Width: 1, Length: 1}
	source := loadedTeleportRoomForTest(t, runtime, 9, worldfurniture.Item{ID: 1, OwnerPlayerID: 7, Definition: definition, Point: grid.MustPoint(1, 1), Rotation: worldunit.RotationSouth})
	target := loadedTeleportRoomForTest(t, runtime, 10, worldfurniture.Item{ID: 2, OwnerPlayerID: 7, Definition: definition, Point: grid.MustPoint(3, 1), Rotation: worldunit.RotationSouth})
	connections := netconn.NewRegistry()
	sent := make([]codec.Packet, 0, 1)
	if err := connections.Register(connectedSessionForTeleportTest(t, &sent)); err != nil {
		t.Fatalf("register connection: %v", err)
	}
	if _, err := runtime.Join(context.Background(), 9, occupantForTeleportTest(7, "one")); err != nil {
		t.Fatalf("join source room: %v", err)
	}
	sourceRoomID, targetRoomID, sourceX, targetX, y, z := int64(9), int64(10), 1, 3, 1, 0.0
	furniture := &teleportFurnitureForTest{
		items: map[int64]furnituremodel.Item{
			1: {Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 1}}, DefinitionID: 8, OwnerPlayerID: 7, RoomID: &sourceRoomID, X: &sourceX, Y: &y, Z: &z},
			2: {Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 2}}, DefinitionID: 8, OwnerPlayerID: 7, RoomID: &targetRoomID, X: &targetX, Y: &y, Z: &z},
		},
		definition: furnituremodel.Definition{SpriteID: 202, Width: 1, Length: 1, InteractionType: "teleport"},
	}
	service := NewService(Config{}, teleportpair.NewService(teleportPairStoreForTest{}, furniture), runtime, connections, nil, bus.New())
	service.now = func() time.Time { return now }

	return service, source, target, &sent, now
}

// loadedTeleportRoomForTest activates one room with one teleport item.
func loadedTeleportRoomForTest(t *testing.T, runtime *roomlive.Registry, roomID int64, item worldfurniture.Item) *roomlive.Room {
	t.Helper()
	roomGrid, err := grid.Parse("00000\r00000\r00000", grid.WithDoor(1, 2))
	if err != nil {
		t.Fatalf("parse room grid: %v", err)
	}
	active, err := runtime.Activate(roomlive.Snapshot{ID: roomID, OwnerPlayerID: 7, MaxUsers: 25})
	if err != nil {
		t.Fatalf("activate room: %v", err)
	}
	if err := active.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Furniture: []worldfurniture.Item{item}, Door: worldpath.Position{Point: grid.MustPoint(1, 2)}, Body: worldunit.RotationNorth, Head: worldunit.RotationNorth}); err != nil {
		t.Fatalf("load room world: %v", err)
	}

	return active
}

// connectedSessionForTeleportTest creates a connected outbound-capable session.
func connectedSessionForTeleportTest(t *testing.T, sent *[]codec.Packet) *netconn.Session {
	t.Helper()
	outbound := netconn.NewHandlerRegistry()
	for _, header := range []uint16{outforward.Header, outupdate.Header, outstatus.Header} {
		if err := outbound.Register(header, func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated()); err != nil {
			t.Fatalf("register outbound handler %d: %v", header, err)
		}
	}
	session, err := netconn.NewSession(netconn.SessionConfig{ID: "one", Kind: "websocket", Outbound: outbound, Sender: func(_ context.Context, packet codec.Packet) error { *sent = append(*sent, packet); return nil }, Disposer: func(context.Context, netconn.Reason) error { return nil }})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	for _, event := range []netconn.Event{netconn.EventPacketReceived, netconn.EventAuthenticationStarted, netconn.EventAuthenticationAccepted, netconn.EventSessionReady} {
		if err := session.Transition(event); err != nil {
			t.Fatalf("transition session: %v", err)
		}
	}

	return session
}

// occupantForTeleportTest creates a stable room occupant.
func occupantForTeleportTest(playerID int64, connectionID netconn.ID) roomlive.Occupant {
	return roomlive.Occupant{PlayerID: playerID, Username: "demo", ConnectionID: connectionID, ConnectionKind: "websocket"}
}
