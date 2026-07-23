package profile

import (
	"testing"

	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeUsesNativeProfileShape verifies USER_PROFILE wire fields.
func TestEncodeUsesNativeProfileShape(t *testing.T) {
	groups := []grouprecord.PlayerGroup{{Group: grouprecord.Group{ID: 4, Name: "Pixels", BadgeCode: "b001010", OwnerPlayerID: 7}, Favorite: true}}
	packet, err := Encode(7, "demo", "look", "motto", "01-01-2026", groups, 3, true, false, true, 10, true)
	values, decodeErr := codec.DecodePacketExact(packet, codec.Definition{
		codec.Int32Field, codec.StringField, codec.StringField, codec.StringField, codec.StringField,
		codec.Int32Field, codec.Int32Field, codec.BooleanField, codec.BooleanField, codec.BooleanField, codec.Int32Field,
		codec.Int32Field, codec.StringField, codec.StringField, codec.StringField, codec.StringField, codec.BooleanField, codec.Int32Field, codec.BooleanField,
		codec.Int32Field, codec.BooleanField,
	})
	if err != nil || decodeErr != nil || values[5].Int32 != 0 || values[6].Int32 != 3 || !values[7].Boolean || values[10].Int32 != 1 || values[11].Int32 != 4 || values[19].Int32 != 10 || !values[20].Boolean {
		t.Fatalf("unexpected values=%#v err=%v decode=%v", values, err, decodeErr)
	}
}
