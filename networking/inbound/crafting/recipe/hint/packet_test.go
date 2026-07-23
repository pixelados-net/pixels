package hint

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies empty and populated hint bags.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Int32(9), codec.Int32(1), codec.Int32(11))
	payload, err := Decode(packet)
	if err != nil || payload.AltarItemID != 9 || len(payload.ItemIDs) != 1 || payload.ItemIDs[0] != 11 {
		t.Fatalf("unexpected payload %#v error=%v", payload, err)
	}
}
