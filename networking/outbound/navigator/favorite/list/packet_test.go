package favorites

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies USER_FAVORITE_ROOM_COUNT packet construction.
func TestEncode(t *testing.T) {
	packet, err := Encode(30, []int32{1, 2})
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}
	values, rest, err := codec.DecodePacket(packet, Definition)
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if packet.Header != Header || values[1].Int32 != 2 || len(rest) == 0 {
		t.Fatalf("unexpected packet %#v values %#v", packet, values)
	}
}
