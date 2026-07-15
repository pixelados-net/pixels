package handitemreceived

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies private hand-item receipt fields.
func TestEncode(t *testing.T) {
	packet, err := Encode(9, 27)
	values, decodeErr := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field, codec.Int32Field})
	if err != nil || decodeErr != nil || packet.Header != Header || values[0].Int32 != 9 || values[1].Int32 != 27 {
		t.Fatalf("unexpected packet %#v values=%#v err=%v/%v", packet, values, err, decodeErr)
	}
}
