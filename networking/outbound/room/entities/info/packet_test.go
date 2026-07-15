package info

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies complete room-unit profile fields.
func TestEncode(t *testing.T) {
	definition := codec.Definition{codec.Int32Field, codec.StringField, codec.StringField, codec.StringField, codec.Int32Field}
	packet, err := Encode(9, "hd-1-1", "M", "hello", 12)
	values, decodeErr := codec.DecodePacketExact(packet, definition)
	if err != nil || decodeErr != nil || packet.Header != Header || values[0].Int32 != 9 || values[1].String != "hd-1-1" || values[4].Int32 != 12 {
		t.Fatalf("unexpected packet %#v values=%#v err=%v/%v", packet, values, err, decodeErr)
	}
}
