package favoritechanged

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies USER_FAVORITE_ROOM packet construction.
func TestEncode(t *testing.T) {
	packet, err := Encode(8, true)
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if packet.Header != Header || !values[1].Boolean {
		t.Fatalf("unexpected packet %#v values %#v", packet, values)
	}
}
