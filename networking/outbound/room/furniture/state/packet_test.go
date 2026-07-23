package state

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies FURNITURE_STATE encoding.
func TestEncode(t *testing.T) {
	packet, err := Encode(17, 2)
	if err != nil {
		t.Fatalf("encode state: %v", err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		t.Fatalf("decode state: %v", err)
	}
	if packet.Header != Header || values[0].Int32 != 17 || values[1].Int32 != 2 {
		t.Fatalf("unexpected state packet %#v values=%#v", packet, values)
	}
}
