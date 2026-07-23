package model

import "testing"

// TestQueryNormalizeBoundsPageSize verifies history pagination limits.
func TestQueryNormalizeBoundsPageSize(t *testing.T) {
	if got := (Query{}).Normalize().Limit; got != 50 {
		t.Fatalf("expected default 50, got %d", got)
	}
	if got := (Query{Limit: 500}).Normalize().Limit; got != 200 {
		t.Fatalf("expected maximum 200, got %d", got)
	}
}
