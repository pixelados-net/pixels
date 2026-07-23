package change

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies the retired email payload shape.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.String("demo@example.test"))
	if err != nil {
		t.Fatalf("build packet: %v", err)
	}
	payload, err := Decode(packet)
	if err != nil || payload.Email != "demo@example.test" {
		t.Fatalf("decode email: %#v, %v", payload, err)
	}
}
