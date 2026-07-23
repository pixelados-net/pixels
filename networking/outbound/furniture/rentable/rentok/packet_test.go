package rentok

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies the rentable expiry timestamp.
func TestEncode(t *testing.T) {
	packet, err := Encode(123)
	if err != nil {
		t.Fatal(err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || packet.Header != Header || values[0].Int32 != 123 {
		t.Fatalf("values=%#v err=%v", values, err)
	}
}
