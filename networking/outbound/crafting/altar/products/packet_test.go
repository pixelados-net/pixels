package products

import (
	"testing"

	craftrecord "github.com/niflaot/pixels/internal/realm/crafting/record"
	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeGolden verifies recipe ordering and aggregate ingredient de-duplication.
func TestEncodeGolden(t *testing.T) {
	packet, err := Encode([]craftrecord.Recipe{{Name: "r", RewardName: "reward", Ingredients: []craftrecord.Ingredient{{Name: "a"}, {Name: "a"}, {Name: "b"}}}})
	if err != nil || packet.Header != Header {
		t.Fatalf("encode products: packet=%#v error=%v", packet, err)
	}
	values, rest, err := codec.DecodePayload(nil, codec.Definition{codec.Int32Field, codec.StringField, codec.StringField, codec.Int32Field}, packet.Payload)
	if err != nil || values[0].Int32 != 1 || values[1].String != "r" || values[3].Int32 != 2 || len(rest) == 0 {
		t.Fatalf("unexpected values %#v rest=%v error=%v", values, rest, err)
	}
}
