package recycle

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies the recycler item list wire.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Int32(2), codec.Int32(21), codec.Int32(22))
	payload, err := Decode(packet)
	if err != nil || len(payload.ItemIDs) != 2 || payload.ItemIDs[0] != 21 {
		t.Fatalf("unexpected payload %#v error=%v", payload, err)
	}
}
