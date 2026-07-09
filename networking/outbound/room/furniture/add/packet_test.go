package add

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies ADD_FLOOR_ITEM encoding.
func TestEncode(t *testing.T) {
	packet, err := Encode(FloorItem{
		ID: 1, SpriteID: 39, X: 3, Y: 3, Rotation: 4, Z: "0", ExtraHeight: "1",
		ExtraData: "0", UsagePolicy: 1, OwnerID: 7, OwnerName: "demo",
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
	if values[0].Int32 != 1 || values[1].Int32 != 39 || values[2].Int32 != 3 || values[3].Int32 != 3 {
		t.Fatalf("unexpected floor fields %#v", values)
	}
	if values[5].String != "0" || values[6].String != "1" {
		t.Fatalf("unexpected height fields %#v", values)
	}
	if values[7].Int32 != defaultKind || values[8].Int32 != nonLimitedFlag || values[10].Int32 != unknownExpiration {
		t.Fatalf("unexpected constant fields %#v", values)
	}
	if values[9].String != "0" || values[11].Int32 != 1 || values[12].Int32 != 7 || values[13].String != "demo" {
		t.Fatalf("unexpected owner/usage fields %#v", values)
	}
}
