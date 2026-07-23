package handitem

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies the Nitro hand item field order.
func TestEncode(t *testing.T) {
	packet, err := Encode(9, 27)
	if err != nil {
		t.Fatalf("encode hand item: %v", err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		t.Fatalf("decode hand item: %v", err)
	}
	if packet.Header != Header || values[0].Int32 != 9 || values[1].Int32 != 27 {
		t.Fatalf("unexpected packet %#v %#v", packet, values)
	}
}
