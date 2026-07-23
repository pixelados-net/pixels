package flood

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies FLOOD_CONTROL encoding.
func TestEncode(t *testing.T) {
	packet, err := Encode(8)
	values, decodeErr := codec.DecodePacketExact(packet, Definition)
	if err != nil || decodeErr != nil || packet.Header != Header || values[0].Int32 != 8 {
		t.Fatalf("packet=%#v values=%#v err=%v decodeErr=%v", packet, values, err, decodeErr)
	}
}
