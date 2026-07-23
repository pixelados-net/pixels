package create

import (
	"context"
	"testing"
	"time"

	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// managerForTest stores navigator room creation behavior.
type managerForTest struct {
	// rooms stores existing owner rooms.
	rooms []roommodel.Room
	// created stores the created room.
	created roommodel.Room
	// createErr stores a creation failure.
	createErr error
	// listErr stores an ownership read failure.
	listErr error
}

// Create returns configured creation behavior.
func (manager *managerForTest) Create(context.Context, roomservice.CreateParams) (roommodel.Room, error) {
	return manager.created, manager.createErr
}

// ListByOwner returns configured owner rooms.
func (manager *managerForTest) ListByOwner(context.Context, int64) ([]roommodel.Room, error) {
	return manager.rooms, manager.listErr
}

// commandFixture creates one bound player and packet-capturing connection.
func commandFixture(t *testing.T) (*playerlive.Registry, *binding.Registry, netconn.Context, *[]codec.Packet) {
	t.Helper()
	inbound := netconn.NewHandlerRegistry()
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	var connection netconn.Context
	if err := inbound.Register(1, func(ctx netconn.Context, _ codec.Packet) error { connection = ctx; return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated()); err != nil {
		t.Fatalf("register capture handler: %v", err)
	}
	packets := make([]codec.Packet, 0, 1)
	session, err := netconn.NewSession(netconn.SessionConfig{ID: "create", Kind: "websocket", Inbound: inbound, Outbound: outbound,
		Sender: func(_ context.Context, packet codec.Packet) error { packets = append(packets, packet); return nil }, Disposer: func(context.Context, netconn.Reason) error { return nil }})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if err := session.Receive(context.Background(), codec.Packet{Header: 1}); err != nil {
		t.Fatalf("capture context: %v", err)
	}
	peer, _ := playerlive.NewSessionPeer("create", "websocket", time.Now())
	player, _ := playerlive.NewPlayer(playerlive.Snapshot{ID: 7, Username: "demo"}, peer)
	players := playerlive.NewRegistry()
	_ = players.Add(player)
	bindings := binding.NewRegistry()
	_ = bindings.Add(binding.Binding{PlayerID: 7, ConnectionID: "create", ConnectionKind: "websocket"})

	return players, bindings, connection, &packets
}

// createdRoomForTest creates one durable room fixture.
func createdRoomForTest() roommodel.Room {
	return roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 12}}, Name: "Created Room"}
}
