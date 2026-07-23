package settings

import (
	"context"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/command"
	"github.com/niflaot/pixels/internal/permission"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	muteallchanged "github.com/niflaot/pixels/internal/realm/room/control/events/muteallchanged"
	roomsettings "github.com/niflaot/pixels/internal/realm/room/control/settings"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// muteRoomFinderForTest returns one room.
type muteRoomFinderForTest struct {
	// room stores the returned room.
	room roommodel.Room
}

// FindByID finds one room.
func (finder muteRoomFinderForTest) FindByID(context.Context, int64) (roommodel.Room, bool, error) {
	return finder.room, true, nil
}

// mutePermissionsForTest allows every requested node.
type mutePermissionsForTest struct{}

// HasPermission allows every requested node.
func (mutePermissionsForTest) HasPermission(context.Context, int64, permission.Node) (bool, error) {
	return true, nil
}

// muteEventsForTest captures one event.
type muteEventsForTest struct {
	// events stores published events.
	events []bus.Event
}

// Publish captures one event.
func (publisher *muteEventsForTest) Publish(_ context.Context, event bus.Event) error {
	publisher.events = append(publisher.events, event)
	return nil
}

// TestHandleTogglesBroadcastsAndPublishes verifies complete mute-all command behavior.
func TestHandleTogglesBroadcastsAndPublishes(t *testing.T) {
	players, bindings, player := actorForTest(t)
	if err := player.EnterRoom(9); err != nil {
		t.Fatalf("enter room: %v", err)
	}
	runtime := roomlive.NewRegistry(nil)
	active, err := runtime.Activate(roomlive.Snapshot{ID: 9, OwnerPlayerID: 7, MaxUsers: 10})
	if err != nil {
		t.Fatalf("activate room: %v", err)
	}
	if _, err = runtime.Join(context.Background(), 9, roomlive.Occupant{PlayerID: 7, Username: "demo", ConnectionID: "conn", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("join room: %v", err)
	}
	connections := netconn.NewRegistry()
	sent := registerConnectionForTest(t, connections)
	events := &muteEventsForTest{}
	room := roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}}, OwnerPlayerID: 7}
	authorizer := roomsettings.New(mutePermissionsForTest{}, roomsettings.Nodes{OwnManage: "own", AnyManage: "any"})
	handler := MuteAllHandler{Players: players, Bindings: bindings, Rooms: muteRoomFinderForTest{room: room}, Authorize: authorizer, Runtime: runtime, Connections: connections, Events: events}
	err = handler.Handle(context.Background(), command.Envelope[MuteAllCommand]{Command: MuteAllCommand{Handler: netconn.Context{ConnectionID: "conn", ConnectionKind: "websocket"}}})
	if err != nil {
		t.Fatalf("toggle mute-all: %v", err)
	}
	if !active.MuteAll() || len(*sent) != 1 || (*sent)[0].Header != 2533 {
		t.Fatalf("muted=%v packets=%#v", active.MuteAll(), *sent)
	}
	if len(events.events) != 1 || events.events[0].Name != muteallchanged.Name {
		t.Fatalf("unexpected events %#v", events.events)
	}
}

// actorForTest creates one room actor and binding.
func actorForTest(t *testing.T) (*playerlive.Registry, *binding.Registry, *playerlive.Player) {
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

// registerConnectionForTest registers one outbound packet capture.
func registerConnectionForTest(t *testing.T, connections *netconn.Registry) *[]codec.Packet {
	t.Helper()
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	sent := make([]codec.Packet, 0, 1)
	session, err := netconn.NewSession(netconn.SessionConfig{ID: "conn", Kind: "websocket", Outbound: outbound,
		Sender:   func(_ context.Context, packet codec.Packet) error { sent = append(sent, packet); return nil },
		Disposer: func(context.Context, netconn.Reason) error { return nil }})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if err = connections.Register(session); err != nil {
		t.Fatalf("register session: %v", err)
	}

	return &sent
}
