package quick

import (
	"github.com/niflaot/pixels/networking/codec"
	"testing"
)

// TestDecode verifies focused room settings fields.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Int32(9), codec.Int32(3), codec.Int32(2))
	value, err := Decode(packet)
	if err != nil || value.RoomID != 9 || value.CategoryID != 3 || value.TradeMode != 2 {
		t.Fatalf("value=%#v err=%v", value, err)
	}
}
