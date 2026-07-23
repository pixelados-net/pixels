package settings

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies USER_SETTINGS wire order and selected style.
func TestEncode(t *testing.T) {
	packet, err := Encode(80, 70, 60, true, true, true, 3, 7)
	values, decodeErr := codec.DecodePacketExact(packet, Definition)
	if err != nil || decodeErr != nil || packet.Header != Header || values[0].Int32 != 80 || values[1].Int32 != 70 || values[2].Int32 != 60 || !values[3].Boolean || !values[4].Boolean || !values[5].Boolean || values[6].Int32 != 3 || values[7].Int32 != 7 {
		t.Fatalf("packet=%#v values=%#v err=%v decodeErr=%v", packet, values, err, decodeErr)
	}
}
