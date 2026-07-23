package init

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies camera price order.
func TestEncode(t *testing.T) {
	packet, err := Encode(2, 3, 10)
	values, decodeErr := codec.DecodePacketExact(packet, Definition)
	if err != nil || decodeErr != nil || values[0].Int32 != 2 || values[2].Int32 != 10 {
		t.Fatalf("values=%+v err=%v decode=%v", values, err, decodeErr)
	}
}
