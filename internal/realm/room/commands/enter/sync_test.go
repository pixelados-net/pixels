package enter

import (
	"context"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomlive "github.com/niflaot/pixels/internal/realm/room/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// TestHandleBroadcastsSecondPlayerToFirst verifies multi-user room synchronization.
func TestHandleBroadcastsSecondPlayerToFirst(t *testing.T) {
	connections := netconn.NewRegistry()
	firstConnection, firstSent := sessionConnectionWithIDForSyncTest(t, connections, "first")
	secondConnection, secondSent := sessionConnectionWithIDForSyncTest(t, connections, "second")
	players := playerlive.NewRegistry()
	firstPlayer := livePlayerForSyncTest(t, 7, "first", "first")
	secondPlayer := livePlayerForSyncTest(t, 8, "second", "second")
	addPlayerForSyncTest(t, players, firstPlayer)
	addPlayerForSyncTest(t, players, secondPlayer)
	bindings := binding.NewRegistry()
	addBindingForSyncTest(t, bindings, 7, "first")
	addBindingForSyncTest(t, bindings, 8, "second")
	handler := Handler{
		Players: players, Bindings: bindings, Rooms: roomManagerForTest{room: roomForTest(), found: true},
		Layouts: layoutManagerForTest{roomLayout: layoutForTest(), found: true}, Runtime: roomlive.NewRegistry(nil),
		Connections: connections,
	}

	if err := handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{Handler: firstConnection, RoomID: 9}}); err != nil {
		t.Fatalf("handle first enter: %v", err)
	}
	firstCount := len(*firstSent)
	if err := handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{Handler: secondConnection, RoomID: 9}}); err != nil {
		t.Fatalf("handle second enter: %v", err)
	}

	if len(*firstSent) != firstCount+2 {
		t.Fatalf("expected first player to receive second spawn, got %#v", *firstSent)
	}
	if len(*secondSent) != 8 {
		t.Fatalf("expected second player room snapshot, got %#v", *secondSent)
	}
	if (*secondSent)[7].Header != 780 {
		t.Fatalf("expected room rights level bootstrap, got %#v", (*secondSent)[7:])
	}
}

// sessionConnectionWithIDForSyncTest creates a registered connection context.
func sessionConnectionWithIDForSyncTest(t *testing.T, connections *netconn.Registry, id netconn.ID) (netconn.Context, *[]codec.Packet) {
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

	sent := make([]codec.Packet, 0, 8)
	session, err := netconn.NewSession(netconn.SessionConfig{
		ID:       id,
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
	if err := connections.Register(session); err != nil {
		t.Fatalf("register session: %v", err)
	}
	if err := session.Receive(context.Background(), codec.Packet{Header: 1}); err != nil {
		t.Fatalf("receive context packet: %v", err)
	}

	return captured, &sent
}

// livePlayerForSyncTest creates a live player for multi-session tests.
func livePlayerForSyncTest(t *testing.T, playerID int64, username string, connectionID netconn.ID) *playerlive.Player {
	t.Helper()

	peer, err := playerlive.NewSessionPeer(connectionID, netconn.Kind("websocket"), time.Now())
	if err != nil {
		t.Fatalf("create peer: %v", err)
	}
	player, err := playerlive.NewPlayer(playerlive.Snapshot{ID: playerID, Username: username}, peer)
	if err != nil {
		t.Fatalf("create player: %v", err)
	}

	return player
}

// addPlayerForSyncTest adds a player to the live registry.
func addPlayerForSyncTest(t *testing.T, players *playerlive.Registry, player *playerlive.Player) {
	t.Helper()

	if err := players.Add(player); err != nil {
		t.Fatalf("add player: %v", err)
	}
}

// addBindingForSyncTest adds a player connection binding.
func addBindingForSyncTest(t *testing.T, bindings *binding.Registry, playerID int64, connectionID netconn.ID) {
	t.Helper()

	if err := bindings.Add(binding.Binding{PlayerID: playerID, ConnectionID: connectionID, ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("bind player: %v", err)
	}
}
