package status

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies all rentable-space state fields.
func TestEncode(t *testing.T) {
	packet, err := Encode(true, 0, 7, "demo", 60, 10)
	if err != nil {
		t.Fatal(err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || packet.Header != Header || !values[0].Boolean || values[1].Int32 != 0 || values[2].Int32 != 7 || values[3].String != "demo" || values[4].Int32 != 60 || values[5].Int32 != 10 {
		t.Fatalf("values=%#v err=%v", values, err)
	}
}
