package pickup

import (
	"context"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outbubble "github.com/niflaot/pixels/networking/outbound/session/bubblealert"
)

// TestHandleRejectsGuestWithoutFurnitureRights verifies visitors cannot pick up furniture.
func TestHandleRejectsGuestWithoutFurnitureRights(t *testing.T) {
	handler, connections := handlerForTest(t)
	handler.Players, handler.Bindings = guestPresenceForTest(t)
	connection, sent := registeredConnectionForTest(t, connections, "conn")

	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: connection, ItemID: 1},
	})
	if err != nil {
		t.Fatalf("reject guest pickup: %v", err)
	}
	if len(*sent) != 1 || (*sent)[0].Header != outbubble.Header {
		t.Fatalf("expected one no-rights bubble, got %#v", *sent)
	}
}

// guestPresenceForTest creates a visitor bound to the room owned by player seven.
func guestPresenceForTest(t *testing.T) (*playerlive.Registry, *binding.Registry) {
	t.Helper()

	peer, err := playerlive.NewSessionPeer(netconn.ID("conn"), netconn.Kind("websocket"), time.Now())
	if err != nil {
		t.Fatalf("create guest peer: %v", err)
	}
	player, err := playerlive.NewPlayer(playerlive.Snapshot{ID: 8, Username: "guest"}, peer)
	if err != nil {
		t.Fatalf("create guest player: %v", err)
	}
	if err := player.EnterRoom(9); err != nil {
		t.Fatalf("enter guest room: %v", err)
	}
	players := playerlive.NewRegistry()
	if err := players.Add(player); err != nil {
		t.Fatalf("add guest player: %v", err)
	}
	bindings := binding.NewRegistry()
	if err := bindings.Add(binding.Binding{PlayerID: 8, ConnectionID: "conn", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("bind guest player: %v", err)
	}

	return players, bindings
}
