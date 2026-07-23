package model

import "testing"

// TestPageAccessibleAppliesVisibilityGates verifies page access policy.
func TestPageAccessibleAppliesVisibilityGates(t *testing.T) {
	page := Page{Visible: true, Enabled: true, ClubOnly: true}
	if page.Accessible(false) {
		t.Fatal("expected club gate")
	}
	if !page.Accessible(true) {
		t.Fatal("expected eligible page access")
	}
}
