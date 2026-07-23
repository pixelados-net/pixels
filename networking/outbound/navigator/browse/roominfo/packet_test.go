package roominfo

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/outbound/navigator/browse/roomcard"
)

// TestEncode verifies ROOM_INFO packet construction.
func TestEncode(t *testing.T) {
	packet, err := Encode(Params{
		RoomEnter: true,
		Room:      roomcard.Card{RoomID: 1, RoomName: "Demo", OwnerName: "demo", MaxUserCount: 25},
		Chat:      ChatSettings{Distance: 50},
	})
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}
	values, rest, err := codec.DecodePacket(packet, Definition)
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if packet.Header != Header || !values[0].Boolean || len(rest) == 0 {
		t.Fatalf("unexpected packet %#v values %#v", packet, values)
	}
}
