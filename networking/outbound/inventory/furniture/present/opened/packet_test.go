package opened

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies PRESENT_OPENED encoding.
func TestEncode(t *testing.T) {
	packet, err := Encode(39, "chair", 41, true)
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
	if values[0].String != FloorItemType || values[1].Int32 != 39 || values[2].String != "chair" {
		t.Fatalf("unexpected product fields %#v", values)
	}
	if values[3].Int32 != 41 || values[4].String != FloorItemType || !values[5].Boolean || values[6].String != "" {
		t.Fatalf("unexpected placement fields %#v", values)
	}
}
