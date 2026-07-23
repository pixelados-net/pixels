package homeroom

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies home-room decoding.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, Definition, codec.Int32(7))
	payload, err := Decode(packet)
	if err != nil || payload.RoomID != 7 {
		t.Fatalf("decode home room: %#v, %v", payload, err)
	}
}
