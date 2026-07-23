package respect

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies respect decoding.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, Definition, codec.Int32(8))
	payload, err := Decode(packet)
	if err != nil || payload.TargetPlayerID != 8 {
		t.Fatalf("decode respect: %#v, %v", payload, err)
	}
}
