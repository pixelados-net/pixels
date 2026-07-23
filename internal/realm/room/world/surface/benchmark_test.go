package surface

import (
	"testing"

	"github.com/niflaot/pixels/internal/realm/room/world/grid"
)

// BenchmarkResolveBaseColumn measures implicit base column resolution.
func BenchmarkResolveBaseColumn(b *testing.B) {
	resolver := resolverForBenchmark(b, nil)
	point := grid.MustPoint(1, 1)

	for range b.N {
		if _, err := resolver.Column(point); err != nil {
			b.Fatalf("resolve column: %v", err)
		}
	}
}

// BenchmarkResolveDynamicColumn measures dynamic fixture column resolution.
func BenchmarkResolveDynamicColumn(b *testing.B) {
	point := grid.MustPoint(1, 1)
	fixture := fixtureForBenchmark(b, FixtureParams{Point: point, Z: 7, Top: 8, State: StateOpen})
	resolver := resolverForBenchmark(b, []Fixture{fixture})

	for range b.N {
		if _, err := resolver.Column(point); err != nil {
			b.Fatalf("resolve column: %v", err)
		}
	}
}

// resolverForBenchmark creates a surface resolver for benchmarks.
func resolverForBenchmark(b *testing.B, fixtures []Fixture) *Resolver {
	b.Helper()

	roomGrid, err := grid.Parse("xxx\rx4x\rxxx")
	if err != nil {
		b.Fatalf("parse grid: %v", err)
	}
	resolver, err := NewResolver(roomGrid, fixtures)
	if err != nil {
		b.Fatalf("create resolver: %v", err)
	}

	return resolver
}

// fixtureForBenchmark creates a fixture for benchmarks.
func fixtureForBenchmark(b *testing.B, params FixtureParams) Fixture {
	b.Helper()

	fixture, err := NewFixture(params)
	if err != nil {
		b.Fatalf("create fixture: %v", err)
	}

	return fixture
}
