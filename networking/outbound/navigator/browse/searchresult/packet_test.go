package searchresult

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/outbound/navigator/browse/roomcard"
)

// TestEncode verifies NAVIGATOR_SEARCH packet construction.
func TestEncode(t *testing.T) {
	packet, err := Encode("hotel_view", "", []ResultList{{
		Code:  "popular",
		Mode:  1,
		Rooms: []roomcard.Card{{RoomID: 1, RoomName: "Demo", OwnerName: "demo", MaxUserCount: 25}},
	}})
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}
	values, rest, err := codec.DecodePacket(packet, Definition)
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if packet.Header != Header || values[2].Int32 != 1 || len(rest) == 0 {
		t.Fatalf("unexpected packet %#v values %#v", packet, values)
	}
}
