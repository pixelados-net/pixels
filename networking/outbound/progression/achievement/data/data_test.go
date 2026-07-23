package data

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestAppendPreservesThirteenFieldOrder verifies Nitro's exact nested achievement record.
func TestAppendPreservesThirteenFieldOrder(t *testing.T) {
	payload, err := Append(nil, Achievement{ID: 7, Level: 3, BadgeCode: "ACH_Test3", ScoreAtStart: 20, ScoreLimit: 50, LevelRewardPoints: 5, RewardPointType: 0, CurrentPoints: 24, FinalLevel: false, Category: "explore", Subcategory: "rooms", LevelCount: 10, DisplayMethod: 1})
	if err != nil {
		t.Fatalf("append achievement: %v", err)
	}
	values, rest, err := codec.DecodePayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field, codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.BooleanField, codec.StringField, codec.StringField, codec.Int32Field, codec.Int32Field}, payload)
	if err != nil || len(rest) != 0 {
		t.Fatalf("decode achievement: %v rest=%d", err, len(rest))
	}
	if len(values) != 13 || values[0].Int32 != 7 || values[2].String != "ACH_Test3" || values[7].Int32 != 24 || values[9].String != "explore" || values[12].Int32 != 1 {
		t.Fatalf("unexpected achievement values %#v", values)
	}
}
