package talk

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies the UNIT_CHAT wire shape.
func TestEncode(t *testing.T) {
	packet, err := Encode(7, "hello", 1, 3, 5)
	if err != nil || packet.Header != Header {
		t.Fatalf("unexpected packet=%#v err=%v", packet, err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || values[0].Int32 != 7 || values[1].String != "hello" || values[4].Int32 != 0 || values[5].Int32 != 5 {
		t.Fatalf("unexpected values=%#v err=%v", values, err)
	}
}
