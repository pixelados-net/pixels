package inventory

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/command"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outlist "github.com/niflaot/pixels/networking/outbound/inventory/furniture/list"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// TestHandleSendsSingleFragmentForSmallInventory verifies the common single-fragment path.
func TestHandleSendsSingleFragmentForSmallInventory(t *testing.T) {
	connection, sent := connectionForTest(t)
	handler := Handler{
		Players:  playersForTest(t),
		Bindings: bindingsForTest(t),
		Furniture: &fakeManager{
			definitions: []furnituremodel.Definition{{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 2}}, SpriteID: 39, AllowInventoryStack: true}},
			items:       []furnituremodel.Item{{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 1}}, DefinitionID: 2, ExtraData: "0"}},
		},
	}

	err := handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{Handler: connection}})
	if err != nil {
		t.Fatalf("handle command: %v", err)
	}
	if len(*sent) != 1 || (*sent)[0].Header != outlist.Header {
		t.Fatalf("expected one inventory list packet, got %#v", *sent)
	}
}

// TestHandleSendsEmptyFragmentForEmptyInventory verifies the empty-inventory shape.
func TestHandleSendsEmptyFragmentForEmptyInventory(t *testing.T) {
	connection, sent := connectionForTest(t)
	handler := Handler{Players: playersForTest(t), Bindings: bindingsForTest(t), Furniture: &fakeManager{}}

	err := handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{Handler: connection}})
	if err != nil {
		t.Fatalf("handle command: %v", err)
	}
	if len(*sent) != 1 {
		t.Fatalf("expected one empty fragment packet, got %#v", *sent)
	}
}

// TestInventoryCategoryMapsRoomEffects verifies special room consumable categories.
func TestInventoryCategoryMapsRoomEffects(t *testing.T) {
	tests := map[string]outlist.Category{
		"chair": outlist.CategoryDefault, "wallpaper": outlist.CategoryWallpaper,
		"floor": outlist.CategoryFloor, "landscape": outlist.CategoryLandscape,
	}
	for name, expected := range tests {
		if actual := inventoryCategory(name); actual != expected {
			t.Fatalf("inventoryCategory(%q) = %d, want %d", name, actual, expected)
		}
	}
}

// TestHandleSendsMultipleFragmentsForLargeInventory verifies pagination beyond 1000 items.
func TestHandleSendsMultipleFragmentsForLargeInventory(t *testing.T) {
	connection, sent := connectionForTest(t)
	items := make([]furnituremodel.Item, 0, 1500)
	for index := int64(1); index <= 1500; index++ {
		items = append(items, furnituremodel.Item{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: index}}, DefinitionID: 2})
	}
	handler := Handler{
		Players:  playersForTest(t),
		Bindings: bindingsForTest(t),
		Furniture: &fakeManager{
			definitions: []furnituremodel.Definition{{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 2}}, SpriteID: 39}},
			items:       items,
		},
	}

	err := handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{Handler: connection}})
	if err != nil {
		t.Fatalf("handle command: %v", err)
	}
	if len(*sent) != 2 {
		t.Fatalf("expected two fragments, got %d", len(*sent))
	}
}

// TestHandlePropagatesStoreErrors verifies persistence errors surface.
func TestHandlePropagatesStoreErrors(t *testing.T) {
	connection, _ := connectionForTest(t)
	expected := errors.New("store failed")
	handler := Handler{Players: playersForTest(t), Bindings: bindingsForTest(t), Furniture: &fakeManager{itemsErr: expected}}

	err := handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{Handler: connection}})
	if !errors.Is(err, expected) {
		t.Fatalf("expected store error, got %v", err)
	}
}

// playersForTest creates a live player registry with one bound demo player.
func playersForTest(t *testing.T) *playerlive.Registry {
	t.Helper()

	peer, err := playerlive.NewSessionPeer(netconn.ID("conn"), netconn.Kind("websocket"), time.Now())
	if err != nil {
		t.Fatalf("create peer: %v", err)
	}
	player, err := playerlive.NewPlayer(playerlive.Snapshot{ID: 7, Username: "demo"}, peer)
	if err != nil {
		t.Fatalf("create player: %v", err)
	}
	registry := playerlive.NewRegistry()
	if err := registry.Add(player); err != nil {
		t.Fatalf("add player: %v", err)
	}

	return registry
}

// bindingsForTest creates a connection binding registry for the demo player.
func bindingsForTest(t *testing.T) *binding.Registry {
	t.Helper()

	registry := binding.NewRegistry()
	if err := registry.Add(binding.Binding{PlayerID: 7, ConnectionID: netconn.ID("conn"), ConnectionKind: netconn.Kind("websocket")}); err != nil {
		t.Fatalf("add binding: %v", err)
	}

	return registry
}

// connectionForTest creates a session-backed connection context that records sent packets.
func connectionForTest(t *testing.T) (netconn.Context, *[]codec.Packet) {
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
		ID:       netconn.ID("conn"),
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
	if err := session.Receive(context.Background(), codec.Packet{Header: 1}); err != nil {
		t.Fatalf("receive context packet: %v", err)
	}

	return captured, &sent
}

// fakeManager stubs furniture persistence for tests.
type fakeManager struct {
	definitions []furnituremodel.Definition
	items       []furnituremodel.Item
	itemsErr    error
}

// FindDefinitionByID finds a definition for tests.
func (manager *fakeManager) FindDefinitionByID(context.Context, int64) (furnituremodel.Definition, bool, error) {
	return furnituremodel.Definition{}, false, nil
}

// ListDefinitions lists definitions for tests.
func (manager *fakeManager) ListDefinitions(context.Context) ([]furnituremodel.Definition, error) {
	return manager.definitions, nil
}

// FindItemByID finds an item for tests.
func (manager *fakeManager) FindItemByID(context.Context, int64) (furnituremodel.Item, bool, error) {
	return furnituremodel.Item{}, false, nil
}

// ListInventory lists inventory items for tests.
func (manager *fakeManager) ListInventory(context.Context, int64) ([]furnituremodel.Item, error) {
	return manager.items, manager.itemsErr
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
	return furnituremodel.Item{}, nil
}
