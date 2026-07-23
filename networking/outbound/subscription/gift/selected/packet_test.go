package selected

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
	catalogoffer "github.com/niflaot/pixels/networking/outbound/catalog/offer"
)

// TestEncodeWritesCompleteProduct verifies the required limited marker.
func TestEncodeWritesCompleteProduct(t *testing.T) {
	packet, err := Encode("gift", []catalogoffer.Product{{Type: "s", ClassID: 3, ExtraData: "0", Amount: 1}})
	if err != nil || packet.Header != Header {
		t.Fatalf("packet=%#v error=%v", packet, err)
	}
	values, remaining, err := codec.DecodePayload(nil, codec.Definition{
		codec.StringField, codec.Int32Field, codec.StringField, codec.Int32Field,
		codec.StringField, codec.Int32Field, codec.BooleanField,
	}, packet.Payload)
	if err != nil || len(remaining) != 0 || values[0].String != "gift" ||
		values[1].Int32 != 1 || values[3].Int32 != 3 || values[6].Boolean {
		t.Fatalf("values=%#v remaining=%d error=%v", values, len(remaining), err)
	}
}
