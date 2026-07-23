package data

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestAppendPreservesSixteenFieldOrder verifies Nitro's exact nested quest record.
func TestAppendPreservesSixteenFieldOrder(t *testing.T) {
	payload, err := Append(nil, Quest{CampaignCode: "explore", CompletedInCampaign: 1, CampaignCount: 4, RewardCurrencyType: 0, ID: 9, Accepted: true, Type: "room.entered", ImageVersion: "1", RewardAmount: 10, LocalizationCode: "explore.enter", CompletedSteps: 2, TotalSteps: 3, SortOrder: 5, CatalogPage: "", ChainCode: "chain", Easy: true})
	if err != nil {
		t.Fatalf("append quest: %v", err)
	}
	definition := codec.Definition{codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.BooleanField, codec.StringField, codec.StringField, codec.Int32Field, codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.StringField, codec.StringField, codec.BooleanField}
	values, rest, err := codec.DecodePayload(nil, definition, payload)
	if err != nil || len(rest) != 0 {
		t.Fatalf("decode quest: %v rest=%d", err, len(rest))
	}
	if len(values) != 16 || values[0].String != "explore" || values[4].Int32 != 9 || !values[5].Boolean || values[11].Int32 != 3 || !values[15].Boolean {
		t.Fatalf("unexpected quest values %#v", values)
	}
}
