package pickup

import (
	"context"
	"testing"
	"time"

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
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// handlerForTest creates a pickup command handler with a loaded room and joined player.
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
