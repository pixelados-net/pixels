package toggle

import (
	"testing"

	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
)

// TestNextCyclesDeclaredStates verifies normal and defensive state transitions.
func TestNextCyclesDeclaredStates(t *testing.T) {
	cases := []struct {
		name     string
		current  string
		modes    int
		expected string
		commit   bool
	}{
		{name: "first", current: "0", modes: 3, expected: "1", commit: true},
		{name: "second", current: "1", modes: 3, expected: "2", commit: true},
		{name: "wrap", current: "2", modes: 3, expected: "0", commit: true},
		{name: "corrupt", current: "abc", modes: 3, expected: "1", commit: true},
		{name: "out of range", current: "9", modes: 3, expected: "1", commit: true},
		{name: "inert", current: "0", modes: 1, expected: "0", commit: false},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			item := worldfurniture.Item{ExtraData: test.current, Definition: worldfurniture.Definition{InteractionModesCount: test.modes}}
			next, rebuild, commit := (Behavior{}).Next(nil, item)
			if next != test.expected || rebuild || commit != test.commit {
				t.Fatalf("unexpected transition next=%q rebuild=%v commit=%v", next, rebuild, commit)
			}
		})
	}
}

// BenchmarkNext measures the generic state transition calculation.
func BenchmarkNext(b *testing.B) {
	item := worldfurniture.Item{ExtraData: "1", Definition: worldfurniture.Definition{InteractionModesCount: 3}}
	b.ReportAllocs()
	for b.Loop() {
		_, _, _ = (Behavior{}).Next(nil, item)
	}
}
