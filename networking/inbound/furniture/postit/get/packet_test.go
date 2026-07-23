package get

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies post-it read fields and header validation.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, Definition, codec.Int32(42))
	payload, err := Decode(packet)
	if err != nil || payload.ItemID != 42 {
		t.Fatalf("unexpected payload %#v err=%v", payload, err)
	}
	if _, err = Decode(codec.Packet{Header: Header + 1}); err == nil {
		t.Fatal("expected header error")
	}
}
