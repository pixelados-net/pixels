package wallremove

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies removed wall-item fields.
func TestEncode(t *testing.T) {
	packet, err := Encode(42, 7)
	values, decodeErr := codec.DecodePacketExact(packet, codec.Definition{codec.StringField, codec.Int32Field})
	if err != nil || decodeErr != nil || packet.Header != Header || values[0].String != "42" || values[1].Int32 != 7 {
		t.Fatalf("unexpected packet %#v values=%#v err=%v/%v", packet, values, err, decodeErr)
	}
}
