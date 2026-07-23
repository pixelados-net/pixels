package roommessage

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies the legacy room-notice shape.
func TestEncode(t *testing.T) {
	packet, err := Encode(7, "notice", 2)
	if err != nil {
		t.Fatal(err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || packet.Header != Header || values[0].Int32 != 7 || values[1].String != "notice" || values[2].Int32 != 2 {
		t.Fatalf("values=%#v err=%v", values, err)
	}
}
