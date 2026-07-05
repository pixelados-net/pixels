package forward

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies ROOM_FORWARD packet construction.
func TestEncode(t *testing.T) {
	packet, err := Encode(7)
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if values[0].Int32 != 7 {
		t.Fatalf("unexpected values %#v", values)
	}
}
