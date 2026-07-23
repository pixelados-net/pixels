package weekly

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies the renderer's five-field leaderboard parser shape.
func TestEncode(t *testing.T) {
	packet, err := Encode(2026, 30, 4, 2, 900)
	if err != nil {
		t.Fatal(err)
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field})
	if err != nil {
		t.Fatal(err)
	}
	if packet.Header != Header || values[0].Int32 != 2026 || values[4].Int32 != 900 {
		t.Fatalf("unexpected packet: %+v", packet)
	}
}
