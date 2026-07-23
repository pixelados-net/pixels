package definition

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies conditions contain no unconsumed padding.
func TestEncode(t *testing.T) {
	packet, err := Encode(false, 5, nil, 10, 20, "", nil, 0, 24)
	if err != nil || packet.Header != Header {
		t.Fatalf("encode condition: %v %#v", err, packet)
	}
	definition := codec.Definition{codec.BooleanField, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field}
	values, err := codec.DecodePacketExact(packet, definition)
	if err != nil || values[8].Int32 != 24 {
		t.Fatalf("unexpected condition %#v %v", values, err)
	}
}
