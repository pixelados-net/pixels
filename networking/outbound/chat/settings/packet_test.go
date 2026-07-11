package settings

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies USER_SETTINGS wire order and selected style.
func TestEncode(t *testing.T) {
	packet, err := Encode(7)
	values, decodeErr := codec.DecodePacketExact(packet, Definition)
	if err != nil || decodeErr != nil || packet.Header != Header || values[0].Int32 != 100 || values[7].Int32 != 7 {
		t.Fatalf("packet=%#v values=%#v err=%v decodeErr=%v", packet, values, err, decodeErr)
	}
}
