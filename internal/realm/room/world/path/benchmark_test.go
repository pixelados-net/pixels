package path

import (
	"strings"
	"testing"

	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	"github.com/niflaot/pixels/internal/realm/room/world/surface"
)

// BenchmarkFindOpenRoom measures A* over an open room.
func BenchmarkFindOpenRoom(b *testing.B) {
	finder := benchmarkFinder(b, benchmarkHeightmap(32, 32, false), nil)
	start := Position{Point: grid.MustPoint(1, 1), Z: 0}
	goal := grid.MustPoint(30, 30)

	for range b.N {
		if _, err := finder.Find(start, goal); err != nil {
			b.Fatalf("find path: %v", err)
		}
	}
}

// BenchmarkFindObstacleRoom measures A* around repeated obstacles.
func BenchmarkFindObstacleRoom(b *testing.B) {
	finder := benchmarkFinder(b, benchmarkHeightmap(32, 32, true), nil)
	start := Position{Point: grid.MustPoint(1, 1), Z: 0}
	goal := grid.MustPoint(29, 30)

	for range b.N {
		if _, err := finder.Find(start, goal); err != nil {
			b.Fatalf("find path: %v", err)
		}
	}
}

// BenchmarkFindDynamicSections measures A* with dynamic columns.
func BenchmarkFindDynamicSections(b *testing.B) {
	fixtures := make([]surface.Fixture, 0, 64)
	for y := 2; y < 30; y += 4 {
		for x := 2; x < 30; x += 4 {
			fixture, err := surface.NewFixture(surface.FixtureParams{
				Point: grid.MustPoint(x, y),
				Z:     1,
				Top:   1,
				State: surface.StateOpen,
			})
			if err != nil {
				b.Fatalf("create fixture: %v", err)
			}
			fixtures = append(fixtures, fixture)
		}
	}
	finder := benchmarkFinder(b, benchmarkHeightmap(32, 32, false), fixtures)
	start := Position{Point: grid.MustPoint(1, 1), Z: 0}
	goal := grid.MustPoint(30, 30)

	for range b.N {
		if _, err := finder.Find(start, goal); err != nil {
			b.Fatalf("find path: %v", err)
		}
	}
}

// BenchmarkFindOccupiedRoom measures A* with many occupied positions.
func BenchmarkFindOccupiedRoom(b *testing.B) {
	resolver := benchmarkResolver(b, benchmarkHeightmap(32, 32, false), nil)
	occupancy := NewOccupancy(benchmarkOccupiedPositions())
	finder := NewFinderWithOccupancy(resolver, DefaultRules(), occupancy)
	start := Position{Point: grid.MustPoint(1, 1), Z: 0}
	goal := grid.MustPoint(30, 30)

	for range b.N {
		if _, err := finder.Find(start, goal); err != nil {
			b.Fatalf("find path: %v", err)
		}
	}
}

// benchmarkFinder creates a finder for benchmarks.
func benchmarkFinder(b *testing.B, heightmap string, fixtures []surface.Fixture) *Finder {
	b.Helper()

	return NewFinder(benchmarkResolver(b, heightmap, fixtures), DefaultRules())
}

// benchmarkResolver creates a resolver for benchmarks.
func benchmarkResolver(b *testing.B, heightmap string, fixtures []surface.Fixture) *surface.Resolver {
	b.Helper()

	roomGrid, err := grid.Parse(heightmap)
	if err != nil {
		b.Fatalf("parse grid: %v", err)
	}
	resolver, err := surface.NewResolver(roomGrid, fixtures)
	if err != nil {
		b.Fatalf("create resolver: %v", err)
	}

	return resolver
}

// benchmarkHeightmap creates a deterministic benchmark heightmap.
func benchmarkHeightmap(width int, height int, obstacles bool) string {
	var builder strings.Builder
	builder.Grow(width*height + height - 1)
	for y := 0; y < height; y++ {
		if y > 0 {
			builder.WriteByte('\r')
		}
		for x := 0; x < width; x++ {
			builder.WriteByte(benchmarkTile(x, y, width, height, obstacles))
		}
	}

	return builder.String()
}

// benchmarkTile returns one benchmark heightmap tile.
func benchmarkTile(x int, y int, width int, height int, obstacles bool) byte {
	if x == 0 || y == 0 || x == width-1 || y == height-1 {
		return 'x'
	}
	if obstacles && x%5 == 0 && y%7 != 0 {
		return 'x'
	}

	return '0'
}

// benchmarkOccupiedPositions creates deterministic occupied positions.
func benchmarkOccupiedPositions() []Position {
	positions := make([]Position, 0, 84)
	for y := 2; y < 30; y += 3 {
		for x := 2; x < 30; x += 4 {
			if x == y {
				continue
			}
			positions = append(positions, Position{Point: grid.MustPoint(x, y), Z: 0})
		}
	}

	return positions
}
