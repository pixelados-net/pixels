package wardrobe

import (
	"context"
	"testing"

	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inredeem "github.com/niflaot/pixels/networking/inbound/user/clothing/redeem"
	inget "github.com/niflaot/pixels/networking/inbound/user/wardrobe/get"
	insave "github.com/niflaot/pixels/networking/inbound/user/wardrobe/save"
	outremove "github.com/niflaot/pixels/networking/outbound/inventory/furniture/remove"
	outalert "github.com/niflaot/pixels/networking/outbound/session/alert"
	outclothing "github.com/niflaot/pixels/networking/outbound/user/clothing/sets"
	outoutfits "github.com/niflaot/pixels/networking/outbound/user/wardrobe/outfits"
)

// TestWardrobeHandlersSaveGetAndRedeem verifies all wardrobe protocol adapters.
func TestWardrobeHandlersSaveGetAndRedeem(t *testing.T) {
	store := &wardrobeStore{}
	handler, connection, packets := wardrobeFixture(t, New(store))
	savePacket, _ := codec.NewPacket(insave.Header, insave.Definition, codec.Int32(1), codec.String("hd-180-1"), codec.String("M"))
	if err := handler.save(connection, savePacket); err != nil || store.outfit.SlotID != 1 {
		t.Fatalf("outfit=%#v err=%v", store.outfit, err)
	}
	getPacket, _ := codec.NewPacket(inget.Header, inget.Definition, codec.Int32(4))
	if err := handler.get(connection, getPacket); err != nil || wardrobeLastHeader(*packets) != outoutfits.Header {
		t.Fatalf("get packets=%#v err=%v", *packets, err)
	}
	redeemPacket, _ := codec.NewPacket(inredeem.Header, inredeem.Definition, codec.Int32(99))
	if err := handler.redeem(connection, redeemPacket); err != nil || len(*packets) != 3 || (*packets)[1].Header != outremove.Header || (*packets)[2].Header != outclothing.Header {
		t.Fatalf("redeem packets=%#v err=%v", *packets, err)
	}
}

// TestWardrobeHandlersRejectInvalidAndMalformedRequests verifies validation feedback and decode gates.
func TestWardrobeHandlersRejectInvalidAndMalformedRequests(t *testing.T) {
	handler, connection, packets := wardrobeFixture(t, New(&wardrobeStore{}))
	invalid, _ := codec.NewPacket(insave.Header, insave.Definition, codec.Int32(0), codec.String("bad"), codec.String("X"))
	if err := handler.save(connection, invalid); err != nil || wardrobeLastHeader(*packets) != outalert.Header {
		t.Fatalf("packets=%#v err=%v", *packets, err)
	}
	for _, call := range []func() error{
		func() error { return handler.save(connection, codec.Packet{Header: inget.Header}) },
		func() error { return handler.get(connection, codec.Packet{Header: insave.Header}) },
		func() error { return handler.redeem(connection, codec.Packet{Header: insave.Header}) },
	} {
		if err := call(); err == nil {
			t.Fatal("expected malformed packet rejection")
		}
	}
}

// TestWardrobeRegistrationAndMissingBinding verifies registration and authentication gates.
func TestWardrobeRegistrationAndMissingBinding(t *testing.T) {
	RegisterHandlers(nil, Handler{})
	registry := netconn.NewHandlerRegistry()
	RegisterHandlers(registry, Handler{})
	if registry.Len() != 3 {
		t.Fatalf("handlers=%d", registry.Len())
	}
	if _, found := (Handler{}).playerID(netconn.Context{}); found {
		t.Fatal("expected missing binding")
	}
}

// wardrobeFixture creates one bound packet-capturing wardrobe handler.
func wardrobeFixture(t *testing.T, service *Service) (Handler, netconn.Context, *[]codec.Packet) {
	t.Helper()
	inbound, outbound := netconn.NewHandlerRegistry(), netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	var connection netconn.Context
	_ = inbound.Register(1, func(ctx netconn.Context, _ codec.Packet) error { connection = ctx; return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	packets := make([]codec.Packet, 0, 3)
	session, _ := netconn.NewSession(netconn.SessionConfig{ID: "wardrobe", Kind: "websocket", Inbound: inbound, Outbound: outbound, Sender: func(_ context.Context, packet codec.Packet) error { packets = append(packets, packet); return nil }, Disposer: func(context.Context, netconn.Reason) error { return nil }})
	_ = session.Receive(context.Background(), codec.Packet{Header: 1})
	bindings := binding.NewRegistry()
	_ = bindings.Add(binding.Binding{PlayerID: 7, ConnectionID: "wardrobe", ConnectionKind: "websocket"})
	return Handler{Service: service, Bindings: bindings}, connection, &packets
}

// wardrobeLastHeader returns the latest captured packet header.
func wardrobeLastHeader(packets []codec.Packet) uint16 {
	if len(packets) == 0 {
		return 0
	}
	return packets[len(packets)-1].Header
}
