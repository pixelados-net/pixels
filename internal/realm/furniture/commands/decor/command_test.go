package decor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/command"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomdecor "github.com/niflaot/pixels/internal/realm/room/decoration"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// TestDecorationSoftErrorClassifiesClientFailures verifies only expected client failures are suppressed.
func TestDecorationSoftErrorClassifiesClientFailures(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{name: "surface", err: roomdecor.ErrInvalidSurface, expected: true},
		{name: "surface value", err: roomdecor.ErrInvalidSurfaceValue, expected: true},
		{name: "wall", err: roomdecor.ErrInvalidWallPosition, expected: true},
		{name: "dimmer", err: roomdecor.ErrInvalidDimmerPreset, expected: true},
		{name: "unavailable", err: roomdecor.ErrDecorationUnavailable, expected: true},
		{name: "wrapped", err: errors.Join(errors.New("request"), roomdecor.ErrDecorationUnavailable), expected: true},
		{name: "database", err: errors.New("database unavailable"), expected: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if actual := decorationSoftError(test.err); actual != test.expected {
				t.Fatalf("decorationSoftError() = %t, want %t", actual, test.expected)
			}
		})
	}
}

// TestHandleIgnoresUnknownKinds verifies future decorator packets remain non-fatal.
func TestHandleIgnoresUnknownKinds(t *testing.T) {
	handler, connection, _, _ := decoratorFixture(t)
	err := handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{Handler: connection, Kind: 255}})
	if err != nil {
		t.Fatalf("ignore unknown decoration kind: %v", err)
	}
	if (Command{}).CommandName() != Name {
		t.Fatalf("unexpected command name %s", (Command{}).CommandName())
	}
}

// decoratorFixture creates an authenticated room owner and captured transport.
func decoratorFixture(t *testing.T) (Handler, netconn.Context, *[]codec.Packet, *roomlive.Room) {
	return decoratorFixtureForOwner(t, 7)
}

// decoratorFixtureForOwner creates an authenticated player in a room with the selected owner.
func decoratorFixtureForOwner(t *testing.T, ownerPlayerID int64) (Handler, netconn.Context, *[]codec.Packet, *roomlive.Room) {
	t.Helper()
	peer, err := playerlive.NewSessionPeer("decor", "websocket", time.Now())
	if err != nil {
		t.Fatalf("create peer: %v", err)
	}
	player, err := playerlive.NewPlayer(playerlive.Snapshot{ID: 7, Username: "demo", Gender: "M", Look: "hd-180-1.ch-1-1"}, peer)
	if err != nil {
		t.Fatalf("create player: %v", err)
	}
	if err = player.EnterRoom(9); err != nil {
		t.Fatalf("enter room: %v", err)
	}
	players := playerlive.NewRegistry()
	if err = players.Add(player); err != nil {
		t.Fatalf("add player: %v", err)
	}
	bindings := binding.NewRegistry()
	if err = bindings.Add(binding.Binding{PlayerID: 7, ConnectionID: "decor", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("bind player: %v", err)
	}
	runtime := roomlive.NewRegistry(nil)
	active, err := runtime.Activate(roomlive.Snapshot{ID: 9, OwnerPlayerID: ownerPlayerID, MaxUsers: 25})
	if err != nil {
		t.Fatalf("activate room: %v", err)
	}
	roomGrid, err := grid.Parse("000", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatalf("parse grid: %v", err)
	}
	if err = active.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}, Body: worldunit.RotationSouth, Head: worldunit.RotationSouth}); err != nil {
		t.Fatalf("load world: %v", err)
	}
	if _, err = runtime.Join(context.Background(), 9, roomlive.Occupant{PlayerID: 7, Username: "demo", ConnectionID: "decor", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("join room: %v", err)
	}
	connections := netconn.NewRegistry()
	connection, sent := decoratorConnection(t, connections)
	store := &decorationStore{changed: true, found: true, state: roomdecor.DimmerState{ItemID: 1, ExtraData: "2,1,1,#000000,255", Presets: []roomdecor.Preset{{ID: 1, Color: "#000000", Brightness: 255}}}}
	return Handler{Players: players, Bindings: bindings, Decoration: roomdecor.New(store), Runtime: runtime, Connections: connections}, connection, sent, active
}

// decoratorConnection registers one session and captures its handler context.
func decoratorConnection(t *testing.T, connections *netconn.Registry) (netconn.Context, *[]codec.Packet) {
	t.Helper()
	inbound := netconn.NewHandlerRegistry()
	outbound := netconn.NewHandlerRegistry()
	var captured netconn.Context
	if err := inbound.Register(1, func(value netconn.Context, _ codec.Packet) error { captured = value; return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated()); err != nil {
		t.Fatalf("register capture handler: %v", err)
	}
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	sent := make([]codec.Packet, 0, 4)
	session, err := netconn.NewSession(netconn.SessionConfig{ID: "decor", Kind: "websocket", Inbound: inbound, Outbound: outbound, Sender: func(_ context.Context, packet codec.Packet) error { sent = append(sent, packet); return nil }, Disposer: func(context.Context, netconn.Reason) error { return nil }})
	if err != nil {
		t.Fatalf("create connection: %v", err)
	}
	if err = connections.Register(session); err != nil {
		t.Fatalf("register connection: %v", err)
	}
	if err = session.Receive(context.Background(), codec.Packet{Header: 1}); err != nil {
		t.Fatalf("capture context: %v", err)
	}
	return captured, &sent
}

// furnitureManager stores deterministic decorator furniture records.
type furnitureManager struct {
	// item stores the selected furniture item.
	item furnituremodel.Item
	// definition stores the selected furniture definition.
	definition furnituremodel.Definition
	// inventory stores post-operation inventory rows.
	inventory []furnituremodel.Item
}

// FindDefinitionByID returns the configured definition.
func (manager *furnitureManager) FindDefinitionByID(context.Context, int64) (furnituremodel.Definition, bool, error) {
	return manager.definition, true, nil
}

// ListDefinitions returns the configured definition.
func (manager *furnitureManager) ListDefinitions(context.Context) ([]furnituremodel.Definition, error) {
	return []furnituremodel.Definition{manager.definition}, nil
}

// FindItemByID returns the configured item.
func (manager *furnitureManager) FindItemByID(context.Context, int64) (furnituremodel.Item, bool, error) {
	return manager.item, true, nil
}

// ListInventory returns the configured inventory.
func (manager *furnitureManager) ListInventory(context.Context, int64) ([]furnituremodel.Item, error) {
	return manager.inventory, nil
}

// ListRoomItems returns the configured placed item.
func (manager *furnitureManager) ListRoomItems(context.Context, int64) ([]furnituremodel.Item, error) {
	return []furnituremodel.Item{manager.item}, nil
}

// Place returns the configured item.
func (manager *furnitureManager) Place(context.Context, furnitureservice.PlaceParams) (furnituremodel.Item, error) {
	return manager.item, nil
}

// Move returns the configured item.
func (manager *furnitureManager) Move(context.Context, furnitureservice.MoveParams) (furnituremodel.Item, error) {
	return manager.item, nil
}

// Pickup returns the configured item.
func (manager *furnitureManager) Pickup(context.Context, furnitureservice.PickupParams) (furnituremodel.Item, error) {
	return manager.item, nil
}

// stateUpdater stores the last decorator item state.
type stateUpdater struct {
	// item stores the returned mutation record.
	item furnituremodel.Item
}

// UpdateState applies the requested state to the configured item.
func (updater *stateUpdater) UpdateState(_ context.Context, params furnitureservice.StateParams) (furnituremodel.Item, error) {
	updater.item.ExtraData = params.Next
	return updater.item, nil
}

// decorationStore stores deterministic room decoration state.
type decorationStore struct {
	// changed controls guarded mutations.
	changed bool
	// found controls dimmer reads.
	found bool
	// state stores the current dimmer state.
	state roomdecor.DimmerState
}

// ConsumeSurface returns the configured mutation decision.
func (store *decorationStore) ConsumeSurface(context.Context, int64, int64, int64, roomdecor.Surface, string) (bool, error) {
	return store.changed, nil
}

// PlacePostIt returns the configured mutation decision.
func (store *decorationStore) PlacePostIt(context.Context, int64, int64, int64, string, string) (bool, error) {
	return store.changed, nil
}

// LoadDimmer returns the configured dimmer state.
func (store *decorationStore) LoadDimmer(context.Context, int64) (roomdecor.DimmerState, bool, error) {
	return store.state, store.found, nil
}

// SaveDimmer stores and returns the requested preset.
func (store *decorationStore) SaveDimmer(_ context.Context, _ int64, _ int64, preset roomdecor.Preset, apply bool) (roomdecor.DimmerState, bool, error) {
	store.state.Presets = []roomdecor.Preset{preset}
	store.state.ExtraData = "2,1,1," + preset.Color + ",255"
	return store.state, store.changed, nil
}

// ToggleDimmer returns the configured dimmer state.
func (store *decorationStore) ToggleDimmer(context.Context, int64, int64) (roomdecor.DimmerState, bool, error) {
	return store.state, store.changed, nil
}

// inventoryDecoratorItem creates an owned unplaced decorator item.
func inventoryDecoratorItem(id int64, definitionID int64, extraData string) furnituremodel.Item {
	return furnituremodel.Item{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: id}}, DefinitionID: definitionID, OwnerPlayerID: 7, ExtraData: extraData}
}
