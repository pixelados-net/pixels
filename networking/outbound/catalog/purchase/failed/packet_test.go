package failed

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies PURCHASE_ERROR packet construction.
func TestEncode(t *testing.T) {
	packet, err := Encode(CodeServer)
	values, decodeErr := codec.DecodePacketExact(packet, Definition)
	if err != nil || decodeErr != nil || packet.Header != Header || values[0].Int32 != CodeServer {
		t.Fatalf("unexpected packet %#v errors %v %v", packet, err, decodeErr)
	}
}
