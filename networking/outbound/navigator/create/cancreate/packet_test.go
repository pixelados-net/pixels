package cancreate

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies CAN_CREATE_ROOM packet construction.
func TestEncode(t *testing.T) {
	packet, err := Encode(0, 50)
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if packet.Header != Header || values[1].Int32 != 50 {
		t.Fatalf("unexpected packet %#v values %#v", packet, values)
	}
}
