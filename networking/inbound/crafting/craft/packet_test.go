package craft

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies the named craft field order.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.StringField}, codec.Int32(7), codec.String("normal"))
	payload, err := Decode(packet)
	if err != nil || payload.AltarItemID != 7 || payload.RecipeName != "normal" {
		t.Fatalf("unexpected payload %#v error=%v", payload, err)
	}
}
