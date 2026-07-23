package furniture

import (
	"testing"

	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// benchmarkGateOpenResult prevents the gate-state result from being optimized away.
var benchmarkGateOpenResult bool

// BenchmarkFixturesManyPlacedItems measures fixture building cost across many placed items.
func BenchmarkFixturesManyPlacedItems(b *testing.B) {
	items := make([]Item, 0, 1000)
	for index := int64(0); index < 1000; index++ {
		items = append(items, Item{
			ID:       index,
			Point:    grid.MustPoint(int(index%200), int(index/200)),
			Z:        0,
			Rotation: worldunit.Rotation((index % 4) * 2),
			Definition: Definition{
				Width: 2, Length: 1, StackHeight: 1, AllowSit: true,
				Slots: []SlotDefinition{
					{DX: 0, DY: 0, Status: SlotStatusSit, BodyRotation: worldunit.RotationSouth},
					{DX: 1, DY: 0, Status: SlotStatusSit, BodyRotation: worldunit.RotationSouth},
				},
			},
		})
	}

	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		for _, item := range items {
			if _, err := Fixtures(item); err != nil {
				b.Fatalf("build fixtures: %v", err)
			}
		}
	}
}

// BenchmarkGateOpenReversedState measures the configured gate-state hot path.
func BenchmarkGateOpenReversedState(b *testing.B) {
	item := Item{
		ExtraData: "0",
		Definition: Definition{
			InteractionType: "gate",
			CustomParams:    "open_state=0",
		},
	}

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		benchmarkGateOpenResult = gateOpen(item)
	}
}
