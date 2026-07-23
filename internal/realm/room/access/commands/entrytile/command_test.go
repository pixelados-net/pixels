package entrytile

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	"github.com/niflaot/pixels/internal/realm/room/world/layout"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outentrytile "github.com/niflaot/pixels/networking/outbound/room/entrytile"
	outthickness "github.com/niflaot/pixels/networking/outbound/room/thickness/updated"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// TestHandleSendsEntryTile verifies entry tile command handling.
func TestHandleSendsEntryTile(t *testing.T) {
	connection, sent := sessionConnectionForTest(t)
	player := playerForTest(t)
	if (Command{}).CommandName() != Name {
		t.Fatalf("unexpected command name %s", (Command{}).CommandName())
	}
	if err := player.EnterRoom(9); err != nil {
		t.Fatalf("enter room: %v", err)
	}
	handler := Handler{
		Players:  playerRegistryForTest(t, player),
		Bindings: bindingRegistryForTest(t),
		Rooms:    roomManagerForTest{room: roomForTest(), found: true},
		Layouts:  layoutManagerForTest{roomLayout: layoutForTest(), found: true},
	}

	err := handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{Handler: connection}})
	if err != nil {
		t.Fatalf("handle entry tile: %v", err)
	}
	if len(*sent) != 2 || (*sent)[0].Header != outentrytile.Header || (*sent)[1].Header != outthickness.Header {
		t.Fatalf("unexpected packets %#v", *sent)
	}
	values, err := codec.DecodePacketExact((*sent)[0], outentrytile.Definition)
	if err != nil {
		t.Fatalf("decode entry tile: %v", err)
	}
	if values[0].Int32 != 1 || values[1].Int32 != 1 || values[2].Int32 != 2 {
		t.Fatalf("unexpected entry tile values %#v", values)
	}
}

// TestHandleRejectsMissingRoom verifies room presence validation.
func TestHandleRejectsMissingRoom(t *testing.T) {
	connection, _ := sessionConnectionForTest(t)
	player := playerForTest(t)
	handler := Handler{Players: playerRegistryForTest(t, player), Bindings: bindingRegistryForTest(t)}

	err := handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{Handler: connection}})
	if !errors.Is(err, roomservice.ErrRoomNotFound) {
		t.Fatalf("expected room not found, got %v", err)
	}
}

// TestHandleLookupFailures verifies room and layout lookup failures.
func TestHandleLookupFailures(t *testing.T) {
	cases := []struct {
		// name identifies the failure case.
		name string
		// rooms stores the room manager fixture.
		rooms roomManagerForTest
		// layouts stores the layout manager fixture.
		layouts layoutManagerForTest
		// expected stores the expected error.
		expected error
	}{
		{name: "room error", rooms: roomManagerForTest{err: errLookupFailed}, expected: errLookupFailed},
		{name: "room missing", rooms: roomManagerForTest{}, expected: roomservice.ErrRoomNotFound},
		{name: "layout error", rooms: roomManagerForTest{room: roomForTest(), found: true}, layouts: layoutManagerForTest{err: errLookupFailed}, expected: errLookupFailed},
		{name: "layout missing", rooms: roomManagerForTest{room: roomForTest(), found: true}, layouts: layoutManagerForTest{}, expected: roomservice.ErrLayoutNotAvailable},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			connection, _ := sessionConnectionForTest(t)
			player := playerForTest(t)
			if err := player.EnterRoom(9); err != nil {
				t.Fatalf("enter room: %v", err)
			}
			handler := Handler{
				Players:  playerRegistryForTest(t, player),
				Bindings: bindingRegistryForTest(t),
				Rooms:    testCase.rooms,
				Layouts:  testCase.layouts,
			}

			err := handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{Handler: connection}})
			if !errors.Is(err, testCase.expected) {
				t.Fatalf("expected %v, got %v", testCase.expected, err)
			}
		})
	}
}

// playerForTest creates a live player.
func playerForTest(t *testing.T) *playerlive.Player {
	t.Helper()

	peer, err := playerlive.NewSessionPeer("conn", "websocket", time.Now())
	if err != nil {
		t.Fatalf("create peer: %v", err)
	}
	player, err := playerlive.NewPlayer(playerlive.Snapshot{ID: 7, Username: "demo"}, peer)
	if err != nil {
		t.Fatalf("create player: %v", err)
	}

	return player
}

// playerRegistryForTest creates a player registry.
func playerRegistryForTest(t *testing.T, player *playerlive.Player) *playerlive.Registry {
	t.Helper()

	players := playerlive.NewRegistry()
	if err := players.Add(player); err != nil {
		t.Fatalf("add player: %v", err)
	}

	return players
}

// bindingRegistryForTest creates a player binding registry.
func bindingRegistryForTest(t *testing.T) *binding.Registry {
	t.Helper()

	bindings := binding.NewRegistry()
	err := bindings.Add(binding.Binding{PlayerID: 7, ConnectionID: "conn", ConnectionKind: "websocket"})
	if err != nil {
		t.Fatalf("add binding: %v", err)
	}

	return bindings
}

// sessionConnectionForTest creates a captured connection context.
func sessionConnectionForTest(t *testing.T) (netconn.Context, *[]codec.Packet) {
	t.Helper()

	inbound := netconn.NewHandlerRegistry()
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error {
		return nil
	}, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	var captured netconn.Context
	if err := inbound.Register(1, func(context netconn.Context, _ codec.Packet) error {
		captured = context
		return nil
	}, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated()); err != nil {
		t.Fatalf("register inbound: %v", err)
	}
	sent := make([]codec.Packet, 0, 1)
	session, err := netconn.NewSession(netconn.SessionConfig{
		ID:       "conn",
		Kind:     "websocket",
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

// roomForTest creates a room model fixture.
func roomForTest() roommodel.Room {
	return roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}}, ModelName: "model_a"}
}

// layoutForTest creates a layout fixture.
func layoutForTest() layout.Layout {
	return layout.Layout{Name: "model_a", Heightmap: "0", TileSize: 1, DoorX: 1, DoorY: 1, DoorZ: 0, DoorDirection: 2, Enabled: true}
}
