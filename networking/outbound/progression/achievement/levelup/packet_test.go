package levelup

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodePreservesTwelveFieldOrder verifies first-level and replacement badge data.
func TestEncodePreservesTwelveFieldOrder(t *testing.T) {
	packet, err := Encode(Data{Type: 7, Level: 2, BadgeID: 11, BadgeCode: "ACH_Test2", Points: 20, RewardPoints: 5, RewardPointType: 0, BonusPoints: 10, AchievementID: 7, RemovedBadgeCode: "ACH_Test1", Category: "explore", ShowDialog: true})
	if err != nil || packet.Header != Header {
		t.Fatalf("encode level up: %#v %v", packet, err)
	}
	definition := codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.StringField, codec.StringField, codec.BooleanField}
	values, rest, err := codec.DecodePayload(nil, definition, packet.Payload)
	if err != nil || len(rest) != 0 || len(values) != 12 || values[3].String != "ACH_Test2" || values[9].String != "ACH_Test1" || !values[11].Boolean {
		t.Fatalf("unexpected values %#v rest=%d error=%v", values, len(rest), err)
	}
}
