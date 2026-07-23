package info

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies the exact BONUS_RARE_INFO wire shape.
func TestEncode(t *testing.T) {
	packet, err := Encode("Bonus Rare", 712, 120, 75)
	if err != nil {
		t.Fatal(err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		t.Fatal(err)
	}
	if packet.Header != Header || values[0].String != "Bonus Rare" || values[1].Int32 != 712 || values[2].Int32 != 120 || values[3].Int32 != 75 {
		t.Fatalf("unexpected packet %#v values=%#v", packet, values)
	}
}
