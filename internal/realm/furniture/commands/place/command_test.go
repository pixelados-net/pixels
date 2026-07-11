package place

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/command"
	placedevent "github.com/niflaot/pixels/internal/realm/furniture/events/placed"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	roomfurniture "github.com/niflaot/pixels/internal/realm/room/world/items"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outinvremove "github.com/niflaot/pixels/networking/outbound/inventory/furniture/remove"
	outadd "github.com/niflaot/pixels/networking/outbound/room/furniture/add"
	outheightmapupdate "github.com/niflaot/pixels/networking/outbound/room/heightmapupdate"
	outbubble "github.com/niflaot/pixels/networking/outbound/session/bubblealert"
	"github.com/niflaot/pixels/pkg/bus"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
	"go.uber.org/zap"
)

// TestHandlePlacesItemAndBroadcasts verifies the successful place path.
func TestHandlePlacesItemAndBroadcasts(t *testing.T) {
	handler, connections := handlerForTest(t)
	connection, sent := registeredConnectionForTest(t, connections, "conn")
	handler.Furniture = &fakeManager{
		definition: chairDefinitionForTest(), definitionFound: true,
		item: inventoryItemForTest(), itemFound: true,
	}

	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: connection, ItemID: 1, X: 3, Y: 0, Rotation: 0},
	})
	if err != nil {
		t.Fatalf("handle place: %v", err)
	}

	room, _ := handler.Runtime.Find(9)
	if items := room.FurnitureItems(); len(items) != 1 || items[0].Point.X != 3 {
		t.Fatalf("unexpected furniture items %#v", items)
	}
	if len(*sent) != 3 || (*sent)[0].Header != outinvremove.Header || (*sent)[1].Header != outadd.Header || (*sent)[2].Header != outheightmapupdate.Header {
		t.Fatalf("expected inventory remove, floor add, then height map update packets, got %#v", *sent)
	}
}

// TestHandleBroadcastsAddToOtherOccupants verifies both connections in a room observe a placement.
func TestHandleBroadcastsAddToOtherOccupants(t *testing.T) {
	handler, connections := handlerForTest(t)
	actor, actorSent := registeredConnectionForTest(t, connections, "conn")
	_, bystanderSent := registeredConnectionForTest(t, connections, "bystander-conn")
	room, _ := handler.Runtime.Find(9)
	if _, err := room.Join(roomlive.Occupant{
		PlayerID: 8, Username: "bystander", ConnectionID: netconn.ID("bystander-conn"), ConnectionKind: netconn.Kind("websocket"),
	}); err != nil {
		t.Fatalf("join bystander: %v", err)
	}
	handler.Furniture = &fakeManager{
		definition: chairDefinitionForTest(), definitionFound: true,
		item: inventoryItemForTest(), itemFound: true,
	}

	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: actor, ItemID: 1, X: 3, Y: 0, Rotation: 0},
	})
	if err != nil {
		t.Fatalf("handle place: %v", err)
	}

	if len(*actorSent) != 3 || (*actorSent)[0].Header != outinvremove.Header || (*actorSent)[1].Header != outadd.Header || (*actorSent)[2].Header != outheightmapupdate.Header {
		t.Fatalf("expected actor to receive inventory remove, floor add, then height map update, got %#v", *actorSent)
	}
	if len(*bystanderSent) != 2 || (*bystanderSent)[0].Header != outadd.Header || (*bystanderSent)[1].Header != outheightmapupdate.Header {
		t.Fatalf("expected bystander to receive floor add then height map update, got %#v", *bystanderSent)
	}
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

// TestHandleIgnoresSoftPlacementErrors verifies gameplay misses stay silent and send a bubble alert.
func TestHandleIgnoresSoftPlacementErrors(t *testing.T) {
	handler, connections := handlerForTest(t)
	connection, sent := registeredConnectionForTest(t, connections, "conn")
	handler.Furniture = &fakeManager{item: inventoryItemForTest(), itemFound: true}

	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: connection, ItemID: 1, X: 3, Y: 0, Rotation: 0},
	})
	if err != nil {
		t.Fatalf("expected soft placement error to stay silent, got %v", err)
	}
	if len(*sent) != 1 || (*sent)[0].Header != outbubble.Header {
		t.Fatalf("expected a bubble alert packet, got %#v", *sent)
	}
}

// TestHandleReturnsNilWhenItemNotFound verifies missing items stay silent.
func TestHandleReturnsNilWhenItemNotFound(t *testing.T) {
	handler, _ := handlerForTest(t)
	handler.Furniture = &fakeManager{}

	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: connectionForTest(), ItemID: 1, X: 3, Y: 0, Rotation: 0},
	})
	if err != nil {
		t.Fatalf("expected no error for missing item, got %v", err)
	}
}

// TestHandleSkipsBroadcastWhenConnectionsNil verifies a nil connection registry is tolerated.
func TestHandleSkipsBroadcastWhenConnectionsNil(t *testing.T) {
	handler, standaloneConnections := handlerForTest(t)
	connection, _ := registeredConnectionForTest(t, standaloneConnections, "conn")
	handler.Connections = nil
	handler.Furniture = &fakeManager{
		definition: chairDefinitionForTest(), definitionFound: true,
		item: inventoryItemForTest(), itemFound: true,
	}

	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: connection, ItemID: 1, X: 3, Y: 0, Rotation: 0},
	})
	if err != nil {
		t.Fatalf("expected no error with nil connections, got %v", err)
	}
}

// TestHandlePropagatesPlaceHardErrors verifies unexpected persistence errors surface.
func TestHandlePropagatesPlaceHardErrors(t *testing.T) {
	handler, _ := handlerForTest(t)
	expected := errors.New("place failed")
	handler.Furniture = &fakeManager{
		definition: chairDefinitionForTest(), definitionFound: true,
		item: inventoryItemForTest(), itemFound: true, placeErr: expected,
	}

	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: connectionForTest(), ItemID: 1, X: 3, Y: 0, Rotation: 0},
	})
	if !errors.Is(err, expected) {
		t.Fatalf("expected place error, got %v", err)
	}
}

// TestHandlePropagatesHardErrors verifies unexpected store errors surface.
func TestHandlePropagatesHardErrors(t *testing.T) {
	handler, _ := handlerForTest(t)
	expected := errors.New("store failed")
	handler.Furniture = &fakeManager{findItemErr: expected}

	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: connectionForTest(), ItemID: 1, X: 3, Y: 0, Rotation: 0},
	})
	if !errors.Is(err, expected) {
		t.Fatalf("expected store error, got %v", err)
	}
}

// TestHandlePlacesAndPublishesWithLogger verifies event publication and logging on success.
func TestHandlePlacesAndPublishesWithLogger(t *testing.T) {
	handler, connections := handlerForTest(t)
	connection, _ := registeredConnectionForTest(t, connections, "conn")
	publisher := &fakePublisher{}
	handler.Events = publisher
	handler.Log = zap.NewNop()
	handler.Furniture = &fakeManager{
		definition: chairDefinitionForTest(), definitionFound: true,
		item: inventoryItemForTest(), itemFound: true,
	}

	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: connection, ItemID: 1, X: 3, Y: 0, Rotation: 0},
	})
	if err != nil {
		t.Fatalf("handle place: %v", err)
	}
	if len(publisher.events) != 1 || publisher.events[0].Name != placedevent.Name {
		t.Fatalf("expected placed event published, got %#v", publisher.events)
	}
}

// TestHandleLogsRejectionWithLogger verifies rejected placements log without error when a logger is set.
func TestHandleLogsRejectionWithLogger(t *testing.T) {
	handler, connections := handlerForTest(t)
	connection, _ := registeredConnectionForTest(t, connections, "conn")
	handler.Log = zap.NewNop()
	handler.Furniture = &fakeManager{item: inventoryItemForTest(), itemFound: true}

	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: connection, ItemID: 1, X: 3, Y: 0, Rotation: 0},
	})
	if err != nil {
		t.Fatalf("expected soft placement error to stay silent, got %v", err)
	}
}

// TestBubbleErrorKeyMapsKnownErrors verifies every mapped soft error resolves to its bubble key.
func TestBubbleErrorKeyMapsKnownErrors(t *testing.T) {
	cases := []struct {
		err      error
		wantKey  string
		wantSoft bool
	}{
		{roomlive.ErrInvalidPlacement, "session.bubble.furniture.invalid_move", true},
		{roomfurniture.ErrInvalidTarget, "session.bubble.furniture.invalid_move", true},
		{furnitureservice.ErrInvalidPlacement, "session.bubble.furniture.invalid_move", true},
		{furnitureservice.ErrInvalidItemID, "session.bubble.furniture.invalid_move", true},
		{furnitureservice.ErrInvalidRoomID, "session.bubble.furniture.invalid_move", true},
		{furnitureservice.ErrInvalidPlayerID, "session.bubble.furniture.invalid_move", true},
		{roomlive.ErrTileOccupied, "session.bubble.furniture.tile_has_units", true},
		{roomlive.ErrCannotStack, "session.bubble.furniture.cant_stack", true},
		{furnitureservice.ErrNotItemOwner, "session.bubble.furniture.no_rights", true},
		{furnitureservice.ErrItemNotInInventory, "session.bubble.furniture.item_not_in_inventory", true},
		{furnitureservice.ErrItemNotFound, "session.bubble.furniture.item_not_found", true},
		{furnitureservice.ErrItemNotPlaced, "session.bubble.furniture.item_not_found", true},
		{roomfurniture.ErrDefinitionNotFound, "session.bubble.furniture.item_not_found", true},
		{roomlive.ErrWorldNotLoaded, "", true},
		{errors.New("unmapped"), "", false},
	}

	for _, testCase := range cases {
		key, soft := bubbleErrorKey(testCase.err)
		if key != testCase.wantKey || soft != testCase.wantSoft {
			t.Fatalf("bubbleErrorKey(%v) = (%q, %v), want (%q, %v)", testCase.err, key, soft, testCase.wantKey, testCase.wantSoft)
		}
	}
}

// TestHandleSoftErrorReturnsNilWithoutBubbleForWorldNotLoaded verifies the no-bubble soft error path.
func TestHandleSoftErrorReturnsNilWithoutBubbleForWorldNotLoaded(t *testing.T) {
	handler := Handler{Log: zap.NewNop()}

	err := handler.handleSoftError(context.Background(), Command{Handler: connectionForTest()}, roomlive.ErrWorldNotLoaded)
	if err != nil {
		t.Fatalf("expected nil for world-not-loaded soft error, got %v", err)
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

// TestCommandName verifies the stable command name.
func TestCommandName(t *testing.T) {
	if (Command{}).CommandName() != Name {
		t.Fatalf("unexpected command name %s", (Command{}).CommandName())
	}
}

// handlerForTest creates a place command handler with a loaded room and joined player.
func handlerForTest(t *testing.T) (Handler, *netconn.Registry) {
	t.Helper()

	players := playersForTest(t, true)
	bindings := binding.NewRegistry()
	if err := bindings.Add(binding.Binding{PlayerID: 7, ConnectionID: netconn.ID("conn"), ConnectionKind: netconn.Kind("websocket")}); err != nil {
		t.Fatalf("add binding: %v", err)
	}

	runtime := roomlive.NewRegistry(nil)
	room, err := runtime.Activate(roomlive.Snapshot{ID: 9, OwnerPlayerID: 7, MaxUsers: 10})
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

// chairDefinitionForTest returns a placeable sit definition fixture.
func chairDefinitionForTest() furnituremodel.Definition {
	return furnituremodel.Definition{
		Base:  sharedmodel.Base{Identity: sharedmodel.Identity{ID: 2}},
		Width: 1, Length: 1, StackHeight: 1, AllowSit: true,
	}
}

// inventoryItemForTest returns an unplaced item fixture owned by the demo player.
func inventoryItemForTest() furnituremodel.Item {
	return furnituremodel.Item{
		Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 1}}, DefinitionID: 2, OwnerPlayerID: 7,
	}
}

// fakeManager stubs furniture persistence for tests.
type fakeManager struct {
	definition      furnituremodel.Definition
	definitionFound bool
	findDefErr      error
	item            furnituremodel.Item
	itemFound       bool
	findItemErr     error
	placeErr        error
}

// FindDefinitionByID finds a definition for tests.
func (manager *fakeManager) FindDefinitionByID(context.Context, int64) (furnituremodel.Definition, bool, error) {
	return manager.definition, manager.definitionFound, manager.findDefErr
}

// ListDefinitions lists definitions for tests.
func (manager *fakeManager) ListDefinitions(context.Context) ([]furnituremodel.Definition, error) {
	return nil, nil
}

// FindItemByID finds an item for tests.
func (manager *fakeManager) FindItemByID(context.Context, int64) (furnituremodel.Item, bool, error) {
	return manager.item, manager.itemFound, manager.findItemErr
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
func (manager *fakeManager) Place(_ context.Context, params furnitureservice.PlaceParams) (furnituremodel.Item, error) {
	if manager.placeErr != nil {
		return furnituremodel.Item{}, manager.placeErr
	}

	x, y, z, roomID := params.Placement.X, params.Placement.Y, params.Placement.Z, params.RoomID

	return furnituremodel.Item{
		Base:         sharedmodel.Base{Identity: sharedmodel.Identity{ID: params.ItemID}},
		DefinitionID: manager.item.DefinitionID, OwnerPlayerID: params.ActorPlayerID,
		RoomID: &roomID, X: &x, Y: &y, Z: &z, Rotation: params.Placement.Rotation,
	}, nil
}

// Move moves an item for tests.
func (manager *fakeManager) Move(context.Context, furnitureservice.MoveParams) (furnituremodel.Item, error) {
	return furnituremodel.Item{}, nil
}

// Pickup picks up an item for tests.
func (manager *fakeManager) Pickup(context.Context, furnitureservice.PickupParams) (furnituremodel.Item, error) {
	return furnituremodel.Item{}, nil
}
