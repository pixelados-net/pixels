package updated

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies CATALOG_PUBLISHED packet construction.
func TestEncode(t *testing.T) {
	packet, err := Encode()
	values, decodeErr := codec.DecodePacketExact(packet, Definition)
	if err != nil || decodeErr != nil || packet.Header != Header || values[0].Boolean {
		t.Fatalf("unexpected packet %#v errors %v %v", packet, err, decodeErr)
	}
}
