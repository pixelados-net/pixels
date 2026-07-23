package readmarker

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeReadsMarkers verifies bounded marker triples.
func TestDecodeReadsMarkers(t *testing.T) {
	packet, _ := codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Int32(1), codec.Int32(7), codec.Int32(8), codec.Int32(0))
	markers, err := Decode(packet)
	if err != nil || len(markers) != 1 || markers[0].LastMessageID != 8 {
		t.Fatalf("markers=%#v err=%v", markers, err)
	}
}
