package playing

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies the renderer wire shape.
func TestEncode(t *testing.T) {
	packet, err := Encode(true)
	if err != nil {
		t.Fatal(err)
	}
	if packet.Header != Header {
		t.Fatalf("unexpected header: %d", packet.Header)
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.BooleanField})
	if err != nil {
		t.Fatal(err)
	}
	if values[0].Boolean != true {
		t.Fatalf("unexpected field 0: %v", values[0].Boolean)
	}
}
