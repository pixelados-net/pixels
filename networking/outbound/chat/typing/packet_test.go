package typing

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies both UNIT_TYPING states.
func TestEncode(t *testing.T) {
	for _, active := range []bool{false, true} {
		packet, err := Encode(12, active)
		if err != nil || packet.Header != Header {
			t.Fatalf("active=%v packet=%#v err=%v", active, packet, err)
		}
		values, err := codec.DecodePacketExact(packet, Definition)
		if err != nil || values[1].Int32 != int32(map[bool]int{false: 0, true: 1}[active]) {
			t.Fatalf("active=%v values=%#v err=%v", active, values, err)
		}
	}
}
