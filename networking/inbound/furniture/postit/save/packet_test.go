package save

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies initial post-it fields and header validation.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, Definition, codec.Int32(42), codec.String(":w=2,3 l=4,5 r"), codec.String("FFFF33"), codec.String("hello"))
	payload, err := Decode(packet)
	if err != nil || payload.ItemID != 42 || payload.Color != "FFFF33" || payload.Text != "hello" {
		t.Fatalf("unexpected payload %#v err=%v", payload, err)
	}
	if _, err = Decode(codec.Packet{Header: Header + 1}); err == nil {
		t.Fatal("expected header error")
	}
}
