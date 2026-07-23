package chat

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies the renderer wire shape and header validation.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, codec.Definition{codec.StringField}, codec.String("hello"))
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := Decode(packet)
	if err != nil {
		t.Fatal(err)
	}
	if decoded != "hello" {
		t.Fatalf("unexpected value: %v", decoded)
	}
	packet.Header++
	if _, err := Decode(packet); err == nil {
		t.Fatal("expected header validation")
	}
}
