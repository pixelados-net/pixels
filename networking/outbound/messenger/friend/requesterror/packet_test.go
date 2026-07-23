package requesterror

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeWritesRequestError verifies request-error encoding.
func TestEncodeWritesRequestError(t *testing.T) {
	packet, err := Encode(0, 4)
	values, decodeErr := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field, codec.Int32Field})
	if err != nil || decodeErr != nil || values[1].Int32 != 4 {
		t.Fatalf("unexpected values=%#v err=%v decode=%v", values, err, decodeErr)
	}
}
