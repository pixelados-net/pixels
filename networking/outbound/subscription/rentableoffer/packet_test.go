package rentableoffer

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies all rentable product offer fields.
func TestEncode(t *testing.T) {
	packet, err := Encode(false, "chair", true, 10, 2, 5)
	if err != nil {
		t.Fatal(err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || packet.Header != Header || values[0].Boolean || values[1].String != "chair" || !values[2].Boolean || values[3].Int32 != 10 || values[4].Int32 != 2 || values[5].Int32 != 5 {
		t.Fatalf("values=%#v err=%v", values, err)
	}
}
