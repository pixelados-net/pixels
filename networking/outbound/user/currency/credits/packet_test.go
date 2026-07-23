package credits

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeWritesCreditsString verifies the legacy decimal string shape.
func TestEncodeWritesCreditsString(t *testing.T) {
	packet, err := Encode(125)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if packet.Header != Header || values[0].String != "125.0" {
		t.Fatalf("unexpected packet %#v values=%#v", packet, values)
	}
}
