package visituser

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies VISIT_USER wire validation.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.String("Alice"))
	if err != nil {
		t.Fatalf("new packet: %v", err)
	}
	payload, err := Decode(packet)
	if err != nil || payload.Username != "Alice" {
		t.Fatalf("payload=%#v err=%v", payload, err)
	}
}
