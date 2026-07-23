package opened

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies WIRED_OPEN encoding.
func TestEncode(t *testing.T) {
	packet, err := Encode(44)
	if err != nil || packet.Header != Header {
		t.Fatalf("encode opened: %v %#v", err, packet)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || values[0].Int32 != 44 {
		t.Fatalf("unexpected opened %#v %v", values, err)
	}
}
