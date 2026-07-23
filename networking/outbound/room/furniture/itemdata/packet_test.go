package itemdata

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies editable wall-item fields.
func TestEncode(t *testing.T) {
	packet, err := Encode(42, "FFFF33 hello")
	values, decodeErr := codec.DecodePacketExact(packet, codec.Definition{codec.StringField, codec.StringField})
	if err != nil || decodeErr != nil || packet.Header != Header || values[0].String != "42" || values[1].String != "FFFF33 hello" {
		t.Fatalf("unexpected packet %#v values=%#v err=%v/%v", packet, values, err, decodeErr)
	}
}
