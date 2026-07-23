package craft

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeConditionalGolden verifies failure omits and success includes result strings.
func TestEncodeConditionalGolden(t *testing.T) {
	failure, err := Encode(false)
	if err != nil || len(failure.Payload) != 1 || failure.Payload[0] != 0 {
		t.Fatalf("unexpected failure %#v error=%v", failure, err)
	}
	success, err := Encode(true, WithProduct("recipe", "item"))
	values, decodeErr := codec.DecodePacketExact(success, codec.Definition{codec.BooleanField, codec.StringField, codec.StringField})
	if err != nil || decodeErr != nil || !values[0].Boolean || values[1].String != "recipe" || values[2].String != "item" {
		t.Fatalf("unexpected success %#v error=%v/%v", values, err, decodeErr)
	}
}
