package thumbnailupdate

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies the legacy composition wire.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, Definition, codec.Int32(1), codec.Int32(2), codec.Int32(3), codec.Int32(4))
	payload, err := Decode(packet)
	if err != nil || payload.FlatID != 1 || payload.ObjectCount != 4 {
		t.Fatalf("payload=%+v err=%v", payload, err)
	}
}
