package status

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies EMAIL_STATUS wire compatibility.
func TestEncode(t *testing.T) {
	packet, err := Encode("", false, false)
	if err != nil {
		t.Fatalf("encode status: %v", err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || values[0].String != "" || values[1].Boolean || values[2].Boolean {
		t.Fatalf("decode status: %#v, %v", values, err)
	}
}
