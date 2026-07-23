package model

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies ROOM_MODEL packet construction.
func TestEncode(t *testing.T) {
	packet, err := Encode(true, 0, "xxx\rx0x\rxxx")
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if packet.Header != Header || !values[0].Boolean || values[2].String == "" {
		t.Fatalf("unexpected packet %#v values %#v", packet, values)
	}
}
