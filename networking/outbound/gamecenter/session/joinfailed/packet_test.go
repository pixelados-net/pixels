package joinfailed

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies the renderer wire shape.
func TestEncode(t *testing.T) {
	packet, err := Encode(11)
	if err != nil {
		t.Fatal(err)
	}
	if packet.Header != Header {
		t.Fatalf("unexpected header: %d", packet.Header)
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field})
	if err != nil {
		t.Fatal(err)
	}
	if values[0].Int32 != 11 {
		t.Fatalf("unexpected field 0: %v", values[0].Int32)
	}
}
