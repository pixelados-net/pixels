package definition

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies trigger code and conflicts trail the shared fields.
func TestEncode(t *testing.T) {
	packet, err := Encode(false, 5, nil, 10, 20, "", nil, 0, 7, []int32{64})
	if err != nil || packet.Header != Header {
		t.Fatalf("encode trigger: %v %#v", err, packet)
	}
	definition := codec.Definition{codec.BooleanField, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field}
	values, err := codec.DecodePacketExact(packet, definition)
	if err != nil || values[8].Int32 != 7 || values[10].Int32 != 64 {
		t.Fatalf("unexpected trigger %#v %v", values, err)
	}
}
