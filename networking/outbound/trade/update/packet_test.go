package update

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies an empty TRADE_UPDATE projection.
func TestEncode(t *testing.T) {
	packet, err := Encode(Participant{PlayerID: 1}, Participant{PlayerID: 2})
	definition := codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field}
	values, decodeErr := codec.DecodePacketExact(packet, definition)
	if err != nil || decodeErr != nil || packet.Header != Header || values[0].Int32 != 1 || values[4].Int32 != 2 {
		t.Fatalf("packet=%#v values=%#v err=%v decode=%v", packet, values, err, decodeErr)
	}
}
