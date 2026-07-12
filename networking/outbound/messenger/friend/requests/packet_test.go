package requests

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeIncludesPendingCount verifies native request count placement.
func TestEncodeIncludesPendingCount(t *testing.T) {
	packet, err := Encode(3, []Request{{PlayerID: 7, Username: "demo", Look: "look"}})
	values, _, decodeErr := codec.DecodePacket(packet, codec.Definition{codec.Int32Field, codec.Int32Field})
	if err != nil || decodeErr != nil || values[0].Int32 != 3 || values[1].Int32 != 1 {
		t.Fatalf("unexpected values=%#v err=%v decode=%v", values, err, decodeErr)
	}
}
