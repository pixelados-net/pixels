package equip

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode validates fixed badge slot decoding.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition,
		codec.Int32(1), codec.String("ADM"), codec.Int32(2), codec.String(""),
		codec.Int32(3), codec.String("HC1"), codec.Int32(4), codec.String(""),
		codec.Int32(5), codec.String(""))
	if err != nil {
		t.Fatal(err)
	}
	badges, err := Decode(packet)
	if err != nil || badges[0] != "ADM" || badges[2] != "HC1" {
		t.Fatalf("badges=%v err=%v", badges, err)
	}
}

// TestDecodeRejectsInvalidSlots verifies slot integrity.
func TestDecodeRejectsInvalidSlots(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition,
		codec.Int32(1), codec.String("ADM"), codec.Int32(1), codec.String("HC1"),
		codec.Int32(3), codec.String(""), codec.Int32(4), codec.String(""),
		codec.Int32(5), codec.String(""))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = Decode(packet); !errors.Is(err, ErrInvalidSlot) {
		t.Fatalf("slot error=%v", err)
	}
}
