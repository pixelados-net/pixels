package surface

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/internal/realm/room/world/grid"
)

// TestResolverReturnsBaseSection verifies implicit base column resolution.
func TestResolverReturnsBaseSection(t *testing.T) {
	resolver := resolverForTest(t, nil)

	column, err := resolver.Column(grid.MustPoint(1, 1))
	if err != nil {
		t.Fatalf("resolve column: %v", err)
	}

	if column.Dynamic() || column.Version() != 0 {
		t.Fatalf("expected implicit static column")
	}
	if len(column.Sections()) != 1 {
		t.Fatalf("expected one section, got %d", len(column.Sections()))
	}
	assertSection(t, column.Sections()[0], grid.Height(4), StateOpen, SourceBase)
}

// TestResolverAppliesFixtures verifies dynamic section resolution.
func TestResolverAppliesFixtures(t *testing.T) {
	point := grid.MustPoint(1, 1)
	fixture := fixtureForTest(t, FixtureParams{
		Point:    point,
		Z:        7,
		Top:      8,
		State:    StateBlocked,
		Stacking: false,
		SourceID: 42,
	})
	resolver := resolverForTest(t, []Fixture{fixture})

	column, err := resolver.Column(point)
	if err != nil {
		t.Fatalf("resolve column: %v", err)
	}

	if !column.Dynamic() || column.Version() != 1 {
		t.Fatalf("expected dynamic version 1, got %d", column.Version())
	}
	if len(column.Sections()) != 2 {
		t.Fatalf("expected two sections, got %d", len(column.Sections()))
	}
	assertSection(t, column.Sections()[0], grid.Height(4), StateOpen, SourceBase)
	assertSection(t, column.Sections()[1], grid.Height(7), StateBlocked, SourceFixture)
}

// TestResolverFindsSections verifies exact and top section lookup.
func TestResolverFindsSections(t *testing.T) {
	point := grid.MustPoint(1, 1)
	resolver := resolverForTest(t, []Fixture{
		fixtureForTest(t, FixtureParams{Point: point, Z: 9, Top: 10, State: StateSit, Source: SourceStack}),
	})

	top, err := resolver.TopSection(point)
	if err != nil {
		t.Fatalf("resolve top section: %v", err)
	}
	assertSection(t, top, grid.Height(9), StateSit, SourceStack)

	base, err := resolver.SectionAt(point, 4)
	if err != nil {
		t.Fatalf("resolve base section: %v", err)
	}
	assertSection(t, base, grid.Height(4), StateOpen, SourceBase)
}

// TestResolverReportsMissingSections verifies lookup failures.
func TestResolverReportsMissingSections(t *testing.T) {
	resolver := resolverForTest(t, nil)

	_, err := resolver.SectionAt(grid.MustPoint(1, 1), 9)
	if !errors.Is(err, ErrNoSection) {
		t.Fatalf("expected missing section, got %v", err)
	}

	_, err = resolver.Column(grid.MustPoint(0, 0))
	if !errors.Is(err, ErrInvalidTile) {
		t.Fatalf("expected invalid tile, got %v", err)
	}
}

// TestResolverRejectsInvalidFixtureTile verifies fixture tile validation.
func TestResolverRejectsInvalidFixtureTile(t *testing.T) {
	fixture := fixtureForTest(t, FixtureParams{
		Point: grid.MustPoint(0, 0),
		Z:     1,
		Top:   1,
		State: StateOpen,
	})

	_, err := NewResolver(gridForTest(t), []Fixture{fixture})
	if !errors.Is(err, ErrInvalidTile) {
		t.Fatalf("expected invalid tile, got %v", err)
	}
}

// TestResolverAddFixtureBumpsVersion verifies dynamic version updates.
func TestResolverAddFixtureBumpsVersion(t *testing.T) {
	point := grid.MustPoint(1, 1)
	resolver := resolverForTest(t, nil)

	first := fixtureForTest(t, FixtureParams{Point: point, Z: 6, Top: 6, State: StateOpen})
	second := fixtureForTest(t, FixtureParams{Point: point, Z: 7, Top: 7, State: StateOpen})
	if err := resolver.AddFixture(first); err != nil {
		t.Fatalf("add first fixture: %v", err)
	}
	if err := resolver.AddFixture(second); err != nil {
		t.Fatalf("add second fixture: %v", err)
	}

	column, err := resolver.Column(point)
	if err != nil {
		t.Fatalf("resolve column: %v", err)
	}
	if column.Version() != 2 || len(column.Sections()) != 3 {
		t.Fatalf("unexpected column version=%d sections=%d", column.Version(), len(column.Sections()))
	}
}

// TestResolverRemoveFixturesClearsTileAndBumpsVersion verifies fixture removal.
func TestResolverRemoveFixturesClearsTileAndBumpsVersion(t *testing.T) {
	point := grid.MustPoint(1, 1)
	resolver := resolverForTest(t, []Fixture{
		fixtureForTest(t, FixtureParams{Point: point, Z: 6, Top: 6, State: StateBlocked, SourceID: 42}),
	})

	removed := resolver.RemoveFixtures(42)
	if removed != 1 {
		t.Fatalf("expected one removed fixture, got %d", removed)
	}

	column, err := resolver.Column(point)
	if err != nil {
		t.Fatalf("resolve column: %v", err)
	}
	if column.Version() != 2 {
		t.Fatalf("expected version bumped twice, got %d", column.Version())
	}
	if len(column.Sections()) != 1 {
		t.Fatalf("expected fixture removed, got %d sections", len(column.Sections()))
	}
	assertSection(t, column.Sections()[0], grid.Height(4), StateOpen, SourceBase)
}

// TestResolverRemoveFixturesIgnoresUnknownSource verifies removal is a no-op for unmatched sources.
func TestResolverRemoveFixturesIgnoresUnknownSource(t *testing.T) {
	point := grid.MustPoint(1, 1)
	resolver := resolverForTest(t, []Fixture{
		fixtureForTest(t, FixtureParams{Point: point, Z: 6, Top: 6, State: StateBlocked, SourceID: 42}),
	})

	if removed := resolver.RemoveFixtures(99); removed != 0 {
		t.Fatalf("expected no fixtures removed, got %d", removed)
	}

	column, err := resolver.Column(point)
	if err != nil {
		t.Fatalf("resolve column: %v", err)
	}
	if column.Version() != 1 || len(column.Sections()) != 2 {
		t.Fatalf("expected untouched column, got version=%d sections=%d", column.Version(), len(column.Sections()))
	}
}
