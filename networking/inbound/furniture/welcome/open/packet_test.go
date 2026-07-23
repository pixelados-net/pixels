package open

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies the retired welcome gift payload.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Int32(7))
	if err != nil {
		t.Fatalf("build packet: %v", err)
	}
	payload, err := Decode(packet)
	if err != nil || payload.ItemID != 7 {
		t.Fatalf("decode open: %#v, %v", payload, err)
	}
}
