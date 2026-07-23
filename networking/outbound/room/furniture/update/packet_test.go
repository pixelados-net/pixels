package update

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies FLOOR_ITEM_UPDATE encoding.
func TestEncode(t *testing.T) {
	packet, err := Encode(FloorItem{
		ID: 1, SpriteID: 39, X: 5, Y: 5, Rotation: 2, Z: "0", ExtraHeight: "1",
		ExtraData: "0", OwnerID: 7,
	})
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}
	if packet.Header != Header {
		t.Fatalf("unexpected header %d", packet.Header)
	}

	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if values[2].Int32 != 5 || values[3].Int32 != 5 || values[4].Int32 != 2 {
		t.Fatalf("unexpected position fields %#v", values)
	}
	if values[7].Int32 != updateKind || values[11].Int32 != updateUsagePolicy {
		t.Fatalf("unexpected constant fields %#v", values)
	}
	if values[12].Int32 != 7 {
		t.Fatalf("unexpected owner field %#v", values[12])
	}
}
