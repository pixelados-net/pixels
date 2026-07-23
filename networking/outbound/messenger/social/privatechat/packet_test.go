package privatechat

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeSupportsOptionalExtraData verifies private-chat wire shape.
func TestEncodeSupportsOptionalExtraData(t *testing.T) {
	packet, err := Encode(7, "hello", 2, WithExtraData("meta"))
	values, decodeErr := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field, codec.StringField, codec.Int32Field, codec.StringField})
	if err != nil || decodeErr != nil || values[1].String != "hello" || values[3].String != "meta" {
		t.Fatalf("unexpected values=%#v err=%v decode=%v", values, err, decodeErr)
	}
}
