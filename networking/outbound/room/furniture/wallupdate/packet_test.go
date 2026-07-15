package wallupdate

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies wall-item update field order.
func TestEncode(t *testing.T) {
	definition := codec.Definition{codec.StringField, codec.Int32Field, codec.StringField, codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field}
	packet, err := Encode(42, 9, ":w=2,3 l=4,5 r", "0", 1, 7)
	values, decodeErr := codec.DecodePacketExact(packet, definition)
	if err != nil || decodeErr != nil || packet.Header != Header || values[0].String != "42" || values[2].String != ":w=2,3 l=4,5 r" || values[6].Int32 != 7 {
		t.Fatalf("unexpected packet %#v values=%#v err=%v/%v", packet, values, err, decodeErr)
	}
}
