package privateerror

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeWritesPrivateError verifies private-message error encoding.
func TestEncodeWritesPrivateError(t *testing.T) {
	packet, err := Encode(1, 7, "failed")
	values, decodeErr := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field, codec.Int32Field, codec.StringField})
	if err != nil || decodeErr != nil || values[1].Int32 != 7 {
		t.Fatalf("unexpected values=%#v err=%v decode=%v", values, err, decodeErr)
	}
}
