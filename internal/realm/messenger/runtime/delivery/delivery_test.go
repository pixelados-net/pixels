package delivery

import (
	"context"
	"testing"

	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// TestSendUsesExistingAuthenticatedBinding verifies delivery without a parallel registry.
func TestSendUsesExistingAuthenticatedBinding(t *testing.T) {
	bindings := binding.NewRegistry()
	connections := netconn.NewRegistry()
	var sent codec.Packet
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	session, err := netconn.NewSession(netconn.SessionConfig{ID: "messenger", Kind: "websocket", Outbound: outbound,
		Sender: func(_ context.Context, packet codec.Packet) error { sent = packet; return nil }, Disposer: func(context.Context, netconn.Reason) error { return nil }})
	if err != nil {
		t.Fatal(err)
	}
	if err = connections.Register(session); err != nil {
		t.Fatal(err)
	}
	if err = bindings.Add(binding.Binding{PlayerID: 7, ConnectionID: session.ID(), ConnectionKind: session.Kind()}); err != nil {
		t.Fatal(err)
	}
	sender := New(bindings, connections)
	delivered, err := sender.Send(context.Background(), 7, codec.Packet{Header: 42})
	if err != nil || !delivered || sent.Header != 42 || !sender.Online(7) || sender.Online(8) {
		t.Fatalf("unexpected delivered=%v packet=%#v err=%v", delivered, sent, err)
	}
}
