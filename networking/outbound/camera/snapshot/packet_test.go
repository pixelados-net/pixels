package snapshot

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies the compatibility snapshot wire.
func TestEncode(t *testing.T) {
	packet, err := Encode("private", 8)
	values, decodeErr := codec.DecodePacketExact(packet, Definition)
	if err != nil || decodeErr != nil || values[0].String != "private" || values[1].Int32 != 8 {
		t.Fatalf("values=%+v err=%v decode=%v", values, err, decodeErr)
	}
}
