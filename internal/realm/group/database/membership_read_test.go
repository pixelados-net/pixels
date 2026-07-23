package database

import (
	"strings"
	"testing"
)

// TestPlayerGroupsQueryNormalizesMissingFavorite verifies optional preferences cannot reach a bool scanner as NULL.
func TestPlayerGroupsQueryNormalizesMissingFavorite(t *testing.T) {
	if !strings.Contains(playerGroupsSQL, "coalesce(preference.favorite_group_id=g.id,false)") {
		t.Fatalf("missing total favorite projection in %q", playerGroupsSQL)
	}
}
