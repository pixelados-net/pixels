package postitplaced

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies post-it placement confirmation fields.
func TestEncode(t *testing.T) {
	packet, err := Encode(42, 3)
	values, decodeErr := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field, codec.Int32Field})
	if err != nil || decodeErr != nil || packet.Header != Header || values[0].Int32 != 42 || values[1].Int32 != 3 {
		t.Fatalf("unexpected packet %#v values=%#v err=%v/%v", packet, values, err, decodeErr)
	}
}
