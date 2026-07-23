package hint

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies match count and exact flag ordering.
func TestEncode(t *testing.T) {
	packet, err := Encode(3, true)
	values, decodeErr := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field, codec.BooleanField})
	if err != nil || decodeErr != nil || values[0].Int32 != 3 || !values[1].Boolean {
		t.Fatalf("unexpected values %#v error=%v/%v", values, err, decodeErr)
	}
}
