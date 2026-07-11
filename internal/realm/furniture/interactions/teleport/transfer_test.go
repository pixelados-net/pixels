package teleport

import (
	"context"
	"errors"
	"testing"
	"time"

	teleportfailed "github.com/niflaot/pixels/internal/realm/furniture/events/teleportfailed"
	furniturewalkedon "github.com/niflaot/pixels/internal/realm/furniture/events/walkedon"
	teleportpair "github.com/niflaot/pixels/internal/realm/furniture/interactions/teleport/pair"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	playerdisconnected "github.com/niflaot/pixels/internal/realm/player/events/disconnected"
	roomentered "github.com/niflaot/pixels/internal/realm/room/access/events/entered"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outforward "github.com/niflaot/pixels/networking/outbound/room/forward"
	"github.com/niflaot/pixels/pkg/bus"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
	"go.uber.org/fx/fxtest"
)

// TestCrossRoomTransitionForwardsAndConsumesSpawn verifies the Nitro navigation handoff.
func TestCrossRoomTransitionForwardsAndConsumesSpawn(t *testing.T) {
	service, source, target, sent, now := crossRoomServiceForTest(t)
	if err := service.Start(context.Background(), StartRequest{PlayerID: 7, Room: source, ItemID: 1}); err != nil {
		t.Fatalf("start cross-room teleport: %v", err)
	}
	if err := service.Cycle(context.Background(), source, now); err != nil {
		t.Fatalf("open source: %v", err)
	}
	if err := service.Cycle(context.Background(), source, now.Add(phaseDelay)); err != nil {
		t.Fatalf("forward target: %v", err)
	}
	if len(*sent) != 1 || (*sent)[0].Header != outforward.Header {
		t.Fatalf("expected room forward packet, got %#v", *sent)
	}
	_, _, _ = service.runtime.Leave(context.Background(), 7)
	if _, err := service.runtime.Join(context.Background(), 10, occupantForTeleportTest()); err != nil {
		t.Fatalf("join target room: %v", err)
	}
	if err := service.entered(context.Background(), roomentered.Payload{PlayerID: 7, RoomID: 10}); err != nil {
		t.Fatalf("consume target spawn: %v", err)
	}
	unit, found := target.Unit(7)
	if !found || unit.Position.Point != grid.MustPoint(3, 1) {
		t.Fatalf("unexpected destination unit %#v found=%v", unit, found)
	}
	if _, found := service.consumePending(7, 10); found {
		t.Fatal("expected one-time destination consumption")
	}
}

// TestPendingExpiryClearAndFailure verifies bounded transfer cleanup paths.
func TestPendingExpiryClearAndFailure(t *testing.T) {
	service, active, _, _, now := crossRoomServiceForTest(t)
	service.pending = map[int64]pendingDestination{
		7: {roomID: 10, expiresAt: now.Add(-time.Second)},
		8: {roomID: 11, expiresAt: now.Add(time.Second)},
	}
	if _, found := service.consumePending(7, 10); found {
		t.Fatal("expected expired destination rejection")
	}
	if _, found := service.consumePending(8, 10); found {
		t.Fatal("expected destination room mismatch")
	}
	service.clearPending(8)
	if service.pending != nil {
		t.Fatalf("expected empty pending map, got %#v", service.pending)
	}
	failed := false
	_, err := service.events.(*bus.Bus).Subscribe(teleportfailed.Name, bus.PriorityNormal, func(_ context.Context, event bus.Event) error {
		_, failed = event.Payload.(teleportfailed.Payload)
		return nil
	})
	if err != nil {
		t.Fatalf("subscribe failure event: %v", err)
	}
	if err := service.fail(context.Background(), active.ID(), Transit{PlayerID: 7, Source: worldfurniture.Item{ID: 1}}, "test"); err != nil || !failed {
		t.Fatalf("expected failure event failed=%v err=%v", failed, err)
	}
}

// TestRegisterCleansPendingOnDisconnect verifies lifecycle event integration.
func TestRegisterCleansPendingOnDisconnect(t *testing.T) {
	service, active, _ := serviceForTest(t, "teleport_tile")
	local := bus.New()
	service.events = local
	lifecycle := fxtest.NewLifecycle(t)
	if err := Register(lifecycle, local, service.runtime, service); err != nil {
		t.Fatalf("register teleport lifecycle: %v", err)
	}
	service.pending = map[int64]pendingDestination{7: {roomID: active.ID(), expiresAt: time.Now().Add(time.Minute)}}
	if err := local.Publish(context.Background(), bus.Event{Name: playerdisconnected.Name, Payload: playerdisconnected.Payload{PlayerID: 7}}); err != nil {
		t.Fatalf("publish disconnect: %v", err)
	}
	if service.pending != nil {
		t.Fatalf("expected disconnect cleanup, got %#v", service.pending)
	}
	if err := local.Publish(context.Background(), bus.Event{Name: furniturewalkedon.Name, Payload: furniturewalkedon.Payload{
		PlayerID: 7, ItemID: 1, RoomID: active.ID(),
	}}); err != nil {
		t.Fatalf("publish walked-on event: %v", err)
	}
	if _, found := service.rooms.Load(active.ID()); !found {
		t.Fatal("expected walk-on subscription to start tile transition")
	}
	lifecycle.RequireStop()
}

// TestStartValidationAndRemoveTransit verifies cheap invalid and reservation cleanup paths.
func TestStartValidationAndRemoveTransit(t *testing.T) {
	service, active, _ := serviceForTest(t, "teleport")
	if err := service.Start(context.Background(), StartRequest{}); err != ErrInvalidUse {
		t.Fatalf("expected invalid use, got %v", err)
	}
	if err := service.Start(context.Background(), StartRequest{PlayerID: 7, Room: active, ItemID: 99}); err != ErrNotTeleport {
		t.Fatalf("expected non-teleport rejection, got %v", err)
	}
	state := service.roomState(active.ID())
	state.transits[7] = Transit{PlayerID: 7}
	service.removeTransit(active.ID(), 7)
	if _, found := service.rooms.Load(active.ID()); found {
		t.Fatal("expected empty room shard removal")
	}
}

// TestTransferNoOpAndMissingConnectionBranches verifies inexpensive defensive paths.
func TestTransferNoOpAndMissingConnectionBranches(t *testing.T) {
	service, active, _ := serviceForTest(t, "teleport")
	if err := service.entered(context.Background(), roomentered.Payload{PlayerID: 7, RoomID: active.ID()}); err != nil {
		t.Fatalf("unexpected entry without destination error: %v", err)
	}
	if _, found := service.playerConnection(active, 99); found {
		t.Fatal("expected missing room connection")
	}
	if err := service.forward(context.Background(), active, Transit{
		PlayerID: 99, Source: worldfurniture.Item{ID: 1}, TargetRoomID: 10,
	}); err != nil {
		t.Fatalf("missing connection should publish a soft failure: %v", err)
	}
	service.events = nil
	if err := service.publishStarted(context.Background(), active.ID(), Transit{}); err != nil {
		t.Fatalf("nil publisher should be a no-op: %v", err)
	}
	expected := errors.New("first")
	if firstError(expected, errors.New("second")) != expected || firstError(nil, expected) != expected {
		t.Fatal("unexpected first-error selection")
	}
}

// crossRoomServiceForTest creates two loaded rooms and a connected session.
func crossRoomServiceForTest(t *testing.T) (*Service, *roomlive.Room, *roomlive.Room, *[]codec.Packet, time.Time) {
	t.Helper()
	now := time.Date(2026, 7, 11, 13, 0, 0, 0, time.UTC)
	runtime := roomlive.NewRegistry(nil)
	definition := worldfurniture.Definition{SpriteID: 202, InteractionType: "teleport", Width: 1, Length: 1}
	source := loadedTeleportRoomForTest(t, runtime, 9, worldfurniture.Item{
		ID: 1, OwnerPlayerID: 7, Definition: definition, Point: grid.MustPoint(1, 1), Rotation: worldunit.RotationSouth,
	})
	target := loadedTeleportRoomForTest(t, runtime, 10, worldfurniture.Item{
		ID: 2, OwnerPlayerID: 7, Definition: definition, Point: grid.MustPoint(3, 1), Rotation: worldunit.RotationSouth,
	})
	connections := netconn.NewRegistry()
	sent := make([]codec.Packet, 0, 1)
	session := connectedSessionForTeleportTest(t, &sent)
	if err := connections.Register(session); err != nil {
		t.Fatalf("register connection: %v", err)
	}
	if _, err := runtime.Join(context.Background(), 9, occupantForTeleportTest()); err != nil {
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
	pairs := teleportpair.NewService(teleportPairStoreForTest{}, furniture)
	service := NewService(Config{}, pairs, runtime, connections, nil, bus.New())
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
	if err := active.LoadWorld(roomlive.WorldConfig{
		Grid: roomGrid, Furniture: []worldfurniture.Item{item}, Door: worldpath.Position{Point: grid.MustPoint(1, 2)},
		Body: worldunit.RotationNorth, Head: worldunit.RotationNorth,
	}); err != nil {
		t.Fatalf("load room world: %v", err)
	}

	return active
}

// connectedSessionForTeleportTest creates a connected outbound-capable session.
func connectedSessionForTeleportTest(t *testing.T, sent *[]codec.Packet) *netconn.Session {
	t.Helper()
	outbound := netconn.NewHandlerRegistry()
	if err := outbound.Register(outforward.Header, func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated()); err != nil {
		t.Fatalf("register outbound handler: %v", err)
	}
	session, err := netconn.NewSession(netconn.SessionConfig{
		ID: "one", Kind: "websocket", Outbound: outbound,
		Sender:   func(_ context.Context, packet codec.Packet) error { *sent = append(*sent, packet); return nil },
		Disposer: func(context.Context, netconn.Reason) error { return nil },
	})
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
func occupantForTeleportTest() roomlive.Occupant {
	return roomlive.Occupant{PlayerID: 7, Username: "demo", ConnectionID: "one", ConnectionKind: "websocket"}
}
