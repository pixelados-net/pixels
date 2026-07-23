package get

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies wardrobe-page decoding.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, Definition, codec.Int32(1))
	payload, err := Decode(packet)
	if err != nil || payload.PageID != 1 {
		t.Fatalf("decode wardrobe page: %#v, %v", payload, err)
	}
}
