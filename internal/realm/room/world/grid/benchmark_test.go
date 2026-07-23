package grid

import (
	"strings"
	"testing"
)

// BenchmarkParseHeightmapSmall measures parser overhead for a small room.
func BenchmarkParseHeightmapSmall(b *testing.B) {
	heightmap := benchmarkHeightmap(8, 8)
	for range b.N {
		if _, err := Parse(heightmap); err != nil {
			b.Fatalf("parse heightmap: %v", err)
		}
	}
}

// BenchmarkParseHeightmapLarge measures parser overhead for a large room.
func BenchmarkParseHeightmapLarge(b *testing.B) {
	heightmap := benchmarkHeightmap(96, 96)
	for range b.N {
		if _, err := Parse(heightmap); err != nil {
			b.Fatalf("parse heightmap: %v", err)
		}
	}
}

// BenchmarkEncodeHeightmapLarge measures encoder overhead for a large room.
func BenchmarkEncodeHeightmapLarge(b *testing.B) {
	roomGrid, err := Parse(benchmarkHeightmap(96, 96))
	if err != nil {
		b.Fatalf("parse heightmap: %v", err)
	}

	b.ResetTimer()
	for range b.N {
		if _, err := roomGrid.Encode(); err != nil {
			b.Fatalf("encode heightmap: %v", err)
		}
	}
}

// benchmarkHeightmap creates a deterministic benchmark heightmap.
func benchmarkHeightmap(width int, height int) string {
	var builder strings.Builder
	builder.Grow(width*height + height - 1)
	for y := 0; y < height; y++ {
		if y > 0 {
			builder.WriteByte('\r')
		}
		for x := 0; x < width; x++ {
			if x == 0 || y == 0 || x == width-1 || y == height-1 {
				builder.WriteByte('x')
				continue
			}
			builder.WriteByte(byte('0' + (x+y)%10))
		}
	}

	return builder.String()
}
