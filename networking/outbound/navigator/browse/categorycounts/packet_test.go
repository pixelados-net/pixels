package categorycounts

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies CATEGORIES_WITH_VISITOR_COUNT packet construction.
func TestEncode(t *testing.T) {
	packet, err := Encode([]Entry{{CategoryID: 2, CurrentVisitorCount: 5, MaxVisitorCount: 100}})
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}
	values, rest, err := codec.DecodePacket(packet, Definition)
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if packet.Header != Header || values[0].Int32 != 1 || len(rest) == 0 {
		t.Fatalf("unexpected packet %#v values %#v", packet, values)
	}
}
