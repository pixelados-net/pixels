package current

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeUsesNativeCurrentBadgeShape verifies player, count, and slots.
func TestEncodeUsesNativeCurrentBadgeShape(t *testing.T) {
	packet, err := Encode(7, []Badge{{Slot: 1, Code: "ADM"}})
	if err != nil {
		t.Fatal(err)
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.StringField})
	if err != nil || values[0].Int32 != 7 || values[1].Int32 != 1 || values[3].String != "ADM" {
		t.Fatalf("values=%#v err=%v", values, err)
	}
}
