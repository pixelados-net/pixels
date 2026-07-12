package inviteerror

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeWritesOneNativeFailureGroup verifies invite error shape.
func TestEncodeWritesOneNativeFailureGroup(t *testing.T) {
	packet, err := Encode(1, []int64{7, 8})
	values, decodeErr := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field})
	if err != nil || decodeErr != nil || values[1].Int32 != 2 || values[3].Int32 != 8 {
		t.Fatalf("unexpected values=%#v err=%v decode=%v", values, err, decodeErr)
	}
}
