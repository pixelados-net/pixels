package floorplan

import (
	"context"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/command"
	"github.com/niflaot/pixels/internal/permission"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	floorplansaved "github.com/niflaot/pixels/internal/realm/room/control/events/floorplansaved"
	domain "github.com/niflaot/pixels/internal/realm/room/control/floorplan"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	roomlayout "github.com/niflaot/pixels/internal/realm/room/world/layout"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outforward "github.com/niflaot/pixels/networking/outbound/room/forward"
	"github.com/niflaot/pixels/pkg/bus"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// roomsForTest returns one persistent room.
type roomsForTest struct {
	// room stores the returned room.
	room roommodel.Room
}

// FindByID finds one room for tests.
func (finder roomsForTest) FindByID(context.Context, int64) (roommodel.Room, bool, error) {
	return finder.room, true, nil
}

// layoutsForTest stores one custom layout mutation.
type layoutsForTest struct {
	roomlayout.Manager
	// saved stores the latest custom layout input.
	saved roomlayout.CustomSaveParams
}

// ResolveForRoom returns fixed starting geometry.
func (manager *layoutsForTest) ResolveForRoom(context.Context, int64, string) (roomlayout.Layout, error) {
	return roomlayout.Layout{Name: "model_a", Heightmap: "00", DoorX: 0, DoorY: 0, DoorDirection: 2, WallHeight: -1}, nil
}

// SaveCustom captures and returns custom geometry.
func (manager *layoutsForTest) SaveCustom(_ context.Context, params roomlayout.CustomSaveParams) (roomlayout.Layout, error) {
	manager.saved = params
	return roomlayout.Layout{
		RoomID: params.RoomID, Name: "model_a", Heightmap: params.Heightmap,
		DoorX: params.DoorX, DoorY: params.DoorY, DoorZ: params.DoorZ,
		DoorDirection: params.DoorDirection, WallThickness: params.WallThickness,
		FloorThickness: params.FloorThickness, WallHeight: params.WallHeight,
	}, nil
}

// WithinTransaction runs custom layout work immediately.
func (manager *layoutsForTest) WithinTransaction(ctx context.Context, work roomlayout.TransactionWork) error {
	return work(ctx)
}

// permissionsForTest allows configured nodes.
type permissionsForTest map[permission.Node]bool

// HasPermission resolves one permission node.
func (permissions permissionsForTest) HasPermission(_ context.Context, _ int64, node permission.Node) (bool, error) {
	return permissions[node], nil
}

// eventsForTest captures published events.
type eventsForTest struct {
	// events stores published events.
	events []bus.Event
}

// Publish captures one event.
func (publisher *eventsForTest) Publish(_ context.Context, event bus.Event) error {
	publisher.events = append(publisher.events, event)

	return nil
}

// TestSaveHandlePersistsAndPublishes verifies the inactive-room save transaction.
func TestSaveHandlePersistsAndPublishes(t *testing.T) {
	players, bindings, player := floorplanActorForTest(t)
	if err := player.EnterRoom(9); err != nil {
		t.Fatalf("enter room: %v", err)
	}
	room := roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}}, OwnerPlayerID: 7, ModelName: "model_a"}
	layouts := &layoutsForTest{}
	events := &eventsForTest{}
	authorizer := domain.NewAuthorizer(permissionsForTest{"own": true}, nil, domain.Nodes{OwnEdit: "own", AnyEdit: "any"})
	handler := SaveHandler{
		Players: players, Bindings: bindings, Rooms: roomsForTest{room: room}, Layouts: layouts,
		Runtime: roomlive.NewRegistry(nil), Authorize: authorizer,
		Config: domain.Config{RejectZeroEffectiveHeight: true}, Events: events,
	}
	input := SaveCommand{Handler: netconn.Context{ConnectionID: "conn", ConnectionKind: "websocket"}, Params: domain.SaveParams{
		Heightmap: "10", DoorX: 0, DoorY: 0, DoorDirection: 4,
		WallThickness: -1, FloorThickness: 1, WallHeight: 3,
	}}
	if input.CommandName() != SaveName {
		t.Fatalf("unexpected command name %q", input.CommandName())
	}
	if err := handler.Handle(context.Background(), command.Envelope[SaveCommand]{Command: input}); err != nil {
		t.Fatalf("save floor plan: %v", err)
	}
	if layouts.saved.RoomID != 9 || layouts.saved.DoorZ != 1 || layouts.saved.WallHeight != 3 {
		t.Fatalf("unexpected save %#v", layouts.saved)
	}
	if len(events.events) != 1 || events.events[0].Name != floorplansaved.Name {
		t.Fatalf("unexpected events %#v", events.events)
	}
}

// TestSaveHandleReloadsAndForwardsActiveWorld verifies server reload and Nitro renderer rebuild.
func TestSaveHandleReloadsAndForwardsActiveWorld(t *testing.T) {
	players, bindings, player := floorplanActorForTest(t)
	if err := player.EnterRoom(9); err != nil {
		t.Fatalf("enter room: %v", err)
	}
	room := roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}}, OwnerPlayerID: 7, OwnerName: "demo", ModelName: "model_a", MaxUsers: 5}
	runtime := roomlive.NewRegistry(nil)
	active, err := runtime.Activate(roomlive.Snapshot{ID: 9, OwnerPlayerID: 7, MaxUsers: 5})
	if err != nil {
		t.Fatalf("activate room: %v", err)
	}
	initial, err := grid.Parse("00", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatalf("parse initial grid: %v", err)
	}
	if err = active.LoadWorld(roomlive.WorldConfig{Grid: initial, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}}); err != nil {
		t.Fatalf("load initial world: %v", err)
	}
	if _, err = active.Join(roomlive.Occupant{PlayerID: 7, Username: "demo", ConnectionID: "conn", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("join room: %v", err)
	}
	connections := netconn.NewRegistry()
	sent := registerFloorplanConnectionForTest(t, connections)
	layouts := &layoutsForTest{}
	authorizer := domain.NewAuthorizer(permissionsForTest{"own": true}, nil, domain.Nodes{OwnEdit: "own", AnyEdit: "any"})
	handler := SaveHandler{
		Players: players, Bindings: bindings, Rooms: roomsForTest{room: room}, Layouts: layouts,
		Furniture: &furnitureForTest{}, Runtime: runtime, Connections: connections, Authorize: authorizer,
		Config: domain.Config{RejectZeroEffectiveHeight: true},
	}
	input := SaveCommand{Handler: netconn.Context{ConnectionID: "conn", ConnectionKind: "websocket"}, Params: domain.SaveParams{
		Heightmap: "10", DoorX: 0, DoorY: 0, DoorDirection: 4, WallHeight: -1,
	}}
	if err = handler.Handle(context.Background(), command.Envelope[SaveCommand]{Command: input}); err != nil {
		t.Fatalf("save active floor plan: %v", err)
	}
	if len(*sent) != 1 || (*sent)[0].Header != outforward.Header {
		t.Fatalf("expected same-room forward, got %#v", *sent)
	}
	unit, found := active.Unit(7)
	if !found || unit.Position.Point != grid.MustPoint(0, 0) || unit.BodyRotation != 4 {
		t.Fatalf("unexpected reloaded unit %#v found=%v", unit, found)
	}
}

// registerFloorplanConnectionForTest registers one outbound packet capture.
func registerFloorplanConnectionForTest(t *testing.T, connections *netconn.Registry) *[]codec.Packet {
	t.Helper()
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	sent := make([]codec.Packet, 0, 7)
	session, err := netconn.NewSession(netconn.SessionConfig{
		ID: "conn", Kind: "websocket", Outbound: outbound,
		Sender: func(_ context.Context, packet codec.Packet) error {
			sent = append(sent, packet)
			return nil
		},
		Disposer: func(context.Context, netconn.Reason) error { return nil },
	})
	if err != nil {
		t.Fatalf("create connection: %v", err)
	}
	if err = connections.Register(session); err != nil {
		t.Fatalf("register connection: %v", err)
	}

	return &sent
}

// floorplanActorForTest creates one player and connection binding.
func floorplanActorForTest(t *testing.T) (*playerlive.Registry, *binding.Registry, *playerlive.Player) {
	t.Helper()
	players := playerlive.NewRegistry()
	peer, err := playerlive.NewSessionPeer("conn", "websocket", time.Now())
	if err != nil {
		t.Fatalf("create peer: %v", err)
	}
	player, err := playerlive.NewPlayer(playerlive.Snapshot{ID: 7, Username: "demo"}, peer)
	if err != nil {
		t.Fatalf("create player: %v", err)
	}
	if err = players.Add(player); err != nil {
		t.Fatalf("add player: %v", err)
	}
	bindings := binding.NewRegistry()
	if err = bindings.Add(binding.Binding{PlayerID: 7, ConnectionID: "conn", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("add binding: %v", err)
	}

	return players, bindings, player
}
