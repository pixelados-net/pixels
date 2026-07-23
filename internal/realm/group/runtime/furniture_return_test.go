package runtime

import (
	"context"
	"errors"
	"testing"
	"time"

	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outrefresh "github.com/niflaot/pixels/networking/outbound/inventory/furniture/refresh"
	outunseen "github.com/niflaot/pixels/networking/outbound/inventory/unseen"
	outremove "github.com/niflaot/pixels/networking/outbound/room/furniture/remove"
	outwallremove "github.com/niflaot/pixels/networking/outbound/room/furniture/wallremove"
	outheightmap "github.com/niflaot/pixels/networking/outbound/room/heightmapupdate"
)

// TestReturnFurnitureClearsActiveWorldAndRefreshesOwner verifies membership removal cannot leave
// invisible furniture, stale collision, or a stale inventory.
func TestReturnFurnitureClearsActiveWorldAndRefreshesOwner(t *testing.T) {
	ctx := context.Background()
	rooms := roomlive.NewRegistry(nil)
	active, err := rooms.Activate(roomlive.Snapshot{ID: 9, MaxUsers: 2})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _, _ = rooms.Close(ctx, 9) })
	roomGrid, err := grid.Parse("000")
	if err != nil {
		t.Fatal(err)
	}
	floor := worldfurniture.Item{
		ID: 39, OwnerPlayerID: 3, Point: grid.MustPoint(1, 0),
		Definition: worldfurniture.Definition{Width: 1, Length: 1, StackHeight: grid.HeightFromInt(1)},
	}
	if err = active.LoadWorld(roomlive.WorldConfig{
		Grid: roomGrid, Furniture: []worldfurniture.Item{floor},
		Door: worldpath.Position{Point: grid.MustPoint(0, 0)},
	}); err != nil {
		t.Fatal(err)
	}

	connections := netconn.NewRegistry()
	bobPackets := make([]codec.Packet, 0, 5)
	guestPackets := make([]codec.Packet, 0, 3)
	bob := returnSessionForTest(t, "bob", &bobPackets)
	guest := returnSessionForTest(t, "guest", &guestPackets)
	if err = connections.Register(bob); err != nil {
		t.Fatal(err)
	}
	if err = connections.Register(guest); err != nil {
		t.Fatal(err)
	}
	if _, err = rooms.Join(ctx, 9, returnedOccupantForTest(3, bob.ID())); err != nil {
		t.Fatal(err)
	}
	if _, err = rooms.Join(ctx, 9, returnedOccupantForTest(4, guest.ID())); err != nil {
		t.Fatal(err)
	}
	if _, err = active.MoveTo(3, grid.MustPoint(1, 0)); !errors.Is(err, worldpath.ErrNoPath) {
		t.Fatalf("expected occupied tile before return, got %v", err)
	}

	bindings := binding.NewRegistry()
	if err = bindings.Add(binding.Binding{PlayerID: 3, ConnectionID: bob.ID(), ConnectionKind: bob.Kind()}); err != nil {
		t.Fatal(err)
	}
	projector := NewProjector(NewCache(), rooms, connections, NewDelivery(bindings, connections))
	returned := grouprecord.FurnitureReturn{RoomID: 9, Items: []grouprecord.ReturnedFurniture{
		{ItemID: 39, OwnerPlayerID: 3},
		{ItemID: 40, OwnerPlayerID: 3, Wall: true},
	}}
	if err = projector.ReturnFurniture(ctx, returned); err != nil {
		t.Fatal(err)
	}

	if _, found := active.FurnitureItem(39); found {
		t.Fatal("returned floor furniture remained in the active world")
	}
	if _, err = active.MoveTo(3, grid.MustPoint(1, 0)); err != nil {
		t.Fatalf("returned furniture tile remained blocked: %v", err)
	}
	assertReturnHeaders(t, bobPackets, outremove.Header, outwallremove.Header, outheightmap.Header, outunseen.Header, outrefresh.Header)
	assertReturnHeaders(t, guestPackets, outremove.Header, outwallremove.Header, outheightmap.Header)
	if hasReturnHeader(guestPackets, outunseen.Header) || hasReturnHeader(guestPackets, outrefresh.Header) {
		t.Fatal("bystander received another player's inventory refresh")
	}
}

// returnSessionForTest creates one packet-recording transport session.
func returnSessionForTest(t *testing.T, id netconn.ID, packets *[]codec.Packet) *netconn.Session {
	t.Helper()
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(
		func(netconn.Context, codec.Packet) error { return nil },
		netconn.AllowAnyActiveState(),
		netconn.AllowUnauthenticated(),
	)
	session, err := netconn.NewSession(netconn.SessionConfig{
		ID: id, Kind: "websocket", StartedAt: time.Now(),
		Outbound: outbound,
		Sender: func(_ context.Context, packet codec.Packet) error {
			*packets = append(*packets, packet)
			return nil
		},
		Disposer: func(context.Context, netconn.Reason) error { return nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	return session
}

// returnedOccupantForTest creates one active-room occupant for return projections.
func returnedOccupantForTest(playerID int64, connectionID netconn.ID) roomlive.Occupant {
	return roomlive.Occupant{PlayerID: playerID, Username: "tester", ConnectionID: connectionID, ConnectionKind: "websocket"}
}

// assertReturnHeaders verifies all required packet families were projected.
func assertReturnHeaders(t *testing.T, packets []codec.Packet, headers ...uint16) {
	t.Helper()
	for _, header := range headers {
		if !hasReturnHeader(packets, header) {
			t.Fatalf("missing header %d in %#v", header, packets)
		}
	}
}

// hasReturnHeader reports whether a packet list contains one header.
func hasReturnHeader(packets []codec.Packet, header uint16) bool {
	for _, packet := range packets {
		if packet.Header == header {
			return true
		}
	}
	return false
}
