package membership

import (
	"context"
	"testing"

	groupconfig "github.com/niflaot/pixels/internal/realm/group/config"
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	groupruntime "github.com/niflaot/pixels/internal/realm/group/runtime"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inaccept "github.com/niflaot/pixels/networking/inbound/group/membership/accept"
	insetfavorite "github.com/niflaot/pixels/networking/inbound/group/membership/favorite/set"
	injoin "github.com/niflaot/pixels/networking/inbound/group/membership/join"
	insearch "github.com/niflaot/pixels/networking/inbound/group/membership/search"
	outinfo "github.com/niflaot/pixels/networking/outbound/group/identity/info"
	outfavorite "github.com/niflaot/pixels/networking/outbound/group/membership/favorite/update"
	outlist "github.com/niflaot/pixels/networking/outbound/group/membership/list"
	outsearch "github.com/niflaot/pixels/networking/outbound/group/membership/search"
	outprofilechanged "github.com/niflaot/pixels/networking/outbound/user/profile/changed"
)

// TestJoinRefreshesGroupListAndInformation verifies Nitro sees the committed membership immediately.
func TestJoinRefreshesGroupListAndInformation(t *testing.T) {
	store := &membershipStore{group: grouprecord.Group{ID: 3, Name: "Pixels"}, members: make(map[int64]grouprecord.Membership)}
	service := New(groupconfig.Config{}, store, nil, groupruntime.NewCache(), nil, nil, nil)
	bindings := binding.NewRegistry()
	if err := bindings.Add(binding.Binding{PlayerID: 7, ConnectionID: "group-join", ConnectionKind: "test"}); err != nil {
		t.Fatal(err)
	}
	packets := make([]codec.Packet, 0, 2)
	inbound, outbound := netconn.NewHandlerRegistry(), netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	handler := Handler{Membership: service, Delivery: groupruntime.NewDelivery(bindings, nil)}
	if err := inbound.Register(injoin.Header, handler.join, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated()); err != nil {
		t.Fatal(err)
	}
	session, err := netconn.NewSession(netconn.SessionConfig{
		ID: "group-join", Kind: "test", Inbound: inbound, Outbound: outbound,
		Sender:   func(_ context.Context, packet codec.Packet) error { packets = append(packets, packet); return nil },
		Disposer: func(context.Context, netconn.Reason) error { return nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	request, err := codec.NewPacket(injoin.Header, codec.Definition{codec.Int32Field}, codec.Int32(3))
	if err != nil {
		t.Fatal(err)
	}
	if err = session.Receive(context.Background(), request); err != nil {
		t.Fatal(err)
	}
	if len(packets) != 2 || packets[0].Header != outlist.Header || packets[1].Header != outinfo.Header {
		t.Fatalf("unexpected packets: %#v", packets)
	}
	values, err := codec.DecodePacketExact(packets[1], groupInformationDefinition)
	if err != nil || values[8].Int32 != 1 || values[9].Int32 != 1 {
		t.Fatalf("information=%#v err=%v", values, err)
	}
}

// TestFavoriteRefreshesOpenProfile verifies Nitro reloads the visible favorite star after persistence.
func TestFavoriteRefreshesOpenProfile(t *testing.T) {
	group := grouprecord.Group{ID: 3, Name: "Pixels"}
	store := &membershipStore{
		group:        group,
		members:      map[int64]grouprecord.Membership{7: {GroupID: group.ID, PlayerID: 7, Role: grouprecord.Member}},
		playerGroups: []grouprecord.PlayerGroup{{Group: group, Role: grouprecord.Member}},
	}
	service := New(groupconfig.Config{}, store, nil, groupruntime.NewCache(), nil, nil, nil)
	bindings := binding.NewRegistry()
	if err := bindings.Add(binding.Binding{PlayerID: 7, ConnectionID: "group-favorite", ConnectionKind: "test"}); err != nil {
		t.Fatal(err)
	}
	packets := make([]codec.Packet, 0, 2)
	inbound, outbound := netconn.NewHandlerRegistry(), netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	handler := Handler{Membership: service, Delivery: groupruntime.NewDelivery(bindings, nil)}
	if err := inbound.Register(insetfavorite.Header, handler.setFavorite, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated()); err != nil {
		t.Fatal(err)
	}
	session, err := netconn.NewSession(netconn.SessionConfig{
		ID: "group-favorite", Kind: "test", Inbound: inbound, Outbound: outbound,
		Sender:   func(_ context.Context, packet codec.Packet) error { packets = append(packets, packet); return nil },
		Disposer: func(context.Context, netconn.Reason) error { return nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	request, err := codec.NewPacket(insetfavorite.Header, codec.Definition{codec.Int32Field}, codec.Int32(int32(group.ID)))
	if err != nil {
		t.Fatal(err)
	}
	if err = session.Receive(context.Background(), request); err != nil {
		t.Fatal(err)
	}
	if store.favorite == nil || *store.favorite != group.ID {
		t.Fatalf("favorite=%v", store.favorite)
	}
	groups, err := service.PlayerGroups(context.Background(), 7)
	if err != nil || len(groups) != 1 || !groups[0].Favorite {
		t.Fatalf("groups=%#v err=%v", groups, err)
	}
	if len(packets) != 2 || packets[0].Header != outfavorite.Header || packets[1].Header != outprofilechanged.Header {
		t.Fatalf("unexpected packets: %#v", packets)
	}
}

// TestDefaultMemberListLevelOpensAllMembers verifies Nitro's omitted-filter sentinel remains compatible.
func TestDefaultMemberListLevelOpensAllMembers(t *testing.T) {
	group := grouprecord.Group{ID: 3, Name: "Pixels Requests"}
	store := &membershipStore{group: group, members: map[int64]grouprecord.Membership{1: {GroupID: group.ID, PlayerID: 1, Role: grouprecord.Admin}}}
	service := New(groupconfig.Config{}, store, nil, groupruntime.NewCache(), nil, nil, nil)
	bindings := binding.NewRegistry()
	if err := bindings.Add(binding.Binding{PlayerID: 1, ConnectionID: "group-members", ConnectionKind: "test"}); err != nil {
		t.Fatal(err)
	}
	packets := make([]codec.Packet, 0, 1)
	inbound, outbound := netconn.NewHandlerRegistry(), netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	handler := Handler{Membership: service, Delivery: groupruntime.NewDelivery(bindings, nil)}
	if err := inbound.Register(insearch.Header, handler.search, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated()); err != nil {
		t.Fatal(err)
	}
	session, err := netconn.NewSession(netconn.SessionConfig{
		ID: "group-members", Kind: "test", Inbound: inbound, Outbound: outbound,
		Sender:   func(_ context.Context, packet codec.Packet) error { packets = append(packets, packet); return nil },
		Disposer: func(context.Context, netconn.Reason) error { return nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	request, err := codec.NewPacket(insearch.Header, codec.Definition{codec.Int32Field, codec.Int32Field, codec.StringField, codec.Int32Field}, codec.Int32(3), codec.Int32(0), codec.String(""), codec.Int32(3))
	if err != nil {
		t.Fatal(err)
	}
	if err = session.Receive(context.Background(), request); err != nil {
		t.Fatal(err)
	}
	if store.memberPageLevel != 0 || len(packets) != 1 || packets[0].Header != outsearch.Header {
		t.Fatalf("level=%d packets=%#v", store.memberPageLevel, packets)
	}
}

// TestAcceptProjectsMembershipInformation verifies an online target refreshes without re-entering the room.
func TestAcceptProjectsMembershipInformation(t *testing.T) {
	group := grouprecord.Group{ID: 3, Name: "Pixels Requests", MemberCount: 3, PendingCount: 1}
	store := &membershipStore{
		group:   group,
		members: map[int64]grouprecord.Membership{1: {GroupID: group.ID, PlayerID: 1, Role: grouprecord.Admin}},
		pending: map[int64]bool{3: true},
	}
	service := New(groupconfig.Config{}, store, nil, groupruntime.NewCache(), nil, nil, nil)
	bindings := binding.NewRegistry()
	if err := bindings.Add(binding.Binding{PlayerID: 1, ConnectionID: "group-admin", ConnectionKind: "test"}); err != nil {
		t.Fatal(err)
	}
	if err := bindings.Add(binding.Binding{PlayerID: 3, ConnectionID: "group-target", ConnectionKind: "test"}); err != nil {
		t.Fatal(err)
	}
	targetPackets := make([]codec.Packet, 0, 1)
	targetInbound, targetOutbound := netconn.NewHandlerRegistry(), netconn.NewHandlerRegistry()
	targetOutbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	target, err := netconn.NewSession(netconn.SessionConfig{
		ID: "group-target", Kind: "test", Inbound: targetInbound, Outbound: targetOutbound,
		Sender: func(_ context.Context, packet codec.Packet) error {
			targetPackets = append(targetPackets, packet)
			return nil
		},
		Disposer: func(context.Context, netconn.Reason) error { return nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	connections := netconn.NewRegistry()
	if err = connections.Register(target); err != nil {
		t.Fatal(err)
	}
	handler := Handler{Membership: service, Delivery: groupruntime.NewDelivery(bindings, connections)}
	request, err := codec.NewPacket(inaccept.Header, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(3), codec.Int32(3))
	if err != nil {
		t.Fatal(err)
	}
	if err = handler.accept(netconn.Context{ConnectionID: "group-admin", ConnectionKind: "test"}, request); err != nil {
		t.Fatal(err)
	}
	if len(targetPackets) != 1 || targetPackets[0].Header != outinfo.Header {
		t.Fatalf("unexpected target packets: %#v", targetPackets)
	}
	values, err := codec.DecodePacketExact(targetPackets[0], groupInformationDefinition)
	if err != nil || values[8].Int32 != 1 || values[9].Int32 != 4 || values[17].Int32 != 0 {
		t.Fatalf("information=%#v err=%v", values, err)
	}
}

// groupInformationDefinition describes Nitro's group information projection for tests.
var groupInformationDefinition = codec.Definition{
	codec.Int32Field, codec.BooleanField, codec.Int32Field, codec.StringField, codec.StringField, codec.StringField,
	codec.Int32Field, codec.StringField, codec.Int32Field, codec.Int32Field, codec.BooleanField, codec.StringField,
	codec.BooleanField, codec.BooleanField, codec.StringField, codec.BooleanField, codec.BooleanField, codec.Int32Field,
}
