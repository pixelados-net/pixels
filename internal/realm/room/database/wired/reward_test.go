package wired

import (
	"testing"

	roomwired "github.com/niflaot/pixels/internal/realm/room/world/wired/record"
)

// TestChooseRewardHandlesWeightsStockExclusionAndLoss verifies deterministic selection policy.
func TestChooseRewardHandlesWeightsStockExclusionAndLoss(t *testing.T) {
	zero, one := int64(0), int64(1)
	rewards := []roomwired.Reward{
		{ID: 1, Weight: 10, Stock: &zero},
		{ID: 2, Weight: 20, Stock: &one},
		{ID: 3, Weight: 30},
	}
	selected, status := chooseReward(rewards, nil, 25)
	if status != roomwired.ClaimDelivered || selected.ID != 3 {
		t.Fatalf("selected=%+v status=%d", selected, status)
	}
	if _, status = chooseReward(rewards, nil, 75); status != roomwired.ClaimMissed {
		t.Fatalf("loss status=%d", status)
	}
	if _, status = chooseReward(rewards, map[int64]struct{}{2: {}, 3: {}}, 0); status != roomwired.ClaimOutOfStock {
		t.Fatalf("excluded status=%d", status)
	}
}
