package followerror

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeWritesCode verifies follow-error encoding.
func TestEncodeWritesCode(t *testing.T) {
	packet, err := Encode(3)
	values, decodeErr := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field})
	if err != nil || decodeErr != nil || values[0].Int32 != 3 {
		t.Fatalf("unexpected values=%#v err=%v decode=%v", values, err, decodeErr)
	}
}
