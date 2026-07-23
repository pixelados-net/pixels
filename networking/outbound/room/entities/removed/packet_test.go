package removed

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies UNIT_REMOVE packet encoding.
func TestEncode(t *testing.T) {
	packet, err := Encode(9)
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if packet.Header != Header || values[0].String != "9" {
		t.Fatalf("unexpected packet %#v values=%#v", packet, values)
	}
}
