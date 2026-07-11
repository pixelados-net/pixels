package votes

import (
	"testing"

	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
)

// BenchmarkScorePackets measures reusable score packet construction.
func BenchmarkScorePackets(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		_, _, _ = scorePackets(42)
	}
}

// BenchmarkOccupantIDs measures batch voter query input construction.
func BenchmarkOccupantIDs(b *testing.B) {
	occupants := make([]roomlive.Occupant, 100)
	for index := range occupants {
		occupants[index].PlayerID = int64(index + 1)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_ = occupantIDs(occupants)
	}
}
