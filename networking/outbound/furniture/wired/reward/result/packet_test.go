package result

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies WIRED_REWARD encoding.
func TestEncode(t *testing.T) {
	packet, err := Encode(8)
	if err != nil {
		t.Fatalf("encode reward: %v", err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || values[0].Int32 != 8 {
		t.Fatalf("unexpected reward %#v %v", values, err)
	}
}
