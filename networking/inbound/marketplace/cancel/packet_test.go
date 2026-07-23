package cancel

import (
	"github.com/niflaot/pixels/networking/codec"
	"testing"
)

// TestDecode verifies Marketplace listing cancellation decoding.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, codec.Definition{codec.Int32Field}, codec.Int32(7))
	value, err := Decode(packet)
	if err != nil || value != 7 {
		t.Fatalf("value=%d err=%v", value, err)
	}
}
