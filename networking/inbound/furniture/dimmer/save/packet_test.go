package save

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies mood-light preset fields and header validation.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, Definition, codec.Int32(2), codec.Int32(1), codec.String("#74F5F5"), codec.Int32(180), codec.Bool(true))
	payload, err := Decode(packet)
	if err != nil || payload.PresetID != 2 || payload.Type != 1 || payload.Color != "#74F5F5" || payload.Brightness != 180 || !payload.Apply {
		t.Fatalf("unexpected payload %#v err=%v", payload, err)
	}
	if _, err = Decode(codec.Packet{Header: Header + 1}); err == nil {
		t.Fatal("expected header error")
	}
}
