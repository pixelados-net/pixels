package figure

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies figure field order.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, Definition, codec.String("M"), codec.String("hd-180-1"))
	payload, err := Decode(packet)
	if err != nil || payload.Gender != "M" || payload.Figure != "hd-180-1" {
		t.Fatalf("decode figure: %#v, %v", payload, err)
	}
}
