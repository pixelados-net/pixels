package publishstatus

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies failed and successful conditional wires.
func TestEncode(t *testing.T) {
	failed, _ := Encode(false, 8)
	values, err := codec.DecodePacketExact(failed, codec.Definition{codec.BooleanField, codec.Int32Field})
	if err != nil || values[0].Boolean || values[1].Int32 != 8 {
		t.Fatalf("failed=%+v err=%v", values, err)
	}
	success, _ := Encode(true, 0, WithURL("https://storage/photo.png"))
	values, err = codec.DecodePacketExact(success, codec.Definition{codec.BooleanField, codec.Int32Field, codec.StringField})
	if err != nil || values[2].String != "https://storage/photo.png" {
		t.Fatalf("success=%+v err=%v", values, err)
	}
}
