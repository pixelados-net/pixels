package floorplan

import (
	"context"
	"testing"

	"github.com/niflaot/pixels/internal/command"
	domain "github.com/niflaot/pixels/internal/realm/room/control/floorplan"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outblocked "github.com/niflaot/pixels/networking/outbound/room/floorplan/blockedtiles"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// TestBlockedTilesHandleSendsFurnitureFootprints verifies occupied tile editor bootstrap.
func TestBlockedTilesHandleSendsFurnitureFootprints(t *testing.T) {
	players, bindings, player := floorplanActorForTest(t)
	if err := player.EnterRoom(9); err != nil {
		t.Fatalf("enter room: %v", err)
	}
	runtime := roomlive.NewRegistry(nil)
	active, err := runtime.Activate(roomlive.Snapshot{ID: 9, OwnerPlayerID: 7, MaxUsers: 5})
	if err != nil {
		t.Fatalf("activate room: %v", err)
	}
	roomGrid, _ := grid.Parse("00", grid.WithDoor(0, 0))
	item := worldfurniture.Item{ID: 5, Point: grid.MustPoint(0, 0), Definition: worldfurniture.Definition{Width: 2, Length: 1}}
	if err = active.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Furniture: []worldfurniture.Item{item}, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}}); err != nil {
		t.Fatalf("load world: %v", err)
	}
	connection, sent := floorplanConnectionContextForTest(t)
	room := roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}}, OwnerPlayerID: 7}
	authorizer := domain.NewAuthorizer(permissionsForTest{"own": true}, nil, domain.Nodes{OwnEdit: "own", AnyEdit: "any"})
	handler := BlockedTilesHandler{Players: players, Bindings: bindings, Rooms: roomsForTest{room: room}, Runtime: runtime, Authorize: authorizer}
	if (BlockedTilesCommand{}).CommandName() != BlockedTilesName {
		t.Fatalf("unexpected command name")
	}
	err = handler.Handle(context.Background(), command.Envelope[BlockedTilesCommand]{Command: BlockedTilesCommand{Handler: connection}})
	if err != nil {
		t.Fatalf("request blocked tiles: %v", err)
	}
	if len(*sent) != 1 || (*sent)[0].Header != outblocked.Header {
		t.Fatalf("unexpected packets %#v", *sent)
	}
}

// floorplanConnectionContextForTest creates one send-capable handler context.
func floorplanConnectionContextForTest(t *testing.T) (netconn.Context, *[]codec.Packet) {
	t.Helper()
	inbound := netconn.NewHandlerRegistry()
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	var connection netconn.Context
	inbound.SetFallback(func(value netconn.Context, _ codec.Packet) error {
		connection = value
		return nil
	}, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	sent := make([]codec.Packet, 0, 2)
	session, err := netconn.NewSession(netconn.SessionConfig{
		ID: "conn", Kind: "websocket", Inbound: inbound, Outbound: outbound,
		Sender: func(_ context.Context, packet codec.Packet) error {
			sent = append(sent, packet)
			return nil
		},
		Disposer: func(context.Context, netconn.Reason) error { return nil },
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if err = session.Receive(context.Background(), codec.Packet{Header: 1}); err != nil {
		t.Fatalf("capture connection context: %v", err)
	}

	return connection, &sent
}
