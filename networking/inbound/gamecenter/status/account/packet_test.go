package account

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies the renderer wire shape and header validation.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, codec.Definition{codec.Int32Field}, codec.Int32(7))
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := Decode(packet)
	if err != nil {
		t.Fatal(err)
	}
	if decoded != 7 {
		t.Fatalf("unexpected value: %v", decoded)
	}
	packet.Header++
	if _, err := Decode(packet); err == nil {
		t.Fatal("expected header validation")
	}
}
