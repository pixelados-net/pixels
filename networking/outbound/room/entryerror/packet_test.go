package entryerror

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies ROOM_ENTER_ERROR packet construction.
func TestEncode(t *testing.T) {
	packet, err := Encode(1)
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if values[0].Int32 != 1 || values[1].String != "" {
		t.Fatalf("unexpected values %#v", values)
	}
}

// TestEncodeWithParameter verifies optional queue error encoding.
func TestEncodeWithParameter(t *testing.T) {
	packet, err := Encode(3, WithParameter("na"))
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || values[1].String != "na" {
		t.Fatalf("unexpected values %#v err=%v", values, err)
	}
}
