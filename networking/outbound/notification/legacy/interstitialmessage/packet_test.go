package interstitialmessage

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies the legacy boolean shape.
func TestEncode(t *testing.T) {
	packet, err := Encode(false)
	if err != nil {
		t.Fatal(err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || packet.Header != Header || values[0].Boolean {
		t.Fatalf("values=%#v err=%v", values, err)
	}
}
