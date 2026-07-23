package prize

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies the mystery-box prize payload.
func TestEncode(t *testing.T) {
	packet, err := Encode("s", 42)
	if err != nil {
		t.Fatal(err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || packet.Header != Header || values[0].String != "s" || values[1].Int32 != 42 {
		t.Fatalf("values=%#v err=%v", values, err)
	}
}
