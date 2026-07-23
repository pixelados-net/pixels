package sets

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies Renderer-compatible two-list clothing wire order.
func TestEncode(t *testing.T) {
	packet, err := Encode([]int32{101, 202}, []string{"shirt_demo"})
	if err != nil || packet.Header != Header {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
	values, remaining, err := codec.DecodePayload(nil, codec.Definition{
		codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.StringField,
	}, packet.Payload)
	if err != nil || len(remaining) != 0 || values[0].Int32 != 2 || values[3].Int32 != 1 || values[4].String != "shirt_demo" {
		t.Fatalf("values=%#v err=%v", values, err)
	}
}
