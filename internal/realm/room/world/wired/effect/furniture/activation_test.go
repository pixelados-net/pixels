package furniture

import (
	"context"
	"testing"
	"time"

	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outstate "github.com/niflaot/pixels/networking/outbound/room/furniture/state"
	outupdate "github.com/niflaot/pixels/networking/outbound/room/furniture/update"
)

// packetConnection records packets broadcast to one room occupant.
type packetConnection struct {
	netconn.Connection
	// packets stores packets in delivery order.
	packets []codec.Packet
}

// ID returns the stable test connection identifier.
func (*packetConnection) ID() netconn.ID { return "wired-connection" }

// Kind returns the stable test transport kind.
func (*packetConnection) Kind() netconn.Kind { return "websocket" }

// Send records one outbound packet.
func (connection *packetConnection) Send(_ context.Context, packet codec.Packet) error {
	connection.packets = append(connection.packets, packet)
	return nil
}

// TestActivateUsesCompactStatePacket verifies animation cannot rewrite placement in Nitro.
func TestActivateUsesCompactStatePacket(t *testing.T) {
	rooms, manager, active := furnitureRoom(t)
	connections := netconn.NewRegistry()
	connection := &packetConnection{}
	if err := connections.Register(connection); err != nil {
		t.Fatal(err)
	}
	if _, err := active.Join(roomlive.Occupant{
		PlayerID: 7, Username: "demo", ConnectionID: connection.ID(), ConnectionKind: connection.Kind(),
	}); err != nil {
		t.Fatal(err)
	}
	service := New(rooms, manager, connections)
	if err := service.Activate(context.Background(), active.ID(), 10); err != nil {
		t.Fatal(err)
	}
	if len(connection.packets) != 1 || connection.packets[0].Header != outstate.Header || connection.packets[0].Header == outupdate.Header {
		t.Fatalf("activation packets=%#v", connection.packets)
	}
	values, err := codec.DecodePacketExact(connection.packets[0], outstate.Definition)
	if err != nil || values[0].Int32 != 10 || values[1].Int32 != 1 {
		t.Fatalf("activation values=%#v err=%v", values, err)
	}
	item, found := active.FurnitureItem(10)
	if !found || item.ExtraData != "1" || item.Point != grid.MustPoint(1, 0) || manager.item.ExtraData != "0" {
		t.Fatalf("activation item=%+v durable=%q found=%t", item, manager.item.ExtraData, found)
	}
	if err := service.Activate(context.Background(), active.ID(), 10); err != nil {
		t.Fatal(err)
	}
	if len(connection.packets) != 1 {
		t.Fatalf("repeated activation inverted state packets=%#v", connection.packets)
	}
	active.RunScheduled(time.Now().Add(time.Second))
	if len(connection.packets) != 2 {
		t.Fatalf("activation reset packets=%#v", connection.packets)
	}
	values, err = codec.DecodePacketExact(connection.packets[1], outstate.Definition)
	if err != nil || values[0].Int32 != 10 || values[1].Int32 != 0 {
		t.Fatalf("activation reset values=%#v err=%v", values, err)
	}
	item, found = active.FurnitureItem(10)
	if !found || item.ExtraData != "0" || manager.item.ExtraData != "0" {
		t.Fatalf("reset activation item=%+v durable=%q found=%t", item, manager.item.ExtraData, found)
	}
}
