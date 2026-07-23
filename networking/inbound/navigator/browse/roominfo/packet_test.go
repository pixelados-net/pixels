package roominfo

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies GET_GUEST_ROOM decoding.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Int32(7), codec.Int32(1), codec.Int32(0))
	if err != nil {
		t.Fatalf("new packet: %v", err)
	}

	payload, err := Decode(packet)
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if payload.RoomID != 7 || payload.EnterRoom != 1 || payload.ForwardRoom != 0 {
		t.Fatalf("unexpected payload %#v", payload)
	}
}
