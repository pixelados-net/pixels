package paint

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies room plane appearance fields.
func TestEncode(t *testing.T) {
	packet, err := Encode("landscape", "3.1")
	values, decodeErr := codec.DecodePacketExact(packet, codec.Definition{codec.StringField, codec.StringField})
	if err != nil || decodeErr != nil || packet.Header != Header || values[0].String != "landscape" || values[1].String != "3.1" {
		t.Fatalf("unexpected packet %#v values=%#v err=%v/%v", packet, values, err, decodeErr)
	}
}
