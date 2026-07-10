package model

import "testing"

// TestItemPriceAndLimitedHelpers verifies offer classification helpers.
func TestItemPriceAndLimitedHelpers(t *testing.T) {
	item := Item{PointsType: CreditsType, LimitedStack: 10}
	if !item.IsCredits() || !item.IsLimited() {
		t.Fatalf("unexpected item classification %#v", item)
	}

	item.PointsType = 5
	item.LimitedStack = 0
	if item.IsCredits() || item.IsLimited() {
		t.Fatalf("unexpected points classification %#v", item)
	}
}
