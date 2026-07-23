package record

import (
	"testing"
	"time"
)

// TestLevelForExperience verifies threshold boundaries and cap.
func TestLevelForExperience(t *testing.T) {
	cases := []struct{ experience, expected int32 }{{0, 1}, {99, 1}, {100, 2}, {51900, 20}, {1_000_000, 20}}
	for _, test := range cases {
		if actual := LevelForExperience(test.experience, 20); actual != test.expected {
			t.Fatalf("experience=%d actual=%d expected=%d", test.experience, actual, test.expected)
		}
	}
}

// TestDerivePlantState verifies absolute lifecycle boundaries.
func TestDerivePlantState(t *testing.T) {
	now := time.Unix(100, 0)
	grow, die := now.Add(-time.Second), now.Add(time.Minute)
	pet := Pet{State: StateRoom, GrowAt: &grow, DieAt: &die, CreatedAt: now.Add(-time.Hour)}
	state := pet.DerivePlantState(now, Species{Plant: true})
	if state.GrowthStage != 7 || !state.FullyGrown || state.Dead || !state.CanHarvest || state.RemainingLifeSeconds != 60 {
		t.Fatalf("unexpected state %#v", state)
	}
	grow, die = now.Add(6*time.Hour), now.Add(24*time.Hour)
	pet.CreatedAt, pet.GrowAt, pet.DieAt = now.Add(-6*time.Hour), &grow, &die
	if stage := pet.DerivePlantState(now, Species{Plant: true}).GrowthStage; stage != 4 {
		t.Fatalf("expected midpoint stage 4, got %d", stage)
	}
}

// TestMaterializeStatsUsesWholeIntervals verifies decay, clamping, and remainder preservation.
func TestMaterializeStatsUsesWholeIntervals(t *testing.T) {
	start := time.Unix(100, 0)
	pet := Pet{Level: 2, Energy: 3, Happiness: 5, StatsAt: start}
	materialized := MaterializeStats(pet, start.Add(2*time.Hour+time.Minute), time.Hour, 2, 3)
	if materialized.Energy != 0 || materialized.Happiness != 0 || !materialized.StatsAt.Equal(start.Add(2*time.Hour)) {
		t.Fatalf("unexpected materialized stats %#v", materialized)
	}
	unchanged := MaterializeStats(pet, start.Add(time.Minute), time.Hour, 2, 3)
	if unchanged.Energy != pet.Energy || unchanged.Happiness != pet.Happiness || !unchanged.StatsAt.Equal(pet.StatsAt) {
		t.Fatalf("unexpected partial-interval mutation %#v", unchanged)
	}
}

// BenchmarkStatMaterialize measures allocation-free absolute stat derivation.
func BenchmarkStatMaterialize(b *testing.B) {
	start := time.Unix(100, 0)
	pet := Pet{Level: 5, Energy: 500, Happiness: 100, StatsAt: start}
	now := start.Add(48 * time.Hour)
	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		_ = MaterializeStats(pet, now, 30*time.Minute, 1, 1)
	}
}
