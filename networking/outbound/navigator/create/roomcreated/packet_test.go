package roomcreated

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies ROOM_CREATED packet construction.
func TestEncode(t *testing.T) {
	packet, err := Encode(8, "Demo")
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if packet.Header != Header || values[0].Int32 != 8 {
		t.Fatalf("unexpected packet %#v values %#v", packet, values)
	}
}
