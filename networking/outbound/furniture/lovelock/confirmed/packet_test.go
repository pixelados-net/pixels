package confirmed

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies the partial lovelock confirmation wire shape.
func TestEncode(t *testing.T) {
	packet, err := Encode(7)
	if err != nil {
		t.Fatal(err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || packet.Header != Header || values[0].Int32 != 7 {
		t.Fatalf("values=%#v err=%v", values, err)
	}
}
