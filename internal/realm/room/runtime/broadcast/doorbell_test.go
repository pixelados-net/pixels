package broadcast

import (
	"context"
	"testing"
	"time"

	roomdoorbell "github.com/niflaot/pixels/internal/realm/room/access/doorbell"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outdoorbelldenied "github.com/niflaot/pixels/networking/outbound/room/doorbell/denied"
)

// TestDoorbellPublisherRejectsWaiterAndClosesOwnerPrompt verifies expiration delivery.
func TestDoorbellPublisherRejectsWaiterAndClosesOwnerPrompt(t *testing.T) {
	connections := netconn.NewRegistry()
	ownerPackets := registerConnectionForTest(t, connections, "owner")
	guestContext, guestPackets := doorbellContextForTest(t, "guest")
	active, err := roomlive.NewRoom(roomlive.Snapshot{ID: 9, OwnerPlayerID: 7, MaxUsers: 25})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	if _, err := active.Join(roomlive.Occupant{PlayerID: 7, Username: "Owner", ConnectionID: "owner", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("join owner: %v", err)
	}
	publisher := NewDoorbellPublisher(connections)
	err = publisher(context.Background(), active, []roomdoorbell.Expired{{
		Entry:  roomdoorbell.Entry{PlayerID: 8, Username: "Guest", Handler: guestContext, RequestedAt: time.Now()},
		Reason: roomdoorbell.ExpiredTimeout,
	}})
	if err != nil {
		t.Fatalf("publish expiration: %v", err)
	}
	if len(*ownerPackets) != 1 || (*ownerPackets)[0].Header != outdoorbelldenied.Header || len(*guestPackets) != 1 || (*guestPackets)[0].Header != outdoorbelldenied.Header {
		t.Fatalf("unexpected packets owner=%#v guest=%#v", *ownerPackets, *guestPackets)
	}
}

// TestDoorbellPublisherSkipsEmptyInput verifies no-op guards.
func TestDoorbellPublisherSkipsEmptyInput(t *testing.T) {
	if err := NewDoorbellPublisher(nil)(context.Background(), nil, nil); err != nil {
		t.Fatalf("unexpected no-op error %v", err)
	}
}

// doorbellContextForTest creates a backed waiting-player context.
func doorbellContextForTest(t *testing.T, id netconn.ID) (netconn.Context, *[]codec.Packet) {
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
	packets := make([]codec.Packet, 0, 1)
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

	return captured, &packets
}
