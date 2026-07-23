package whisper

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies the UNIT_CHAT_WHISPER wire shape.
func TestEncode(t *testing.T) {
	packet, err := Encode(4, "secret", 0, 1, 6)
	if err != nil || packet.Header != Header {
		t.Fatalf("unexpected packet=%#v err=%v", packet, err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || values[0].Int32 != 4 || values[1].String != "secret" || values[5].Int32 != 6 {
		t.Fatalf("unexpected values=%#v err=%v", values, err)
	}
}
