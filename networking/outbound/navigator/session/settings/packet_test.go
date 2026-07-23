package settings

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies NAVIGATOR_SETTINGS packet construction.
func TestEncode(t *testing.T) {
	packet, err := Encode(Params{WindowX: 1, WindowY: 2, WindowWidth: 425, WindowHeight: 592})
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if packet.Header != Header || values[2].Int32 != 425 {
		t.Fatalf("unexpected packet %#v values %#v", packet, values)
	}
}
