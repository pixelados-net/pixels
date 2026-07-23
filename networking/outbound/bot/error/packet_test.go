package error

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode writes the native bot error code.
func TestEncode(t *testing.T) {
	packet, err := Encode(3)
	values, decodeErr := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field})
	if err != nil || decodeErr != nil || values[0].Int32 != 3 {
		t.Fatalf("values=%#v encode=%v decode=%v", values, err, decodeErr)
	}
}
