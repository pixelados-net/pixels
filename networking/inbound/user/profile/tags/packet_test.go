package tags

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies the room unit identifier.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, Definition, codec.Int32(9))
	payload, err := Decode(packet)
	if err != nil || payload.RoomUnitID != 9 {
		t.Fatalf("decode tags: %#v, %v", payload, err)
	}
}
