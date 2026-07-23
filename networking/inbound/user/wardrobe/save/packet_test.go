package save

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies wardrobe save order.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, Definition, codec.Int32(2), codec.String("hd-180-1"), codec.String("M"))
	payload, err := Decode(packet)
	if err != nil || payload.SlotID != 2 || payload.Gender != "M" {
		t.Fatalf("decode wardrobe: %#v, %v", payload, err)
	}
}
