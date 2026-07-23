package classification

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies bounded classification wire.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, Definition, codec.String("peer"))
	payload, err := Decode(packet)
	if err != nil || payload.ClassType != "peer" {
		t.Fatalf("decode classification: %#v, %v", payload, err)
	}
}
