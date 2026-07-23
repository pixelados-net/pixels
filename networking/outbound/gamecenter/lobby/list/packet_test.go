package list

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies the variable game-list wire shape.
func TestEncode(t *testing.T) {
	packet, err := Encode([]Game{{ID: 7, Name: "pixels.game", BackgroundColor: "112233", TextColor: "ffffff", AssetURL: "asset", SupportURL: "support"}})
	if err != nil {
		t.Fatal(err)
	}
	values, rest, err := codec.DecodePacket(packet, codec.Definition{codec.Int32Field})
	if err != nil || values[0].Int32 != 1 {
		t.Fatalf("unexpected count: %v %v", values, err)
	}
	entry, remaining, err := codec.DecodePayload(nil, codec.Definition{codec.Int32Field, codec.StringField, codec.StringField, codec.StringField, codec.StringField, codec.StringField}, rest)
	if err != nil || len(remaining) != 0 || entry[0].Int32 != 7 || entry[5].String != "support" {
		t.Fatalf("unexpected entry: %v %v", entry, err)
	}
}
