package craft

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies the bounded secret combination wire.
func TestDecode(t *testing.T) {
	packet, _ := codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Int32(9), codec.Int32(2), codec.Int32(11), codec.Int32(12))
	payload, err := Decode(packet)
	if err != nil || payload.AltarItemID != 9 || len(payload.ItemIDs) != 2 || payload.ItemIDs[1] != 12 {
		t.Fatalf("unexpected payload %#v error=%v", payload, err)
	}
}

// TestDecodeRejectsOversizedCount verifies hostile counts do not allocate.
func TestDecodeRejectsOversizedCount(t *testing.T) {
	packet, _ := codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(9), codec.Int32(MaxItems+1))
	if _, err := Decode(packet); err == nil {
		t.Fatal("expected oversized count rejection")
	}
}
