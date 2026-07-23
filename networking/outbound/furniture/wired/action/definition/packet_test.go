package definition

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies action code, delay, and conflicts.
func TestEncode(t *testing.T) {
	packet, err := Encode(false, 5, nil, 10, 20, "", nil, 0, 17, 3, []int32{63})
	if err != nil || packet.Header != Header {
		t.Fatalf("encode action: %v %#v", err, packet)
	}
	definition := codec.Definition{codec.BooleanField, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field}
	values, err := codec.DecodePacketExact(packet, definition)
	if err != nil || values[8].Int32 != 17 || values[9].Int32 != 3 || values[11].Int32 != 63 {
		t.Fatalf("unexpected action %#v %v", values, err)
	}
}
