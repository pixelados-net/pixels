package thumbnailupdateresult

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies the compatibility result wire.
func TestEncode(t *testing.T) {
	packet, err := Encode(7, 0)
	values, decodeErr := codec.DecodePacketExact(packet, Definition)
	if err != nil || decodeErr != nil || values[0].Int32 != 7 || values[1].Int32 != 0 {
		t.Fatalf("values=%+v err=%v decode=%v", values, err, decodeErr)
	}
}
