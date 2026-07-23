package customuser

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies every code is serialized as one integer.
func TestEncode(t *testing.T) {
	for _, code := range []int32{1, 2, 3} {
		packet, err := Encode(code)
		if err != nil {
			t.Fatal(err)
		}
		values, err := codec.DecodePacketExact(packet, Definition)
		if err != nil || packet.Header != Header || values[0].Int32 != code {
			t.Fatalf("code=%d values=%#v err=%v", code, values, err)
		}
	}
}
