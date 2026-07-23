package remove

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies REMOVE_FLOOR_ITEM encoding.
func TestEncode(t *testing.T) {
	packet, err := Encode(9, 7)
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}
	if packet.Header != Header {
		t.Fatalf("unexpected header %d", packet.Header)
	}

	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if values[0].String != "9" || values[1].Boolean != false || values[2].Int32 != 7 || values[3].Int32 != 0 {
		t.Fatalf("unexpected values %#v", values)
	}
}
