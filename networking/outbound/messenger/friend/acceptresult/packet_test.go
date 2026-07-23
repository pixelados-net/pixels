package acceptresult

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeWritesFailures verifies acceptance failure pairs.
func TestEncodeWritesFailures(t *testing.T) {
	packet, err := Encode([]Failure{{PlayerID: 7, ErrorCode: 2}})
	values, decodeErr := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field})
	if err != nil || decodeErr != nil || values[0].Int32 != 1 || values[2].Int32 != 2 {
		t.Fatalf("unexpected values=%#v err=%v decode=%v", values, err, decodeErr)
	}
}
