package postitopen

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies post-it editor-open fields.
func TestEncode(t *testing.T) {
	packet, err := Encode(42, ":w=2,3 l=4,5 r")
	values, decodeErr := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field, codec.StringField})
	if err != nil || decodeErr != nil || packet.Header != Header || values[0].Int32 != 42 || values[1].String != ":w=2,3 l=4,5 r" {
		t.Fatalf("unexpected packet %#v values=%#v err=%v/%v", packet, values, err, decodeErr)
	}
}
