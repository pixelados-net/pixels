package favorite

import (
	"context"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/permission"
	"github.com/niflaot/pixels/internal/realm/navigator/core"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomrights "github.com/niflaot/pixels/internal/realm/room/control/rights"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inadd "github.com/niflaot/pixels/networking/inbound/navigator/favorite/add"
	inremove "github.com/niflaot/pixels/networking/inbound/navigator/favorite/remove"
	outchanged "github.com/niflaot/pixels/networking/outbound/navigator/favorite/changed"
	outlist "github.com/niflaot/pixels/networking/outbound/navigator/favorite/list"
	outalert "github.com/niflaot/pixels/networking/outbound/session/alert"
	"github.com/niflaot/pixels/pkg/bus"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// favoriteNavigator stores deterministic favorite state.
type favoriteNavigator struct {
	core.Manager
	ids     []int64
	added   int
	removed int
}

// AddFavorite records one favorite insertion.
func (navigator *favoriteNavigator) AddFavorite(_ context.Context, _ int64, roomID int64, _ int32, _ bool) error {
	navigator.added++
	navigator.ids = append(navigator.ids, roomID)
	return nil
}

// RemoveFavorite records one favorite removal.
func (navigator *favoriteNavigator) RemoveFavorite(_ context.Context, _ int64, roomID int64) error {
	navigator.removed++
	for index, value := range navigator.ids {
		if value == roomID {
			navigator.ids = append(navigator.ids[:index], navigator.ids[index+1:]...)
			break
		}
	}
	return nil
}

// ListFavoriteRoomIDs returns current favorite state.
func (navigator *favoriteNavigator) ListFavoriteRoomIDs(context.Context, int64) ([]int64, error) {
	return navigator.ids, nil
}

// favoriteRooms returns one configured room.
type favoriteRooms struct {
	roomservice.Manager
	room roommodel.Room
}

// FindByID returns the configured room when identifiers match.
func (rooms favoriteRooms) FindByID(_ context.Context, id int64) (roommodel.Room, bool, error) {
	return rooms.room, rooms.room.ID == id, nil
}

// favoritePermissions returns one quota bypass decision.
type favoritePermissions bool

// HasPermission returns the configured quota bypass decision.
func (allowed favoritePermissions) HasPermission(context.Context, int64, permission.Node) (bool, error) {
	return bool(allowed), nil
}

// favoriteRights returns one room visibility decision.
type favoriteRights struct {
	roomrights.Manager
	allowed bool
}

// HasRights returns the configured room visibility decision.
func (rights favoriteRights) HasRights(context.Context, int64, int64) (bool, error) {
	return rights.allowed, nil
}

// TestLimitDefaultsAndConfigures verifies the wire and persistence quota source.
func TestLimitDefaultsAndConfigures(t *testing.T) {
	if actual := (Handler{}).limit(); actual != 30 {
		t.Fatalf("default limit=%d", actual)
	}
	if actual := (Handler{Limit: 42}).limit(); actual != 42 {
		t.Fatalf("configured limit=%d", actual)
	}
	if (Handler{}).hasRights(1, 1) {
		t.Fatal("expected no rights without a rights manager")
	}
}

// TestRegisterHandlersInstallsFavoritePackets verifies both protocol adapters.
func TestRegisterHandlersInstallsFavoritePackets(t *testing.T) {
	RegisterHandlers(nil, Handler{})
	registry := netconn.NewHandlerRegistry()
	RegisterHandlers(registry, Handler{})
	if registry.Len() != 2 {
		t.Fatalf("handlers=%d", registry.Len())
	}
}

// TestFavoriteAdaptersRejectMalformedAndUnboundRequests verifies decode and authentication gates.
func TestFavoriteAdaptersRejectMalformedAndUnboundRequests(t *testing.T) {
	handler := Handler{}
	if err := handler.add(netconn.Context{}, codec.Packet{Header: inremove.Header}); err == nil {
		t.Fatal("expected malformed add rejection")
	}
	if err := handler.remove(netconn.Context{}, codec.Packet{Header: inadd.Header}); err == nil {
		t.Fatal("expected malformed remove rejection")
	}
	if err := handler.change(netconn.Context{}, 9, true); err == nil {
		t.Fatal("expected unbound request rejection")
	}
}

// TestFavoriteMutationFlows verifies add, duplicate, limit, privacy, and remove behavior.
func TestFavoriteMutationFlows(t *testing.T) {
	tests := []struct {
		name          string
		initial       []int64
		room          roommodel.Room
		packet        codec.Packet
		limit         int32
		unlimited     bool
		rights        bool
		expectedFirst uint16
		added         int
		removed       int
	}{
		{name: "add", room: favoriteRoom(9, roommodel.DoorModeOpen), packet: favoriteAddPacket(t, 9), expectedFirst: outchanged.Header, added: 1},
		{name: "duplicate", initial: []int64{9}, room: favoriteRoom(9, roommodel.DoorModeOpen), packet: favoriteAddPacket(t, 9), expectedFirst: outchanged.Header},
		{name: "limit", initial: []int64{8}, room: favoriteRoom(9, roommodel.DoorModeOpen), packet: favoriteAddPacket(t, 9), limit: 1, expectedFirst: outalert.Header},
		{name: "unlimited", initial: []int64{8}, room: favoriteRoom(9, roommodel.DoorModeOpen), packet: favoriteAddPacket(t, 9), limit: 1, unlimited: true, expectedFirst: outchanged.Header, added: 1},
		{name: "invisible_rights", room: favoriteRoom(9, roommodel.DoorModeInvisible), packet: favoriteAddPacket(t, 9), rights: true, expectedFirst: outchanged.Header, added: 1},
		{name: "invisible", room: favoriteRoom(9, roommodel.DoorModeInvisible), packet: favoriteAddPacket(t, 9), expectedFirst: outalert.Header},
		{name: "remove", initial: []int64{9}, packet: favoriteRemovePacket(t, 9), expectedFirst: outchanged.Header, removed: 1},
	}
	for _, current := range tests {
		t.Run(current.name, func(t *testing.T) {
			navigator := &favoriteNavigator{ids: append([]int64(nil), current.initial...)}
			handler, connection, packets := favoriteFixture(t, navigator, current.room, current.limit, current.unlimited, current.rights)
			if current.packet.Header == inadd.Header {
				if err := handler.add(connection, current.packet); err != nil {
					t.Fatalf("add: %v", err)
				}
			} else if err := handler.remove(connection, current.packet); err != nil {
				t.Fatalf("remove: %v", err)
			}
			if len(*packets) == 0 || (*packets)[0].Header != current.expectedFirst {
				t.Fatalf("packets=%#v", *packets)
			}
			if navigator.added != current.added || navigator.removed != current.removed {
				t.Fatalf("added=%d removed=%d", navigator.added, navigator.removed)
			}
			if current.expectedFirst == outchanged.Header && len(*packets) != 2 || len(*packets) == 2 && (*packets)[1].Header != outlist.Header {
				t.Fatalf("refresh packets=%#v", *packets)
			}
		})
	}
}

// favoriteFixture creates one authenticated packet-capturing favorite handler.
func favoriteFixture(t *testing.T, navigator *favoriteNavigator, room roommodel.Room, limit int32, unlimited bool, rights bool) (Handler, netconn.Context, *[]codec.Packet) {
	t.Helper()
	inbound, outbound := netconn.NewHandlerRegistry(), netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	var connection netconn.Context
	_ = inbound.Register(1, func(ctx netconn.Context, _ codec.Packet) error { connection = ctx; return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	packets := make([]codec.Packet, 0, 2)
	session, _ := netconn.NewSession(netconn.SessionConfig{ID: "favorite", Kind: "websocket", Inbound: inbound, Outbound: outbound, Sender: func(_ context.Context, packet codec.Packet) error { packets = append(packets, packet); return nil }, Disposer: func(context.Context, netconn.Reason) error { return nil }})
	_ = session.Receive(context.Background(), codec.Packet{Header: 1})
	peer, _ := playerlive.NewSessionPeer("favorite", "websocket", time.Now())
	player, _ := playerlive.NewPlayer(playerlive.Snapshot{ID: 7, Username: "demo"}, peer)
	players := playerlive.NewRegistry()
	_ = players.Add(player)
	bindings := binding.NewRegistry()
	_ = bindings.Add(binding.Binding{PlayerID: 7, ConnectionID: "favorite", ConnectionKind: "websocket"})
	return Handler{Navigator: navigator, Players: players, Bindings: bindings, Rooms: favoriteRooms{room: room}, Rights: favoriteRights{allowed: rights}, Permissions: favoritePermissions(unlimited), Events: bus.New(), Limit: limit}, connection, &packets
}

// favoriteRoom creates one room fixture.
func favoriteRoom(id int64, mode roommodel.DoorMode) roommodel.Room {
	return roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: id}}, OwnerPlayerID: 8, DoorMode: mode}
}

// favoriteAddPacket creates one add request.
func favoriteAddPacket(t *testing.T, roomID int32) codec.Packet {
	t.Helper()
	packet, err := codec.NewPacket(inadd.Header, inadd.Definition, codec.Int32(roomID))
	if err != nil {
		t.Fatal(err)
	}
	return packet
}

// favoriteRemovePacket creates one remove request.
func favoriteRemovePacket(t *testing.T, roomID int32) codec.Packet {
	t.Helper()
	packet, err := codec.NewPacket(inremove.Header, inremove.Definition, codec.Int32(roomID))
	if err != nil {
		t.Fatal(err)
	}
	return packet
}
