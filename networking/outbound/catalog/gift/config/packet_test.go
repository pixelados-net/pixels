package config

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeWritesWrappingLists verifies all Nitro wrapping lists and counts.
func TestEncodeWritesWrappingLists(t *testing.T) {
	packet, err := Encode(Options{Price: 2, Wrappers: []int32{3372, 3373}, Boxes: []int32{0, 1},
		Ribbons: []int32{3, 4}, DefaultGifts: []int32{187}})
	if err != nil || packet.Header != Header {
		t.Fatalf("packet=%#v error=%v", packet, err)
	}
	definition := codec.Definition{codec.BooleanField, codec.Int32Field, codec.Int32Field, codec.Int32Field,
		codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field,
		codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field}
	values, rest, err := codec.DecodePayload(nil, definition, packet.Payload)
	if err != nil || len(rest) != 0 || !values[0].Boolean || values[1].Int32 != 2 || values[2].Int32 != 2 ||
		values[5].Int32 != 2 || values[8].Int32 != 2 || values[11].Int32 != 1 || values[12].Int32 != 187 {
		t.Fatalf("values=%#v rest=%d error=%v", values, len(rest), err)
	}
}
