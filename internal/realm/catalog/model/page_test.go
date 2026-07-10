package model

import "testing"

// TestPageAccessibleAppliesVisibilityGates verifies page access policy.
func TestPageAccessibleAppliesVisibilityGates(t *testing.T) {
	page := Page{Visible: true, Enabled: true, MinRank: 2, ClubOnly: true}
	if page.Accessible(1, true) || page.Accessible(2, false) {
		t.Fatal("expected rank and club gates")
	}
	if !page.Accessible(2, true) {
		t.Fatal("expected eligible page access")
	}
}
