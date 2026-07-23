package single

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/outbound/furniture/stuffdata"
)

// TestEncode verifies Nitro's single object-data shape.
func TestEncode(t *testing.T) {
	packet, err := Encode(12, stuffdata.IntArray([]int32{1, 2}))
	if err != nil || packet.Header != Header {
		t.Fatalf("encode single: %v %#v", err, packet)
	}
	values, rest, err := codec.DecodePayload(nil, codec.Definition{codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field}, packet.Payload)
	if err != nil || len(rest) != 0 || values[0].String != "12" || values[1].Int32 != 5 {
		t.Fatalf("unexpected single %#v rest=%d %v", values, len(rest), err)
	}
}
