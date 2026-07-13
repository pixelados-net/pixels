package additem

import (
	"github.com/niflaot/pixels/networking/codec"
	"testing"
)

// TestDecode verifies one offered item id.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, codec.Definition{codec.Int32Field}, codec.Int32(4))
	value, err := Decode(packet)
	if err != nil || value != 4 {
		t.Fatalf("value=%d err=%v", value, err)
	}
}
