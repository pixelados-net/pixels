package get

import (
	"testing"

	craftrecord "github.com/niflaot/pixels/internal/realm/crafting/record"
	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies compatibility tier and prize nesting.
func TestEncode(t *testing.T) {
	chances := [6]int32{0, 1, 5, 20, 100, 1000}
	packet, err := Encode([]craftrecord.Prize{{Tier: 3, RewardName: "rare", TypeCode: "s", SpriteID: 44}}, chances)
	values, decodeErr := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.StringField, codec.Int32Field, codec.StringField, codec.Int32Field})
	if err != nil || decodeErr != nil || values[0].Int32 != 1 || values[1].Int32 != 3 || values[2].Int32 != 20 || values[4].String != "rare" {
		t.Fatalf("unexpected values %#v error=%v/%v", values, err, decodeErr)
	}
}
