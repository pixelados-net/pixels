package club

import (
	"testing"

	"github.com/niflaot/pixels/internal/realm/subscription/core"
	"github.com/niflaot/pixels/internal/realm/subscription/record"
)

// TestGiftSummaryReturnsEmptyNonMemberState verifies Nitro receives a stable no-club response.
func TestGiftSummaryReturnsEmptyNonMemberState(t *testing.T) {
	days, available := giftSummary(record.Membership{}, false)
	if days != 0 || available != 0 {
		t.Fatalf("days=%d available=%d", days, available)
	}
}

// TestGiftSummaryReturnsEarnedMemberState verifies active gift availability.
func TestGiftSummaryReturnsEarnedMemberState(t *testing.T) {
	membership := record.Membership{Level: record.LevelVIP, LifetimeActiveSeconds: core.ClubGiftPeriodSeconds}
	days, available := giftSummary(membership, true)
	if days != 0 || available != 1 {
		t.Fatalf("days=%d available=%d", days, available)
	}
}
