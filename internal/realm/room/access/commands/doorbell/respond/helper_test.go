package respond

import (
	"context"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// commandContext creates a backed handler context and captured session.
func commandContext(t *testing.T, id netconn.ID) (netconn.Context, *netconn.Session, *[]codec.Packet) {
	t.Helper()
	var captured netconn.Context
	inbound := netconn.NewHandlerRegistry()
	if err := inbound.Register(1, func(handler netconn.Context, _ codec.Packet) error {
		captured = handler
		return nil
	}, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated()); err != nil {
		t.Fatalf("register inbound: %v", err)
	}
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	packets := make([]codec.Packet, 0, 2)
	session, err := netconn.NewSession(netconn.SessionConfig{
		ID: id, Kind: "websocket", Inbound: inbound, Outbound: outbound,
		Sender: func(_ context.Context, packet codec.Packet) error {
			packets = append(packets, packet)
			return nil
		},
		Disposer: func(context.Context, netconn.Reason) error { return nil },
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if err := session.Receive(context.Background(), codec.Packet{Header: 1}); err != nil {
		t.Fatalf("capture context: %v", err)
	}

	return captured, session, &packets
}

// lastHeader returns the last captured packet header.
func lastHeader(packets *[]codec.Packet) uint16 {
	if packets == nil || len(*packets) == 0 {
		return 0
	}

	return (*packets)[len(*packets)-1].Header
}
