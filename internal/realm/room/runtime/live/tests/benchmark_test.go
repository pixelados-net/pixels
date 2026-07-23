package live_test

import (
	"strings"
	"testing"
)

// BenchmarkRoomMoveTo measures runtime path assignment cost.
func BenchmarkRoomMoveTo(b *testing.B) {
	room := worldRoomForTest(b, heightmapForBenchmark(24), 0, 0)
	if _, err := room.Join(occupantForTest(7)); err != nil {
		b.Fatalf("join room: %v", err)
	}
	goal := pointForTest(b, 23, 23)
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		if _, err := room.MoveTo(7, goal); err != nil {
			b.Fatalf("move unit: %v", err)
		}
	}
}

// BenchmarkRoomTickManyUnits measures one idle tick over many units.
func BenchmarkRoomTickManyUnits(b *testing.B) {
	room := worldRoomForTest(b, heightmapForBenchmark(16), 0, 0)
	for playerID := int64(1); playerID <= 64; playerID++ {
		if _, err := room.Join(occupantForTest(playerID)); err != nil {
			b.Fatalf("join room: %v", err)
		}
	}
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		_ = room.Tick()
	}
}

// heightmapForBenchmark creates a square flat heightmap.
func heightmapForBenchmark(size int) string {
	row := strings.Repeat("0", size)
	rows := make([]string, size)
	for index := range rows {
		rows[index] = row
	}

	return strings.Join(rows, "\r")
}
