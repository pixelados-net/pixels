package sell

import (
	"github.com/niflaot/pixels/networking/codec"
	"testing"
)

// TestDecode verifies Marketplace listing request decoding.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Int32(50), codec.Int32(1), codec.Int32(7))
	value, err := Decode(packet)
	if err != nil || value.Price != 50 || value.FurnitureType != 1 || value.ItemID != 7 {
		t.Fatalf("value=%#v err=%v", value, err)
	}
}
