package freeze

// DropState returns Nitro's native visible state for one broken block reward.
func DropState(power PowerUp) int {
	if power < RangeUp || power > Shield {
		return 1000
	}
	return int(power) * 1000
}

// CollectedState returns Nitro's native pickup animation for one block reward.
func CollectedState(power PowerUp) int { return (int(power) + 10) * 1000 }

// Drop deterministically decides whether one broken block contains a reward.
func Drop(blockID int64, playerID int64, chance int) (PowerUp, bool) {
	if chance <= 0 {
		return 0, false
	}
	value := uint64(blockID)*0x9e3779b185ebca87 ^ uint64(playerID)*0xc2b2ae3d27d4eb4f
	value ^= value >> 33
	value *= 0xff51afd7ed558ccd
	value ^= value >> 33
	if chance < 100 && int(value%100) >= chance {
		return 0, false
	}
	return PowerUp(2 + (value/100)%6), true
}
