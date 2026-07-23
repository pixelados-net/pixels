package add

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
	outlist "github.com/niflaot/pixels/networking/outbound/inventory/furniture/list"
)

// TestEncodeUsesInventoryItemWire verifies the incremental packet reuses Nitro's item shape.
func TestEncodeUsesInventoryItemWire(t *testing.T) {
	packet, err := Encode(outlist.Item{ID: 7, SpriteID: 9, ExtraData: "0", AllowRecycle: true})
	if err != nil {
		t.Fatalf("encode incremental item: %v", err)
	}
	if packet.Header != Header || len(packet.Payload) == 0 {
		t.Fatalf("unexpected packet %#v", packet)
	}
	values, _, err := codec.DecodePayload(nil, codec.Definition{
		codec.Int32Field, codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field,
	}, packet.Payload)
	if err != nil || values[0].Int32 != 7 || values[3].Int32 != 9 {
		t.Fatalf("unexpected prefix %#v error=%v", values, err)
	}
}
