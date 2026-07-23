package welcome

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies the retired welcome email payload shape.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.String("ignored@example.test"))
	if err != nil {
		t.Fatalf("build packet: %v", err)
	}
	payload, err := Decode(packet)
	if err != nil || payload.Email == "" {
		t.Fatalf("decode welcome email: %#v, %v", payload, err)
	}
}
