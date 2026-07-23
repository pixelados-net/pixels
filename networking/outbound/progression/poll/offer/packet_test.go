package offer

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies the renderer wire shape.
func TestEncode(t *testing.T) {
	packet, err := Encode(11, "value1", "value2", "value3")
	if err != nil {
		t.Fatal(err)
	}
	if packet.Header != Header {
		t.Fatalf("unexpected header: %d", packet.Header)
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field, codec.StringField, codec.StringField, codec.StringField})
	if err != nil {
		t.Fatal(err)
	}
	if values[0].Int32 != 11 {
		t.Fatalf("unexpected field 0: %v", values[0].Int32)
	}
	if values[1].String != "value1" {
		t.Fatalf("unexpected field 1: %v", values[1].String)
	}
	if values[2].String != "value2" {
		t.Fatalf("unexpected field 2: %v", values[2].String)
	}
	if values[3].String != "value3" {
		t.Fatalf("unexpected field 3: %v", values[3].String)
	}
}
