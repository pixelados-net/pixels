package settings

import (
	"testing"

	roomdecor "github.com/niflaot/pixels/internal/realm/room/decoration"
	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies mood-light selected slot and ordered presets.
func TestEncode(t *testing.T) {
	presets := []roomdecor.Preset{{ID: 1, Color: "#74F5F5", Brightness: 180}, {ID: 2, BackgroundOnly: true, Color: "#E759DE", Brightness: 220, Selected: true}}
	packet, err := Encode(presets)
	if err != nil || packet.Header != Header {
		t.Fatalf("encode settings: %#v err=%v", packet, err)
	}
	prefix, rest, err := codec.DecodePayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field}, packet.Payload)
	if err != nil || prefix[0].Int32 != 2 || prefix[1].Int32 != 2 {
		t.Fatalf("unexpected prefix %#v err=%v", prefix, err)
	}
	first, _, err := codec.DecodePayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field, codec.StringField, codec.Int32Field}, rest)
	if err != nil || first[0].Int32 != 1 || first[2].String != "#74F5F5" {
		t.Fatalf("unexpected first preset %#v err=%v", first, err)
	}
}
