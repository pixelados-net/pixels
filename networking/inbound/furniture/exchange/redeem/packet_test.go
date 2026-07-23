package redeem

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies exchange derives value from only the item identifier.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, codec.Definition{codec.Int32Field}, codec.Int32(77))
	payload, err := Decode(packet)
	if err != nil || payload.ItemID != 77 {
		t.Fatalf("unexpected payload %#v error=%v", payload, err)
	}
}
