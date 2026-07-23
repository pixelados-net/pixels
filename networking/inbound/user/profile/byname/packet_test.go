package byname

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies profile-name decoding.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, Definition, codec.String("demo"))
	payload, err := Decode(packet)
	if err != nil || payload.Username != "demo" {
		t.Fatalf("decode profile: %#v, %v", payload, err)
	}
}
