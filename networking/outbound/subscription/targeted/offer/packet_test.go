package offer

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeWritesTargetedOfferAndProducts verifies dynamic sub-product framing.
func TestEncodeWritesTargetedOfferAndProducts(t *testing.T) {
	packet, err := Encode(Offer{ID: 4, Identifier: "dev", ProductCode: "sofa", PurchaseLimit: 3, Title: "Title", SubProducts: []string{"chair"}})
	if err != nil || packet.Header != Header {
		t.Fatalf("packet=%#v error=%v", packet, err)
	}
	values, rest, err := codec.DecodePayload(nil, codec.Definition{
		codec.Int32Field, codec.Int32Field, codec.StringField, codec.StringField,
	}, packet.Payload)
	if err != nil || len(rest) == 0 || values[0].Int32 != 0 || values[1].Int32 != 4 ||
		values[2].String != "dev" || values[3].String != "sofa" {
		t.Fatalf("values=%#v rest=%d error=%v", values, len(rest), err)
	}
}
