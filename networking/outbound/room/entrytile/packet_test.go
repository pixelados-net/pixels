package entrytile

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies ROOM_MODEL_DOOR packet construction.
func TestEncode(t *testing.T) {
	packet, err := Encode(1, 2, 3)
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if packet.Header != Header || values[0].Int32 != 1 || values[1].Int32 != 2 || values[2].Int32 != 3 {
		t.Fatalf("unexpected packet %#v values %#v", packet, values)
	}
}
