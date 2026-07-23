package collectionstate

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies the three-counter shape.
func TestEncode(t *testing.T) {
	packet, err := Encode(1, 2, 3)
	if err != nil {
		t.Fatal(err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || packet.Header != Header || values[0].Int32 != 1 || values[1].Int32 != 2 || values[2].Int32 != 3 {
		t.Fatalf("values=%#v err=%v", values, err)
	}
}
