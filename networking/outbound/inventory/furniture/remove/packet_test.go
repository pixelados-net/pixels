package remove

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies REMOVE_FURNITURE_FROM_INVENTORY encoding.
func TestEncode(t *testing.T) {
	packet, err := Encode(9)
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
	if values[0].Int32 != 9 {
		t.Fatalf("unexpected item id %#v", values[0])
	}
}
