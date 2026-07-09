package pickup

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/command"
	pickedupevent "github.com/niflaot/pixels/internal/realm/furniture/events/pickedup"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomlive "github.com/niflaot/pixels/internal/realm/room/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outinvadd "github.com/niflaot/pixels/networking/outbound/inventory/furniture/add"
	outstatus "github.com/niflaot/pixels/networking/outbound/room/entities/status"
	outremove "github.com/niflaot/pixels/networking/outbound/room/furniture/remove"
	outheightmapupdate "github.com/niflaot/pixels/networking/outbound/room/heightmapupdate"
	outbubble "github.com/niflaot/pixels/networking/outbound/session/bubblealert"
	"github.com/niflaot/pixels/pkg/bus"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
	"go.uber.org/zap"
)

// TestHandlePicksUpItemAndBroadcasts verifies the successful pickup path.
func TestHandlePicksUpItemAndBroadcasts(t *testing.T) {
	handler, connections := handlerForTest(t)
	connection, sent := registeredConnectionForTest(t, connections, "conn")
	handler.Furniture = &fakeManager{pickupResult: pickedItemForTest()}

	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: connection, ItemID: 1},
	})
	if err != nil {
		t.Fatalf("handle pickup: %v", err)
	}

	room, _ := handler.Runtime.Find(9)
	if items := room.FurnitureItems(); len(items) != 0 {
		t.Fatalf("expected no furniture items after pickup, got %#v", items)
	}
	if len(*sent) != 2 || (*sent)[0].Header != outremove.Header || (*sent)[1].Header != outinvadd.Header {
		t.Fatalf("expected floor remove then inventory add packets, got %#v", *sent)
	}
}

// TestHandleBroadcastsRemoveToOtherOccupants verifies both connections in a room observe a pickup.
func TestHandleBroadcastsRemoveToOtherOccupants(t *testing.T) {
	handler, connections := handlerForTest(t)
	actor, actorSent := registeredConnectionForTest(t, connections, "conn")
	_, bystanderSent := registeredConnectionForTest(t, connections, "bystander-conn")
	room, _ := handler.Runtime.Find(9)
	if _, err := room.Join(roomlive.Occupant{
		PlayerID: 8, Username: "bystander", ConnectionID: netconn.ID("bystander-conn"), ConnectionKind: netconn.Kind("websocket"),
	}); err != nil {
		t.Fatalf("join bystander: %v", err)
	}
	handler.Furniture = &fakeManager{pickupResult: pickedItemForTest()}

	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: actor, ItemID: 1},
	})
	if err != nil {
		t.Fatalf("handle pickup: %v", err)
	}

	if len(*actorSent) != 2 || (*actorSent)[0].Header != outremove.Header || (*actorSent)[1].Header != outinvadd.Header {
		t.Fatalf("expected actor to receive floor remove then inventory add, got %#v", *actorSent)
	}
	if len(*bystanderSent) != 1 || (*bystanderSent)[0].Header != outremove.Header {
		t.Fatalf("expected bystander to receive only floor remove, got %#v", *bystanderSent)
	}
}

// TestHandlePickingUpOccupiedChairStandsOccupantUp verifies picking up a chair stands its occupant
// up and syncs the resulting status to every room connection.
func TestHandlePickingUpOccupiedChairStandsOccupantUp(t *testing.T) {
	handler, connections := handlerForTest(t)
	actor, actorSent := registeredConnectionForTest(t, connections, "conn")
	room, _ := handler.Runtime.Find(9)
	settleUnitOnChair(t, room, 7, 1)
	handler.Furniture = &fakeManager{pickupResult: pickedItemForTest()}

	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: actor, ItemID: 1},
	})
	if err != nil {
		t.Fatalf("handle pickup: %v", err)
	}

	units := room.Units()
	if len(units) != 1 || unitHasStatus(units[0].Statuses, worldunit.StatusSit) {
		t.Fatalf("expected occupant to stand up after the chair was picked up, got %#v", units)
	}
	if len(*actorSent) != 3 ||
		(*actorSent)[0].Header != outremove.Header ||
		(*actorSent)[1].Header != outstatus.Header ||
		(*actorSent)[2].Header != outinvadd.Header {
		t.Fatalf("expected floor remove, unit status, then inventory add packets, got %#v", *actorSent)
	}
}

// TestHandlePicksUpItemBroadcastsHeightMapUpdate verifies the picked up item's vacated footprint
// is broadcast as a ROOM_HEIGHT_MAP_UPDATE so every client's cached local height map stays in sync.
func TestHandlePicksUpItemBroadcastsHeightMapUpdate(t *testing.T) {
	handler, connections := handlerForTest(t)
	connection, sent := registeredConnectionForTest(t, connections, "conn")
	handler.Furniture = &fakeManager{
		definition:      furnituremodel.Definition{Width: 1, Length: 1},
		definitionFound: true,
		pickupResult:    placedPickedItemForTest(),
	}

	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: connection, ItemID: 1},
	})
	if err != nil {
		t.Fatalf("handle pickup: %v", err)
	}

	if len(*sent) != 3 ||
		(*sent)[0].Header != outremove.Header ||
		(*sent)[1].Header != outheightmapupdate.Header ||
		(*sent)[2].Header != outinvadd.Header {
		t.Fatalf("expected floor remove, height map update, then inventory add packets, got %#v", *sent)
	}
}

// placedPickedItemForTest returns a picked up item fixture that was on the floor before pickup.
func placedPickedItemForTest() furnituremodel.Item {
	roomID, x, y, z := int64(9), 1, 0, 0.0

	return furnituremodel.Item{
		Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 1}}, DefinitionID: 2, OwnerPlayerID: 7,
		RoomID: &roomID, X: &x, Y: &y, Z: &z,
	}
}

// unitHasStatus reports whether a status key is present in a snapshot's status list.
func unitHasStatus(statuses []worldunit.Status, key string) bool {
	for _, status := range statuses {
		if status.Key == key {
			return true
		}
	}

	return false
}

// TestHandleRejectsMissingRoomPresence verifies the room-presence guard.
func TestHandleRejectsMissingRoomPresence(t *testing.T) {
	handler, _ := handlerForTest(t)
	handler.Players = playersForTest(t, false)

	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: connectionForTest()},
	})
	if err != nil {
		t.Fatalf("expected no error for missing room presence, got %v", err)
	}
}

// TestHandleIgnoresSoftPickupErrors verifies gameplay misses stay silent and send a bubble alert.
func TestHandleIgnoresSoftPickupErrors(t *testing.T) {
	handler, connections := handlerForTest(t)
	connection, sent := registeredConnectionForTest(t, connections, "conn")
	handler.Furniture = &fakeManager{pickupErr: furnitureservice.ErrItemNotFound}

	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: connection, ItemID: 1},
	})
	if err != nil {
		t.Fatalf("expected soft pickup error to stay silent, got %v", err)
	}
	if len(*sent) != 1 || (*sent)[0].Header != outbubble.Header {
		t.Fatalf("expected a bubble alert packet, got %#v", *sent)
	}
}

// TestHandlePropagatesHardErrors verifies unexpected persistence errors surface.
func TestHandlePropagatesHardErrors(t *testing.T) {
	handler, _ := handlerForTest(t)
	expected := errors.New("pickup failed")
	handler.Furniture = &fakeManager{pickupErr: expected}

	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: connectionForTest(), ItemID: 1},
	})
	if !errors.Is(err, expected) {
		t.Fatalf("expected pickup error, got %v", err)
	}
}

// TestHandlePicksUpAndPublishesWithLogger verifies event publication and logging on success.
func TestHandlePicksUpAndPublishesWithLogger(t *testing.T) {
	handler, connections := handlerForTest(t)
	connection, _ := registeredConnectionForTest(t, connections, "conn")
	publisher := &fakePublisher{}
	handler.Events = publisher
	handler.Log = zap.NewNop()
	handler.Furniture = &fakeManager{pickupResult: pickedItemForTest()}

	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: connection, ItemID: 1},
	})
	if err != nil {
		t.Fatalf("handle pickup: %v", err)
	}
	if len(publisher.events) != 1 || publisher.events[0].Name != pickedupevent.Name {
		t.Fatalf("expected pickedup event published, got %#v", publisher.events)
	}
}

// TestHandleLogsRejectionWithLogger verifies rejected pickups log without error when a logger is set.
func TestHandleLogsRejectionWithLogger(t *testing.T) {
	handler, connections := handlerForTest(t)
	connection, _ := registeredConnectionForTest(t, connections, "conn")
	handler.Log = zap.NewNop()
	handler.Furniture = &fakeManager{pickupErr: furnitureservice.ErrItemNotFound}

	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: connection, ItemID: 1},
	})
	if err != nil {
		t.Fatalf("expected soft pickup error to stay silent, got %v", err)
	}
}

// TestBubbleErrorKeyMapsKnownErrors verifies every mapped soft error resolves to its bubble key.
func TestBubbleErrorKeyMapsKnownErrors(t *testing.T) {
	cases := []struct {
		err      error
		wantKey  string
		wantSoft bool
	}{
		{furnitureservice.ErrNotItemOwner, "no_rights", true},
		{furnitureservice.ErrItemNotFound, "item_not_found", true},
		{furnitureservice.ErrItemNotPlaced, "item_not_found", true},
		{furnitureservice.ErrInvalidItemID, "invalid_move", true},
		{furnitureservice.ErrInvalidPlayerID, "invalid_move", true},
		{errors.New("unmapped"), "", false},
	}

	for _, testCase := range cases {
		key, soft := bubbleErrorKey(testCase.err)
		if key != testCase.wantKey || soft != testCase.wantSoft {
			t.Fatalf("bubbleErrorKey(%v) = (%q, %v), want (%q, %v)", testCase.err, key, soft, testCase.wantKey, testCase.wantSoft)
		}
	}
}

// fakePublisher records published events for tests.
type fakePublisher struct {
	events []bus.Event
}

// Publish records a published event for tests.
func (publisher *fakePublisher) Publish(_ context.Context, event bus.Event) error {
	publisher.events = append(publisher.events, event)

	return nil
}

// TestHandleSkipsBroadcastWhenConnectionsNil verifies a nil connection registry is tolerated.
func TestHandleSkipsBroadcastWhenConnectionsNil(t *testing.T) {
	handler, standaloneConnections := handlerForTest(t)
	connection, _ := registeredConnectionForTest(t, standaloneConnections, "conn")
	handler.Connections = nil
	handler.Furniture = &fakeManager{pickupResult: pickedItemForTest()}

	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: connection, ItemID: 1},
	})
	if err != nil {
		t.Fatalf("expected no error with nil connections, got %v", err)
	}
}

// TestCommandName verifies the stable command name.
func TestCommandName(t *testing.T) {
	if (Command{}).CommandName() != Name {
		t.Fatalf("unexpected command name %s", (Command{}).CommandName())
	}
}

// handlerForTest creates a pickup command handler with a loaded room and joined player.
func handlerForTest(t *testing.T) (Handler, *netconn.Registry) {
	t.Helper()

	players := playersForTest(t, true)
	bindings := binding.NewRegistry()
	if err := bindings.Add(binding.Binding{PlayerID: 7, ConnectionID: netconn.ID("conn"), ConnectionKind: netconn.Kind("websocket")}); err != nil {
		t.Fatalf("add binding: %v", err)
	}

	runtime := roomlive.NewRegistry(nil)
	room, err := runtime.Activate(roomlive.Snapshot{ID: 9, MaxUsers: 10})
	if err != nil {
		t.Fatalf("activate room: %v", err)
	}
	roomGrid, err := grid.Parse("0000", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatalf("parse grid: %v", err)
	}
	if err := room.LoadWorld(roomlive.WorldConfig{
		Grid: roomGrid,
		Door: worldpath.Position{Point: grid.MustPoint(0, 0)},
		Body: worldunit.RotationSouth,
		Head: worldunit.RotationSouth,
	}); err != nil {
		t.Fatalf("load world: %v", err)
	}
	if _, err := runtime.Join(context.Background(), 9, roomlive.Occupant{
		PlayerID: 7, Username: "demo", ConnectionID: netconn.ID("conn"), ConnectionKind: netconn.Kind("websocket"),
	}); err != nil {
		t.Fatalf("join runtime: %v", err)
	}

	connections := netconn.NewRegistry()

	return Handler{Players: players, Bindings: bindings, Runtime: runtime, Connections: connections}, connections
}

// settleUnitOnChair places a one-tile sit chair and walks a player onto it until they settle.
func settleUnitOnChair(t *testing.T, room *roomlive.Room, playerID int64, itemID int64) {
	t.Helper()

	point := grid.MustPoint(2, 0)
	item := worldfurniture.Item{
		ID: itemID,
		Definition: worldfurniture.Definition{
			Width: 1, Length: 1, StackHeight: 0, AllowSit: true,
			Slots: []worldfurniture.SlotDefinition{{Status: worldfurniture.SlotStatusSit, BodyRotation: worldunit.RotationNorth}},
		},
		Point: point, Rotation: worldunit.RotationNorth,
	}
	if _, err := room.ReloadFurniture(itemID, &item); err != nil {
		t.Fatalf("place chair: %v", err)
	}
	if _, err := room.MoveTo(playerID, point); err != nil {
		t.Fatalf("move to chair: %v", err)
	}
	for range 10 {
		room.Tick()
		if _, occupied := room.SlotOccupant(itemID); occupied {
			return
		}
	}
	t.Fatal("expected unit to settle onto chair")
}

// playersForTest creates a live player registry with one bound demo player.
func playersForTest(t *testing.T, inRoom bool) *playerlive.Registry {
	t.Helper()

	peer, err := playerlive.NewSessionPeer(netconn.ID("conn"), netconn.Kind("websocket"), time.Now())
	if err != nil {
		t.Fatalf("create peer: %v", err)
	}
	player, err := playerlive.NewPlayer(playerlive.Snapshot{ID: 7, Username: "demo"}, peer)
	if err != nil {
		t.Fatalf("create player: %v", err)
	}
	if inRoom {
		if err := player.EnterRoom(9); err != nil {
			t.Fatalf("enter room: %v", err)
		}
	}
	registry := playerlive.NewRegistry()
	if err := registry.Add(player); err != nil {
		t.Fatalf("add player: %v", err)
	}

	return registry
}

// connectionForTest creates a bare connection context without an attached session.
func connectionForTest() netconn.Context {
	return netconn.Context{ConnectionID: netconn.ID("conn"), ConnectionKind: netconn.Kind("websocket")}
}

// registeredConnectionForTest creates a session-backed connection registered for room broadcasts.
func registeredConnectionForTest(t *testing.T, connections *netconn.Registry, id netconn.ID) (netconn.Context, *[]codec.Packet) {
	t.Helper()

	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error {
		return nil
	}, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	inbound := netconn.NewHandlerRegistry()
	var captured netconn.Context
	if err := inbound.Register(1, func(context netconn.Context, _ codec.Packet) error {
		captured = context

		return nil
	}, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated()); err != nil {
		t.Fatalf("register inbound: %v", err)
	}

	sent := make([]codec.Packet, 0, 2)
	session, err := netconn.NewSession(netconn.SessionConfig{
		ID:       id,
		Kind:     netconn.Kind("websocket"),
		Inbound:  inbound,
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
	if err := session.Receive(context.Background(), codec.Packet{Header: 1}); err != nil {
		t.Fatalf("receive context packet: %v", err)
	}

	return captured, &sent
}

// pickedItemForTest returns a picked up item fixture returned to inventory.
func pickedItemForTest() furnituremodel.Item {
	return furnituremodel.Item{
		Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 1}}, DefinitionID: 2, OwnerPlayerID: 7,
	}
}

// fakeManager stubs furniture persistence for tests.
type fakeManager struct {
	definition      furnituremodel.Definition
	definitionFound bool
	pickupResult    furnituremodel.Item
	pickupErr       error
	pickupCalls     int
}

// FindDefinitionByID finds a definition for tests.
func (manager *fakeManager) FindDefinitionByID(context.Context, int64) (furnituremodel.Definition, bool, error) {
	return manager.definition, manager.definitionFound, nil
}

// ListDefinitions lists definitions for tests.
func (manager *fakeManager) ListDefinitions(context.Context) ([]furnituremodel.Definition, error) {
	return nil, nil
}

// FindItemByID finds an item for tests.
func (manager *fakeManager) FindItemByID(context.Context, int64) (furnituremodel.Item, bool, error) {
	return furnituremodel.Item{}, false, nil
}

// ListInventory lists inventory items for tests.
func (manager *fakeManager) ListInventory(context.Context, int64) ([]furnituremodel.Item, error) {
	return nil, nil
}

// ListRoomItems lists room items for tests.
func (manager *fakeManager) ListRoomItems(context.Context, int64) ([]furnituremodel.Item, error) {
	return nil, nil
}

// Place places an item for tests.
func (manager *fakeManager) Place(context.Context, furnitureservice.PlaceParams) (furnituremodel.Item, error) {
	return furnituremodel.Item{}, nil
}

// Move moves an item for tests.
func (manager *fakeManager) Move(context.Context, furnitureservice.MoveParams) (furnituremodel.Item, error) {
	return furnituremodel.Item{}, nil
}

// Pickup picks up an item for tests.
func (manager *fakeManager) Pickup(context.Context, furnitureservice.PickupParams) (furnituremodel.Item, error) {
	manager.pickupCalls++
	if manager.pickupErr != nil {
		return furnituremodel.Item{}, manager.pickupErr
	}

	return manager.pickupResult, nil
}
