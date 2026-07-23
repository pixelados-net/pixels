package figure

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies figure response order.
func TestEncode(t *testing.T) {
	packet, err := Encode("hd-180-1", "M")
	if err != nil {
		t.Fatalf("encode figure: %v", err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || values[0].String != "hd-180-1" || values[1].String != "M" {
		t.Fatalf("decode figure: %#v, %v", values, err)
	}
}
