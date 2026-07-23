package recycler

import (
	craftingrecord "github.com/niflaot/pixels/internal/realm/crafting/record"
	"testing"
)

type fixedRNG struct {
	values []int
	index  int
}

func (rng *fixedRNG) Intn(limit int) int {
	value := rng.values[rng.index%len(rng.values)] % limit
	rng.index++
	return value
}

// TestDrawChecksHighTiersThenGuaranteedFloor verifies independent tier order and fallback.
func TestDrawChecksHighTiersThenGuaranteedFloor(t *testing.T) {
	prizes := []craftingrecord.Prize{{Tier: 1, RewardDefinitionID: 1}, {Tier: 3, RewardDefinitionID: 3}, {Tier: 5, RewardDefinitionID: 5}}
	selected, found := Draw(prizes, map[int32]int{5: 2, 3: 2}, &fixedRNG{values: []int{1, 0, 0}})
	if !found || selected.Tier != 3 {
		t.Fatalf("selected %+v", selected)
	}
	selected, found = Draw(prizes, map[int32]int{5: 2, 3: 2}, &fixedRNG{values: []int{1, 1, 0}})
	if !found || selected.Tier != 1 {
		t.Fatalf("fallback %+v", selected)
	}
}

// TestDrawRejectsMissingFloor verifies invalid prize configuration.
func TestDrawRejectsMissingFloor(t *testing.T) {
	_, found := Draw(nil, map[int32]int{}, &fixedRNG{values: []int{0}})
	if found {
		t.Fatal("expected missing prize rejection")
	}
}

// BenchmarkRecyclerPrizeDraw measures the zero-allocation draw hot path.
func BenchmarkRecyclerPrizeDraw(b *testing.B) {
	prizes := []craftingrecord.Prize{{Tier: 1, RewardDefinitionID: 1}, {Tier: 2, RewardDefinitionID: 2}, {Tier: 3, RewardDefinitionID: 3}}
	chances := map[int32]int{3: 20, 2: 5}
	rng := &fixedRNG{values: []int{1}}
	b.ReportAllocs()
	for b.Loop() {
		_, _ = Draw(prizes, chances, rng)
	}
}
