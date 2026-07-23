package start

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies both lovelock start fields.
func TestEncode(t *testing.T) {
	packet, err := Encode(7, true)
	if err != nil {
		t.Fatal(err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || packet.Header != Header || values[0].Int32 != 7 || !values[1].Boolean {
		t.Fatalf("values=%#v err=%v", values, err)
	}
}
