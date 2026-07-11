package model

import (
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/permission"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// TestClubLevelAtAppliesExpiration verifies active and expired club tiers.
func TestClubLevelAtAppliesExpiration(t *testing.T) {
	now := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	future := now.Add(time.Hour)
	past := now.Add(-time.Second)
	tests := []struct {
		// name stores the case name.
		name string
		// club stores the entitlement fixture.
		club Club
		// expected stores the active tier.
		expected ClubLevel
	}{
		{name: "vip", club: Club{Level: ClubLevelVIP, ExpiresAt: &future}, expected: ClubLevelVIP},
		{name: "expired", club: Club{Level: ClubLevelHC, ExpiresAt: &past}, expected: ClubLevelNone},
		{name: "missing expiration", club: Club{Level: ClubLevelHC}, expected: ClubLevelNone},
	}
	for _, test := range tests {
		if level := test.club.LevelAt(now); level != test.expected {
			t.Fatalf("%s: expected level %d, got %d", test.name, test.expected, level)
		}
	}
	if !(Club{Level: ClubLevelVIP, ExpiresAt: &future}).ActiveAt(now) {
		t.Fatal("expected active club")
	}
	player := Player{Base: sharedBaseForClubTest(7)}
	if player.HolderID() != 7 || player.HolderKind() != permission.HolderPlayer {
		t.Fatalf("unexpected permission holder id=%d kind=%s", player.HolderID(), player.HolderKind())
	}
}

// sharedBaseForClubTest creates a player base fixture.
func sharedBaseForClubTest(id int64) sharedmodel.Base {
	return sharedmodel.Base{Identity: sharedmodel.Identity{ID: id}}
}

// BenchmarkClubLevelAt measures the in-memory entitlement gate.
func BenchmarkClubLevelAt(b *testing.B) {
	now := time.Now()
	expiresAt := now.Add(time.Hour)
	club := Club{Level: ClubLevelVIP, ExpiresAt: &expiresAt}
	b.ReportAllocs()
	for b.Loop() {
		if club.LevelAt(now) != ClubLevelVIP {
			b.Fatal("expected active club")
		}
	}
}
