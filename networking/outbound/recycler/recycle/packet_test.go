package recycle

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies completion and real prize identifiers.
func TestEncode(t *testing.T) {
	packet, err := Encode(Complete, 123)
	values, decodeErr := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field, codec.Int32Field})
	if err != nil || decodeErr != nil || values[0].Int32 != Complete || values[1].Int32 != 123 {
		t.Fatalf("unexpected values %#v error=%v/%v", values, err, decodeErr)
	}
}
