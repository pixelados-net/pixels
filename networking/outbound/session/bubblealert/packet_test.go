package bubblealert

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies BUBBLE_ALERT encoding.
func TestEncode(t *testing.T) {
	packet, err := Encode("furni_placement_error", "cant_stack")
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
	if values[0].String != "furni_placement_error" || values[1].Int32 != 1 ||
		values[2].String != "message" || values[3].String != "cant_stack" {
		t.Fatalf("unexpected values %#v", values)
	}
}
