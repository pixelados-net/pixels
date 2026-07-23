package get

import (
	"testing"

	craftrecord "github.com/niflaot/pixels/internal/realm/crafting/record"
	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeGolden verifies amount precedes the ingredient product name.
func TestEncodeGolden(t *testing.T) {
	packet, err := Encode([]craftrecord.Ingredient{{Amount: 2, Name: "wood"}})
	values, errDecode := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field, codec.Int32Field, codec.StringField})
	if err != nil || errDecode != nil || values[0].Int32 != 1 || values[1].Int32 != 2 || values[2].String != "wood" {
		t.Fatalf("unexpected values %#v error=%v/%v", values, err, errDecode)
	}
}
