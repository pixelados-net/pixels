package walladd

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies newly placed wall-item field order.
func TestEncode(t *testing.T) {
	definition := codec.Definition{codec.StringField, codec.Int32Field, codec.StringField, codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.StringField}
	packet, err := Encode(Item{ID: 42, SpriteID: 9, WallPosition: ":w=2,3 l=4,5 r", ExtraData: "0", OwnerID: 7, OwnerName: "demo"})
	values, decodeErr := codec.DecodePacketExact(packet, definition)
	if err != nil || decodeErr != nil || packet.Header != Header || values[0].String != "42" || values[2].String != ":w=2,3 l=4,5 r" || values[7].String != "demo" {
		t.Fatalf("unexpected packet %#v values=%#v err=%v/%v", packet, values, err, decodeErr)
	}
}
