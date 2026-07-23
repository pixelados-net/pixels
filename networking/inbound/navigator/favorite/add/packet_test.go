package favoriteadd

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies ROOM_FAVORITE decoding.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Int32(9))
	if err != nil {
		t.Fatalf("new packet: %v", err)
	}
	payload, err := Decode(packet)
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if payload.RoomID != 9 {
		t.Fatalf("unexpected payload %#v", payload)
	}
}
