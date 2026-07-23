package welcomestatus

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies WELCOME_GIFT_STATUS wire compatibility.
func TestEncode(t *testing.T) {
	packet, err := Encode("", false, false, 0, false)
	if err != nil {
		t.Fatalf("encode status: %v", err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || len(values) != 5 {
		t.Fatalf("decode status: %#v, %v", values, err)
	}
}
