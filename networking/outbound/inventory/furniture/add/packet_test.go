package add

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies ADD_FURNITURE_TO_INVENTORY encoding.
func TestEncode(t *testing.T) {
	packet, err := Encode(42)
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
	if values[0].Int32 != 1 || values[1].Int32 != ownedFurniCategory || values[2].Int32 != 1 || values[3].Int32 != 42 {
		t.Fatalf("unexpected values %#v", values)
	}
}
