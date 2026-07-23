package list

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeUsesNativeBadgeInventoryShape verifies owned and active sections.
func TestEncodeUsesNativeBadgeInventoryShape(t *testing.T) {
	packet, err := Encode([]Badge{{ID: 10, Code: "ADM", Slot: 1}, {ID: 11, Code: "HC1"}})
	if err != nil {
		t.Fatal(err)
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{
		codec.Int32Field, codec.Int32Field, codec.StringField, codec.Int32Field, codec.StringField,
		codec.Int32Field, codec.Int32Field, codec.StringField,
	})
	if err != nil || values[0].Int32 != 2 || values[5].Int32 != 1 || values[7].String != "ADM" {
		t.Fatalf("values=%#v err=%v", values, err)
	}
}
