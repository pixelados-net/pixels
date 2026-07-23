package room

import (
	"context"
	"testing"

	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	groupruntime "github.com/niflaot/pixels/internal/realm/group/runtime"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	ininfo "github.com/niflaot/pixels/networking/inbound/room/furniture/group/info"
	outcontext "github.com/niflaot/pixels/networking/outbound/room/furniture/group/context"
)

// TestInfoResolvesRoomAndGroupAuthoritatively verifies item-only requests never trust a client room id.
func TestInfoResolvesRoomAndGroupAuthoritatively(t *testing.T) {
	session, sent, cleanup := groupContextFixture(t)
	defer cleanup()

	packet, err := codec.NewPacket(ininfo.Header, codec.Definition{codec.Int32Field}, codec.Int32(910128))
	if err != nil {
		t.Fatal(err)
	}
	if err = session.Receive(context.Background(), packet); err != nil {
		t.Fatalf("request group context: %v", err)
	}
	if len(*sent) != 1 || (*sent)[0].Header != outcontext.Header {
		t.Fatalf("unexpected packets %#v", *sent)
	}
	values, err := codec.DecodePacketExact((*sent)[0], codec.Definition{
		codec.Int32Field, codec.Int32Field, codec.StringField, codec.Int32Field, codec.BooleanField, codec.BooleanField,
	})
	if err != nil || values[0].Int32 != 910128 || values[1].Int32 != 2 || values[2].String != "Pixels Open" || !values[4].Boolean {
		t.Fatalf("unexpected context values %#v error=%v", values, err)
	}
}

// TestInfoRejectsMismatchedProjectedGroup verifies an optional client group cannot retarget an item.
func TestInfoRejectsMismatchedProjectedGroup(t *testing.T) {
	session, sent, cleanup := groupContextFixture(t)
	defer cleanup()

	packet, err := codec.NewPacket(ininfo.Header, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(910128), codec.Int32(99))
	if err != nil {
		t.Fatal(err)
	}
	if err = session.Receive(context.Background(), packet); err != nil {
		t.Fatalf("request mismatched group context: %v", err)
	}
	if len(*sent) != 0 {
		t.Fatalf("expected no spoofed context response, got %#v", *sent)
	}
}

// groupContextFixture creates warmed room, group, binding, and connection state.
func groupContextFixture(t *testing.T) (*netconn.Session, *[]codec.Packet, func()) {
	t.Helper()
	rooms := roomlive.NewRegistry(nil)
	if _, err := rooms.Activate(roomlive.Snapshot{ID: 131, OwnerPlayerID: 1, MaxUsers: 25, RollerSpeed: -1}); err != nil {
		t.Fatal(err)
	}
	if _, err := rooms.Join(context.Background(), 131, roomlive.Occupant{PlayerID: 7, ConnectionID: "group-context", ConnectionKind: "test"}); err != nil {
		t.Fatal(err)
	}
	cache := groupruntime.NewCache()
	group := grouprecord.Group{ID: 2, Name: "Pixels Open", HomeRoomID: 131, ForumEnabled: true, ReadPolicy: grouprecord.Everyone}
	cache.PutRoom(131, groupruntime.GroupSnapshot{Group: group}, map[int64]grouprecord.Role{7: grouprecord.Member})
	cache.PutFurnitureLinks(2, []int64{910128})
	cache.PutPlayer(7, []grouprecord.PlayerGroup{{Group: group, Role: grouprecord.Member}})
	bindings := binding.NewRegistry()
	if err := bindings.Add(binding.Binding{PlayerID: 7, ConnectionID: "group-context", ConnectionKind: "test"}); err != nil {
		t.Fatal(err)
	}
	handler := Handler{Cache: cache, Delivery: groupruntime.NewDelivery(bindings, nil), Rooms: rooms}
	inbound := netconn.NewHandlerRegistry()
	if err := inbound.Register(ininfo.Header, handler.info, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated()); err != nil {
		t.Fatal(err)
	}
	sent := make([]codec.Packet, 0, 1)
	session, err := netconn.NewSession(netconn.SessionConfig{
		ID: "group-context", Kind: "test", Inbound: inbound, Outbound: permissiveOutbound(),
		Sender: func(_ context.Context, packet codec.Packet) error {
			sent = append(sent, packet)
			return nil
		},
		Disposer: func(context.Context, netconn.Reason) error { return nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	cleanup := func() { _, _, _ = rooms.Close(context.Background(), 131) }
	return session, &sent, cleanup
}

// permissiveOutbound creates a packet sink registry for handler tests.
func permissiveOutbound() *netconn.HandlerRegistry {
	registry := netconn.NewHandlerRegistry()
	registry.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	return registry
}
