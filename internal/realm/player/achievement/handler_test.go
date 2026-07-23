package achievement

import (
	"context"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	incurrent "github.com/niflaot/pixels/networking/inbound/user/badge/current"
	inequip "github.com/niflaot/pixels/networking/inbound/user/badge/equip"
	inlist "github.com/niflaot/pixels/networking/inbound/user/badge/list"
	outcurrent "github.com/niflaot/pixels/networking/outbound/user/badge/current"
	outlist "github.com/niflaot/pixels/networking/outbound/user/badge/list"
)

// TestHandlersProjectInventoryCurrentAndSelection verifies the native badge flow.
func TestHandlersProjectInventoryCurrentAndSelection(t *testing.T) {
	store := &achievementStore{badges: []Badge{{ID: 1, Code: "ADM", Equipped: true, Slot: 1}, {ID: 2, Code: "HC1"}}}
	service := New(store)
	bindings := binding.NewRegistry()
	if err := bindings.Add(binding.Binding{PlayerID: 7, ConnectionID: "badges", ConnectionKind: "websocket"}); err != nil {
		t.Fatal(err)
	}
	sent := make([]codec.Packet, 0, 4)
	session := badgeSession(t, Handler{Achievements: service, Bindings: bindings}, &sent)
	if err := session.Receive(context.Background(), codec.Packet{Header: inlist.Header}); err != nil {
		t.Fatal(err)
	}
	current, err := codec.NewPacket(incurrent.Header, incurrent.Definition, codec.Int32(7))
	if err != nil {
		t.Fatal(err)
	}
	if err = session.Receive(context.Background(), current); err != nil {
		t.Fatal(err)
	}
	equip, err := codec.NewPacket(inequip.Header, inequip.Definition,
		codec.Int32(1), codec.String("HC1"), codec.Int32(2), codec.String(""),
		codec.Int32(3), codec.String(""), codec.Int32(4), codec.String(""),
		codec.Int32(5), codec.String(""))
	if err != nil {
		t.Fatal(err)
	}
	if err = session.Receive(context.Background(), equip); err != nil {
		t.Fatal(err)
	}
	want := []uint16{outlist.Header, outcurrent.Header, outlist.Header, outcurrent.Header}
	if len(sent) != len(want) {
		t.Fatalf("sent=%v", sent)
	}
	for index, header := range want {
		if sent[index].Header != header {
			t.Fatalf("packet %d header=%d want=%d", index, sent[index].Header, header)
		}
	}
	if len(store.equipped) != 1 || store.equipped[0] != "HC1" {
		t.Fatalf("equipped=%v", store.equipped)
	}
}

// TestHandlersIgnoreMissingBindings verifies stale authenticated adapters close safely.
func TestHandlersIgnoreMissingBindings(t *testing.T) {
	handler := Handler{}
	if err := handler.inventory(netconn.Context{}, codec.Packet{Header: inlist.Header}); err != nil {
		t.Fatal(err)
	}
	RegisterHandlers(nil, handler)
}

// badgeSession creates one connected authenticated badge protocol session.
func badgeSession(t *testing.T, handler Handler, sent *[]codec.Packet) *netconn.Session {
	t.Helper()
	inbound := netconn.NewHandlerRegistry()
	RegisterHandlers(inbound, handler)
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	session, err := netconn.NewSession(netconn.SessionConfig{
		ID: "badges", Kind: "websocket", Inbound: inbound, Outbound: outbound,
		Sender: func(_ context.Context, packet codec.Packet) error {
			*sent = append(*sent, packet)
			return nil
		},
		Disposer: func(context.Context, netconn.Reason) error { return nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = session.Transition(netconn.EventPacketReceived); err != nil {
		t.Fatal(err)
	}
	if err = session.Transition(netconn.EventAuthenticationStarted); err != nil {
		t.Fatal(err)
	}
	if err = session.Authenticate(time.Now()); err != nil {
		t.Fatal(err)
	}
	if err = session.Transition(netconn.EventSessionReady); err != nil {
		t.Fatal(err)
	}
	return session
}
