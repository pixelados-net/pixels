package volume

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies volume field order.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, Definition, codec.Int32(1), codec.Int32(2), codec.Int32(3))
	payload, err := Decode(packet)
	if err != nil || payload.System != 1 || payload.Furniture != 2 || payload.Trax != 3 {
		t.Fatalf("decode volume: %#v, %v", payload, err)
	}
}
