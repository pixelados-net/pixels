package newrequest

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeWritesRequester verifies live request encoding.
func TestEncodeWritesRequester(t *testing.T) {
	packet, err := Encode(7, "demo", "look")
	values, decodeErr := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field, codec.StringField, codec.StringField})
	if err != nil || decodeErr != nil || values[1].String != "demo" {
		t.Fatalf("unexpected values=%#v err=%v decode=%v", values, err, decodeErr)
	}
}
