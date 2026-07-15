package interact

import (
	"context"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/command"
	furnitureused "github.com/niflaot/pixels/internal/realm/furniture/events/used"
	"github.com/niflaot/pixels/internal/realm/furniture/interactions"
	teleport "github.com/niflaot/pixels/internal/realm/furniture/interactions/teleport"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outstate "github.com/niflaot/pixels/networking/outbound/room/furniture/state"
	outbubble "github.com/niflaot/pixels/networking/outbound/session/bubblealert"
	"github.com/niflaot/pixels/pkg/bus"
)

// TestHandleCyclesStateBroadcastsAndPublishes verifies the complete generic toggle flow.
func TestHandleCyclesStateBroadcastsAndPublishes(t *testing.T) {
	handler, connection, sent, active := interactionHandlerForTest(t, 7, worldfurniture.Item{
		ID: 1, Point: grid.MustPoint(2, 0), ExtraData: "abc",
		Definition: worldfurniture.Definition{InteractionType: "default", InteractionModesCount: 3, Width: 1, Length: 1},
	})
	state := &stateUpdaterForTest{}
	handler.States = state
	used := false
	if _, err := handler.Events.(*bus.Bus).Subscribe(furnitureused.Name, bus.PriorityNormal, func(_ context.Context, event bus.Event) error {
		_, used = event.Payload.(furnitureused.Payload)
		return nil
	}); err != nil {
		t.Fatalf("subscribe used event: %v", err)
	}
	if err := handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{Handler: connection, ItemID: 1}}); err != nil {
		t.Fatalf("interact: %v", err)
	}
	item, _ := active.FurnitureItem(1)
	if item.ExtraData != "1" || state.params.Expected != "abc" || state.params.Next != "1" || !used {
		t.Fatalf("unexpected interaction item=%#v state=%#v used=%v", item, state.params, used)
	}
	if len(*sent) != 1 || (*sent)[0].Header != outstate.Header {
		t.Fatalf("expected compact state broadcast, got %#v", *sent)
	}
}

// TestHandleRejectsGateOccupiedByActor verifies occupied gates remain unchanged.
func TestHandleRejectsGateOccupiedByActor(t *testing.T) {
	handler, connection, sent, _ := interactionHandlerForTest(t, 7, worldfurniture.Item{
		ID: 1, Point: grid.MustPoint(0, 0), ExtraData: "1",
		Definition: worldfurniture.Definition{InteractionType: "gate", InteractionModesCount: 2, Width: 1, Length: 1},
	})
	state := &stateUpdaterForTest{}
	handler.States = state
	if err := handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{Handler: connection, ItemID: 1}}); err != nil {
		t.Fatalf("reject occupied gate: %v", err)
	}
	if state.calls != 0 || len(*sent) != 0 {
		t.Fatalf("expected inert occupied gate calls=%d packets=%#v", state.calls, *sent)
	}
}

// TestHandleRejectsGuestWithLocalizedFurnitureFeedback verifies shared rights enforcement.
func TestHandleRejectsGuestWithLocalizedFurnitureFeedback(t *testing.T) {
	handler, connection, sent, _ := interactionHandlerForTest(t, 99, worldfurniture.Item{
		ID: 1, Point: grid.MustPoint(2, 0), ExtraData: "0",
		Definition: worldfurniture.Definition{InteractionType: "default", InteractionModesCount: 2, Width: 1, Length: 1},
	})
	handler.States = &stateUpdaterForTest{}
	if err := handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{Handler: connection, ItemID: 1}}); err != nil {
		t.Fatalf("reject guest: %v", err)
	}
	if len(*sent) != 1 || (*sent)[0].Header != outbubble.Header {
		t.Fatalf("expected no-rights bubble, got %#v", *sent)
	}
}

// TestHandleDelegatesTeleportWithoutFurnitureRights verifies travel keeps its own policy.
func TestHandleDelegatesTeleportWithoutFurnitureRights(t *testing.T) {
	handler, connection, _, active := interactionHandlerForTest(t, 99, worldfurniture.Item{
		ID: 1, Point: grid.MustPoint(2, 0), ExtraData: "0",
		Definition: worldfurniture.Definition{InteractionType: "teleport", Width: 1, Length: 1},
	})
	starter := &teleporterForTest{}
	handler.Teleports = starter
	if err := handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{Handler: connection, ItemID: 1}}); err != nil {
		t.Fatalf("delegate teleport: %v", err)
	}
	if starter.request.PlayerID != 7 || starter.request.Room != active || starter.request.ItemID != 1 {
		t.Fatalf("unexpected teleport request %#v", starter.request)
	}
}

// TestHandleResyncsConcurrentStateConflict verifies durable state wins a click race.
func TestHandleResyncsConcurrentStateConflict(t *testing.T) {
	handler, connection, sent, active := interactionHandlerForTest(t, 7, worldfurniture.Item{
		ID: 1, Point: grid.MustPoint(2, 0), ExtraData: "0",
		Definition: worldfurniture.Definition{InteractionType: "default", InteractionModesCount: 2, Width: 1, Length: 1},
	})
	roomID := int64(9)
	handler.Furniture = managerForTest{item: furnituremodel.Item{RoomID: &roomID, ExtraData: "1"}, found: true}
	handler.States = &stateUpdaterForTest{err: furnitureservice.ErrStateConflict}
	if err := handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{Handler: connection, ItemID: 1}}); err != nil {
		t.Fatalf("resync conflict: %v", err)
	}
	item, _ := active.FurnitureItem(1)
	if item.ExtraData != "1" || len(*sent) != 1 || (*sent)[0].Header != outstate.Header {
		t.Fatalf("unexpected resync item=%#v packets=%#v", item, *sent)
	}
}

// TestHandlerHelpersCoverSilentBranches verifies unsupported and nil-event paths.
func TestHandlerHelpersCoverSilentBranches(t *testing.T) {
	handler, connection, _, active := interactionHandlerForTest(t, 7, worldfurniture.Item{
		ID: 1, Point: grid.MustPoint(2, 0), Definition: worldfurniture.Definition{InteractionType: "unknown", Width: 1, Length: 1},
	})
	if err := handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{Handler: connection, ItemID: 1}}); err != nil {
		t.Fatalf("ignore unsupported interaction: %v", err)
	}
	handler.Events = nil
	if err := handler.publish(context.Background(), 7, 1, active.ID()); err != nil {
		t.Fatalf("ignore nil publisher: %v", err)
	}
}

// TestHandleIgnoresRollerUse verifies clicks never toggle or animate autonomous rollers.
func TestHandleIgnoresRollerUse(t *testing.T) {
	handler, connection, sent, _ := interactionHandlerForTest(t, 7, worldfurniture.Item{
		ID: 1, Point: grid.MustPoint(2, 0), ExtraData: "0",
		Definition: worldfurniture.Definition{InteractionType: "roller", InteractionModesCount: 1, Width: 1, Length: 1, AllowWalk: true},
	})
	state := &stateUpdaterForTest{}
	handler.States = state
	if err := handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{Handler: connection, ItemID: 1}}); err != nil {
		t.Fatalf("ignore roller use: %v", err)
	}
	if state.calls != 0 || len(*sent) != 0 {
		t.Fatalf("roller click mutated state calls=%d packets=%#v", state.calls, *sent)
	}
}

// TestCommandName verifies the stable generic interaction command name.
func TestCommandName(t *testing.T) {
	if (Command{}).CommandName() != Name {
		t.Fatalf("unexpected command name %s", (Command{}).CommandName())
	}
}

// interactionHandlerForTest creates one loaded room, player, and outbound session.
func interactionHandlerForTest(t *testing.T, ownerID int64, item worldfurniture.Item) (Handler, netconn.Context, *[]codec.Packet, *roomlive.Room) {
	t.Helper()
	peer, err := playerlive.NewSessionPeer("conn", "websocket", time.Now())
	if err != nil {
		t.Fatalf("create peer: %v", err)
	}
	player, err := playerlive.NewPlayer(playerlive.Snapshot{ID: 7, Username: "demo"}, peer)
	if err != nil {
		t.Fatalf("create player: %v", err)
	}
	if err := player.EnterRoom(9); err != nil {
		t.Fatalf("enter room: %v", err)
	}
	players := playerlive.NewRegistry()
	if err := players.Add(player); err != nil {
		t.Fatalf("add player: %v", err)
	}
	bindings := binding.NewRegistry()
	if err := bindings.Add(binding.Binding{PlayerID: 7, ConnectionID: "conn", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("bind player: %v", err)
	}
	runtime := roomlive.NewRegistry(nil)
	active, err := runtime.Activate(roomlive.Snapshot{ID: 9, OwnerPlayerID: ownerID, MaxUsers: 25})
	if err != nil {
		t.Fatalf("activate room: %v", err)
	}
	roomGrid, err := grid.Parse("0000", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatalf("parse grid: %v", err)
	}
	if err := active.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Furniture: []worldfurniture.Item{item}, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}}); err != nil {
		t.Fatalf("load world: %v", err)
	}
	if _, err := runtime.Join(context.Background(), 9, roomlive.Occupant{PlayerID: 7, Username: "demo", ConnectionID: "conn", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("join room: %v", err)
	}
	connections := netconn.NewRegistry()
	connection, sent := interactionConnectionForTest(t, connections)
	local := bus.New()

	return Handler{
		Players: players, Bindings: bindings, Furniture: managerForTest{}, Runtime: runtime,
		Connections: connections, Events: local, Behaviors: interactions.NewRegistry(),
	}, connection, sent, active
}

// interactionConnectionForTest creates one connected session and captures its context.
func interactionConnectionForTest(t *testing.T, connections *netconn.Registry) (netconn.Context, *[]codec.Packet) {
	t.Helper()
	inbound, outbound := netconn.NewHandlerRegistry(), netconn.NewHandlerRegistry()
	var captured netconn.Context
	_ = inbound.Register(1, func(value netconn.Context, _ codec.Packet) error { captured = value; return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	sent := make([]codec.Packet, 0, 1)
	session, err := netconn.NewSession(netconn.SessionConfig{
		ID: "conn", Kind: "websocket", Inbound: inbound, Outbound: outbound,
		Sender:   func(_ context.Context, packet codec.Packet) error { sent = append(sent, packet); return nil },
		Disposer: func(context.Context, netconn.Reason) error { return nil },
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if err := connections.Register(session); err != nil {
		t.Fatalf("register session: %v", err)
	}
	if err := session.Receive(context.Background(), codec.Packet{Header: 1}); err != nil {
		t.Fatalf("capture context: %v", err)
	}

	return captured, &sent
}

// stateUpdaterForTest records durable state mutations.
type stateUpdaterForTest struct {
	// params stores the latest state mutation.
	params furnitureservice.StateParams
	// calls stores the number of mutations.
	calls int
	// err stores the configured mutation failure.
	err error
}

// UpdateState records a state mutation.
func (updater *stateUpdaterForTest) UpdateState(_ context.Context, params furnitureservice.StateParams) (furnituremodel.Item, error) {
	updater.params, updater.calls = params, updater.calls+1

	return furnituremodel.Item{}, updater.err
}

// teleporterForTest records delegated paired teleport starts.
type teleporterForTest struct {
	// request stores the latest teleport start.
	request teleport.StartRequest
	// err stores the configured start failure.
	err error
}

// Start records one paired teleport start.
func (starter *teleporterForTest) Start(_ context.Context, request teleport.StartRequest) error {
	starter.request = request

	return starter.err
}
