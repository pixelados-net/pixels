package stats

import (
	"github.com/niflaot/pixels/networking/codec"
	"testing"
)

// TestDecode verifies Marketplace statistics request decoding.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(1), codec.Int32(22))
	value, err := Decode(packet)
	if err != nil || value.Category != 1 || value.SpriteID != 22 {
		t.Fatalf("value=%#v err=%v", value, err)
	}
}
