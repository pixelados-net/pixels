package move

import (
	"context"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/command"
	"github.com/niflaot/pixels/internal/permission"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outupdate "github.com/niflaot/pixels/networking/outbound/room/furniture/update"
	outbubble "github.com/niflaot/pixels/networking/outbound/session/bubblealert"
)

// globalPermissionForTest grants one global furniture permission.
type globalPermissionForTest bool

// HasPermission returns the configured global permission result.
func (allowed globalPermissionForTest) HasPermission(context.Context, int64, permission.Node) (bool, error) {
	return bool(allowed), nil
}

// TestHandleRejectsGuestWithoutFurnitureRights verifies visitors cannot move furniture.
func TestHandleRejectsGuestWithoutFurnitureRights(t *testing.T) {
	handler, connections := handlerForTest(t)
	handler.Players, handler.Bindings = guestPresenceForTest(t)
	connection, sent := registeredConnectionForTest(t, connections, "conn")
	handler.Furniture = &fakeManager{
		definition: chairDefinitionForTest(), definitionFound: true,
		item: placedItemForTest(), itemFound: true,
	}

	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: connection, ItemID: 1, X: 3, Y: 0, Rotation: 0},
	})
	if err != nil {
		t.Fatalf("reject guest move: %v", err)
	}
	if len(*sent) != 2 || (*sent)[0].Header != outupdate.Header || (*sent)[1].Header != outbubble.Header {
		t.Fatalf("expected authoritative rollback then no-rights bubble, got %#v", *sent)
	}
}

// TestHandleAllowsRightsHolderToMoveForeignFurniture verifies room rights authorize movement independently of item ownership.
func TestHandleAllowsRightsHolderToMoveForeignFurniture(t *testing.T) {
	handler, connections := handlerForTest(t)
	handler.Players, handler.Bindings = guestPresenceForTest(t)
	connection, sent := registeredConnectionForTest(t, connections, "conn")
	room, _ := handler.Runtime.Find(9)
	room.GrantRights(8)
	handler.Furniture = &fakeManager{
		definition: chairDefinitionForTest(), definitionFound: true,
		item: placedItemForTest(), itemFound: true,
	}

	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: connection, ItemID: 1, X: 3, Y: 0, Rotation: 2},
	})
	if err != nil {
		t.Fatalf("move foreign furniture with rights: %v", err)
	}
	if len(*sent) == 0 || (*sent)[0].Header != outupdate.Header {
		t.Fatalf("expected authoritative move update, got %#v", *sent)
	}
}

// TestHandleAllowsGlobalFurnitureManager verifies staff authority does not require persisted room rights.
func TestHandleAllowsGlobalFurnitureManager(t *testing.T) {
	handler, connections := handlerForTest(t)
	handler.Players, handler.Bindings = guestPresenceForTest(t)
	handler.Permissions = globalPermissionForTest(true)
	connection, sent := registeredConnectionForTest(t, connections, "conn")
	handler.Furniture = &fakeManager{
		definition: chairDefinitionForTest(), definitionFound: true,
		item: placedItemForTest(), itemFound: true,
	}

	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: connection, ItemID: 1, X: 3, Y: 0, Rotation: 2},
	})
	if err != nil {
		t.Fatalf("move furniture with global authority: %v", err)
	}
	if len(*sent) == 0 || (*sent)[0].Header != outupdate.Header {
		t.Fatalf("expected authoritative move update, got %#v", *sent)
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
