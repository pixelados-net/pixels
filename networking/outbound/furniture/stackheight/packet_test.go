package stackheight

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies the centimeter-precision stack-height response.
func TestEncode(t *testing.T) {
	packet, err := Encode(7, 123)
	if err != nil {
		t.Fatal(err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || packet.Header != Header || values[0].Int32 != 7 || values[1].Int32 != 123 {
		t.Fatalf("values=%#v err=%v", values, err)
	}
}
