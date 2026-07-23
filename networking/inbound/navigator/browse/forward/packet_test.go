package forward

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies FORWARD_TO_SOME_ROOM decoding.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.String("random_friending_room"))
	if err != nil {
		t.Fatalf("create packet: %v", err)
	}
	payload, err := Decode(packet)
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if payload.Action != "random_friending_room" {
		t.Fatalf("unexpected action %q", payload.Action)
	}
}
