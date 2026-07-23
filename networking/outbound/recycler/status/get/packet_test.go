package get

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies status and reserved timeout ordering.
func TestEncode(t *testing.T) {
	packet, err := Encode(Enabled, 0)
	values, decodeErr := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field, codec.Int32Field})
	if err != nil || decodeErr != nil || values[0].Int32 != Enabled || values[1].Int32 != 0 {
		t.Fatalf("unexpected values %#v error=%v/%v", values, err, decodeErr)
	}
}
