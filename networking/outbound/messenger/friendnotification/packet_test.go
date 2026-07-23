package friendnotification

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodePreservesRendererShape verifies the string, integer, string wire order.
func TestEncodePreservesRendererShape(t *testing.T) {
	packet, err := Encode("42", TypeAchievementCompleted, "ACH_Test1")
	if err != nil {
		t.Fatal(err)
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.StringField, codec.Int32Field, codec.StringField})
	if err != nil {
		t.Fatal(err)
	}
	if values[0].String != "42" || values[1].Int32 != TypeAchievementCompleted || values[2].String != "ACH_Test1" {
		t.Fatalf("unexpected notification: %#v", values)
	}
}
