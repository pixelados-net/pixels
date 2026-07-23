package unavailable

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies PURCHASE_NOT_ALLOWED packet construction.
func TestEncode(t *testing.T) {
	packet, err := Encode(CodeIllegal)
	values, decodeErr := codec.DecodePacketExact(packet, Definition)
	if err != nil || decodeErr != nil || values[0].Int32 != CodeIllegal {
		t.Fatalf("unexpected packet %#v errors %v %v", packet, err, decodeErr)
	}
}
