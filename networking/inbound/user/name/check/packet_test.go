package check

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies username check decoding.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, Definition, codec.String("Pixel"))
	payload, err := Decode(packet)
	if err != nil || payload.Username != "Pixel" {
		t.Fatalf("decode username: %#v, %v", payload, err)
	}
}
