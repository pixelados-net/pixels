package directory

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies the renderer wire shape and header validation.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, codec.Definition{})
	if err != nil {
		t.Fatal(err)
	}
	if err := Decode(packet); err != nil {
		t.Fatal(err)
	}
	if err := Decode(codec.Packet{Header: Header + 1}); err == nil {
		t.Fatal("expected header validation")
	}
}
