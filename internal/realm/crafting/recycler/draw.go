// Package recycler owns exact-batch recycling and prize selection.
package recycler

import (
	"math/rand"

	craftingrecord "github.com/niflaot/pixels/internal/realm/crafting/record"
)

// RNG provides bounded deterministic random values.
type RNG interface{ Intn(int) int }

// Random creates a production pseudo-random source.
func Random() RNG { return rand.New(rand.NewSource(rand.Int63())) }

// Draw chooses the first successful rarity tier and a uniform prize within it.
func Draw(prizes []craftingrecord.Prize, chances map[int32]int, rng RNG) (craftingrecord.Prize, bool) {
	for tier := int32(5); tier >= 2; tier-- {
		denominator := chances[tier]
		count := tierCount(prizes, tier)
		if count > 0 && denominator > 0 && rng.Intn(denominator) == 0 {
			return tierPrize(prizes, tier, rng.Intn(count))
		}
	}
	count := tierCount(prizes, 1)
	if count == 0 {
		return craftingrecord.Prize{}, false
	}
	return tierPrize(prizes, 1, rng.Intn(count))
}

func tierCount(prizes []craftingrecord.Prize, tier int32) int {
	count := 0
	for _, prize := range prizes {
		if prize.Tier == tier {
			count++
		}
	}
	return count
}
func tierPrize(prizes []craftingrecord.Prize, tier int32, index int) (craftingrecord.Prize, bool) {
	for _, prize := range prizes {
		if prize.Tier != tier {
			continue
		}
		if index == 0 {
			return prize, true
		}
		index--
	}
	return craftingrecord.Prize{}, false
}
