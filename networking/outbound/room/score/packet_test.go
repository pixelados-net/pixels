package score

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies ROOM_SCORE field order.
func TestEncode(t *testing.T) {
	packet, err := Encode(12, true)
	if err != nil {
		t.Fatalf("encode score: %v", err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || values[0].Int32 != 12 || !values[1].Boolean {
		t.Fatalf("decode score: values=%+v err=%v", values, err)
	}
}

// BenchmarkEncode measures ROOM_SCORE encoding.
func BenchmarkEncode(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		_, _ = Encode(12, true)
	}
}
