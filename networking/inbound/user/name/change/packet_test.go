package change

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies username change decoding.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, Definition, codec.String("Pixel"))
	payload, err := Decode(packet)
	if err != nil || payload.Username != "Pixel" {
		t.Fatalf("decode username: %#v, %v", payload, err)
	}
}
