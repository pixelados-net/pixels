package motto

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies motto decoding.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, Definition, codec.String("Pixels"))
	payload, err := Decode(packet)
	if err != nil || payload.Motto != "Pixels" {
		t.Fatalf("decode motto: %#v, %v", payload, err)
	}
}
