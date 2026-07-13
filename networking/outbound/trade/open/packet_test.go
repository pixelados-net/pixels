package open

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies TRADE_OPEN encoding.
func TestEncode(t *testing.T) {
	packet, err := Encode(1, true, 2, true)
	values, decodeErr := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field})
	if err != nil || decodeErr != nil || packet.Header != Header || values[1].Int32 != 1 || values[3].Int32 != 1 {
		t.Fatalf("packet=%#v values=%#v err=%v decode=%v", packet, values, err, decodeErr)
	}
}
