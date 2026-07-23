package club

import (
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/realm/subscription/record"
)

// TestStatusProjectionUsesNitroClubProduct verifies VIP state is not filtered by the client purse.
func TestStatusProjectionUsesNitroClubProduct(t *testing.T) {
	now := time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC)
	expires := now.Add(31 * 24 * time.Hour)
	state := statusProjection(record.Membership{Level: record.LevelVIP, ExpiresAt: &expires,
		LifetimeActiveSeconds: 40 * 86_400, LifetimeVIPSeconds: 12 * 86_400}, 2, now, "")
	if state.ProductName != "habbo_club" || !state.VIP || state.DaysToPeriodEnd != 31 ||
		state.PastClubDays != 40 || state.PastVIPDays != 12 {
		t.Fatalf("state=%#v", state)
	}
}
