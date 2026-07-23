package failed

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies Nitro's string validation error.
func TestEncode(t *testing.T) {
	packet, err := Encode("invalid")
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || values[0].String != "invalid" {
		t.Fatalf("unexpected failed %#v %v", values, err)
	}
}
