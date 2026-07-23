package wallmultistate

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies Nitro's two-field wall interaction payload.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Int32(910134), codec.Int32(0))
	if err != nil {
		t.Fatalf("packet: %v", err)
	}
	payload, err := Decode(packet)
	if err != nil || payload.ItemID != 910134 || payload.State != 0 {
		t.Fatalf("unexpected payload=%+v err=%v", payload, err)
	}
}
