package rolling

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies the exact legacy wire shape consumed by Nitro.
func TestEncode(t *testing.T) {
	packet, err := Encode(2, 3, 3, 3, []Item{{ID: 8, FromZ: "0.5", ToZ: "0"}}, 9,
		WithUnit(Unit{RoomIndex: 4, FromZ: "0.5", ToZ: "0"}))
	if err != nil {
		t.Fatalf("encode rolling packet: %v", err)
	}
	definition, _ := payloadShape(1, true)
	values, err := codec.DecodePacketExact(packet, definition)
	if err != nil {
		t.Fatalf("decode rolling packet: %v", err)
	}
	if packet.Header != Header || values[0].Int32 != 2 || values[1].Int32 != 3 || values[2].Int32 != 3 || values[3].Int32 != 3 {
		t.Fatalf("unexpected packet %#v", packet)
	}
	if values[4].Int32 != 1 || values[5].Int32 != 8 || values[8].Int32 != 9 || values[9].Int32 != movementSlide || values[10].Int32 != 4 {
		t.Fatalf("unexpected rolling fields %#v", values)
	}
}

// TestEncodeFurnitureOnlyOmitsUnitTail verifies optional unit fields are absent.
func TestEncodeFurnitureOnlyOmitsUnitTail(t *testing.T) {
	packet, err := Encode(1, 1, 2, 1, []Item{{ID: 8, FromZ: "0.5", ToZ: "0"}}, 9)
	if err != nil {
		t.Fatalf("encode furniture roll: %v", err)
	}
	definition, _ := payloadShape(1, false)
	if _, err = codec.DecodePacketExact(packet, definition); err != nil {
		t.Fatalf("decode furniture roll: %v", err)
	}
}
