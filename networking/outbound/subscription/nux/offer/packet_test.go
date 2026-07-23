package offer

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies the retired NUX offer is an empty list.
func TestEncode(t *testing.T) {
	packet, err := Encode()
	if err != nil {
		t.Fatalf("encode offer: %v", err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || values[0].Int32 != 0 {
		t.Fatalf("decode offer: %#v, %v", values, err)
	}
}
