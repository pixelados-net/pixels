package apply

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies background-toner fields and header validation.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, Definition, codec.Int32(42), codec.Int32(10), codec.Int32(20), codec.Int32(30))
	payload, err := Decode(packet)
	if err != nil || payload.ItemID != 42 || payload.Hue != 10 || payload.Saturation != 20 || payload.Lightness != 30 {
		t.Fatalf("unexpected payload %#v err=%v", payload, err)
	}
	if _, err = Decode(codec.Packet{Header: Header + 1}); err == nil {
		t.Fatal("expected header error")
	}
}
