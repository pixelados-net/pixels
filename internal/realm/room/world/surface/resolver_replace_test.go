package surface

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/internal/realm/room/world/grid"
)

// TestResolverReplaceFixturesMovesFootprint verifies atomic replacement across tiles.
func TestResolverReplaceFixturesMovesFootprint(t *testing.T) {
	oldPoint := grid.MustPoint(1, 1)
	newPoint := grid.MustPoint(2, 1)
	wideGrid, err := grid.Parse("xxxx\rx44x\rxxxx")
	if err != nil {
		t.Fatalf("parse grid: %v", err)
	}
	resolver, err := NewResolver(wideGrid, []Fixture{
		fixtureForTest(t, FixtureParams{Point: oldPoint, Z: 6, Top: 6, State: StateBlocked, SourceID: 7}),
	})
	if err != nil {
		t.Fatalf("create resolver: %v", err)
	}

	if err := resolver.ReplaceFixtures(7, []Fixture{
		fixtureForTest(t, FixtureParams{Point: newPoint, Z: 6, Top: 6, State: StateBlocked, SourceID: 7}),
	}); err != nil {
		t.Fatalf("replace fixtures: %v", err)
	}

	oldColumn, err := resolver.Column(oldPoint)
	if err != nil {
		t.Fatalf("resolve old column: %v", err)
	}
	assertSection(t, oldColumn.Sections()[0], grid.Height(4), StateOpen, SourceBase)

	newColumn, err := resolver.Column(newPoint)
	if err != nil {
		t.Fatalf("resolve new column: %v", err)
	}
	if len(newColumn.Sections()) != 2 {
		t.Fatalf("expected fixture moved to new tile, got %d sections", len(newColumn.Sections()))
	}
	assertSection(t, newColumn.Sections()[1], grid.Height(6), StateBlocked, SourceFixture)
}

// TestResolverReplaceFixturesRejectsInvalidTile verifies replacement validation.
func TestResolverReplaceFixturesRejectsInvalidTile(t *testing.T) {
	resolver := resolverForTest(t, nil)

	invalid := fixtureForTest(t, FixtureParams{Point: grid.MustPoint(0, 0), Z: 1, Top: 1, State: StateOpen})
	err := resolver.ReplaceFixtures(7, []Fixture{invalid})
	if !errors.Is(err, ErrInvalidTile) {
		t.Fatalf("expected invalid tile, got %v", err)
	}
}

// assertSection verifies a resolved section.
func assertSection(t *testing.T, section Section, height grid.Height, state State, source Source) {
	t.Helper()

	if section.Z() != height || section.State() != state || section.Source() != source {
		t.Fatalf("unexpected section height=%d state=%d source=%d", section.Z(), section.State(), section.Source())
	}
}

// resolverForTest creates a surface resolver for tests.
func resolverForTest(t *testing.T, fixtures []Fixture) *Resolver {
	t.Helper()

	resolver, err := NewResolver(gridForTest(t), fixtures)
	if err != nil {
		t.Fatalf("create resolver: %v", err)
	}

	return resolver
}

// gridForTest creates a room grid for tests.
func gridForTest(t *testing.T) grid.Grid {
	t.Helper()

	roomGrid, err := grid.Parse("xxx\rx4x\rxxx")
	if err != nil {
		t.Fatalf("parse grid: %v", err)
	}

	return roomGrid
}

// fixtureForTest creates a fixture for tests.
func fixtureForTest(t *testing.T, params FixtureParams) Fixture {
	t.Helper()

	fixture, err := NewFixture(params)
	if err != nil {
		t.Fatalf("create fixture: %v", err)
	}

	return fixture
}
