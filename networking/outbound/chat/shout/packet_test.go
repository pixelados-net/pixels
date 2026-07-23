package shout

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies the UNIT_CHAT_SHOUT wire shape.
func TestEncode(t *testing.T) {
	packet, err := Encode(9, "hey", 0, 2, 3)
	if err != nil || packet.Header != Header {
		t.Fatalf("unexpected packet=%#v err=%v", packet, err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || values[0].Int32 != 9 || values[1].String != "hey" || values[3].Int32 != 2 {
		t.Fatalf("unexpected values=%#v err=%v", values, err)
	}
}
