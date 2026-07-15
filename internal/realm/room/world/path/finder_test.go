package path

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	"github.com/niflaot/pixels/internal/realm/room/world/surface"
)

// TestFinderFindsStraightPath verifies a basic path over the grid.
func TestFinderFindsStraightPath(t *testing.T) {
	finder := finderForTest(t, "000", nil, DefaultRules())

	roomPath, err := finder.Find(Position{Point: grid.MustPoint(0, 0), Z: 0}, grid.MustPoint(2, 0))
	if err != nil {
		t.Fatalf("find path: %v", err)
	}

	assertSteps(t, roomPath, []Position{
		{Point: grid.MustPoint(1, 0), Z: 0},
		{Point: grid.MustPoint(2, 0), Z: 0},
	})
}

// TestFinderReportsInvalidStartAndGoal verifies request validation.
func TestFinderReportsInvalidStartAndGoal(t *testing.T) {
	finder := finderForTest(t, "x0", nil, DefaultRules())

	_, err := finder.Find(Position{Point: grid.MustPoint(0, 0), Z: 0}, grid.MustPoint(1, 0))
	if !errors.Is(err, ErrInvalidStart) {
		t.Fatalf("expected invalid start, got %v", err)
	}

	_, err = finder.Find(Position{Point: grid.MustPoint(1, 0), Z: 0}, grid.MustPoint(0, 0))
	if !errors.Is(err, ErrInvalidGoal) {
		t.Fatalf("expected invalid goal, got %v", err)
	}
}

// TestFinderReportsNoPath verifies blocked routes.
func TestFinderReportsNoPath(t *testing.T) {
	finder := finderForTest(t, "0x0", nil, DefaultRules())

	_, err := finder.Find(Position{Point: grid.MustPoint(0, 0), Z: 0}, grid.MustPoint(2, 0))
	if !errors.Is(err, ErrNoPath) {
		t.Fatalf("expected no path, got %v", err)
	}
}

// TestFinderRoutesAroundObstacles verifies integration over a larger grid.
func TestFinderRoutesAroundObstacles(t *testing.T) {
	finder := finderForTest(t, "00000\r0xxx0\r00000\r0xxx0\r00000", nil, DefaultRules())

	roomPath, err := finder.Find(Position{Point: grid.MustPoint(0, 0), Z: 0}, grid.MustPoint(4, 4))
	if err != nil {
		t.Fatalf("find obstacle path: %v", err)
	}
	if roomPath.Len() == 0 {
		t.Fatal("expected non-empty obstacle path")
	}
	for _, step := range roomPath.Steps() {
		if step.Position.Point.X > 0 && step.Position.Point.X < 4 && step.Position.Point.Y == 1 {
			t.Fatalf("path crossed first wall at %#v", step.Position)
		}
		if step.Position.Point.X > 0 && step.Position.Point.X < 4 && step.Position.Point.Y == 3 {
			t.Fatalf("path crossed second wall at %#v", step.Position)
		}
	}
}

// TestFinderHonorsStepUp verifies upward step limits.
func TestFinderHonorsStepUp(t *testing.T) {
	finder := finderForTest(t, "012", nil, DefaultRules())

	roomPath, err := finder.Find(Position{Point: grid.MustPoint(0, 0), Z: 0}, grid.MustPoint(2, 0))
	if err != nil {
		t.Fatalf("find stepped path: %v", err)
	}
	assertSteps(t, roomPath, []Position{
		{Point: grid.MustPoint(1, 0), Z: grid.HeightFromInt(1)},
		{Point: grid.MustPoint(2, 0), Z: grid.HeightFromInt(2)},
	})

	finder = finderForTest(t, "02", nil, DefaultRules())
	_, err = finder.Find(Position{Point: grid.MustPoint(0, 0), Z: 0}, grid.MustPoint(1, 0))
	if !errors.Is(err, ErrNoPath) {
		t.Fatalf("expected no path for tall step, got %v", err)
	}
}

// TestFinderAllowsEscapeFromBlockedStart verifies furniture changes cannot permanently freeze a unit.
func TestFinderAllowsEscapeFromBlockedStart(t *testing.T) {
	start := grid.MustPoint(0, 0)
	fixture := fixtureForTest(t, surface.FixtureParams{Point: start, Z: 0, Bottom: 0, HasBottom: true, Top: grid.HeightFromInt(1), State: surface.StateBlocked})
	resolver := resolverForTest(t, "00", []surface.Fixture{fixture})
	finder := NewFinder(resolver, DefaultRules())
	roomPath, err := finder.Find(Position{Point: start, Z: 0}, grid.MustPoint(1, 0))
	if err != nil {
		t.Fatalf("escape blocked start: %v", err)
	}
	assertSteps(t, roomPath, []Position{{Point: grid.MustPoint(1, 0), Z: 0}})
}

// TestFinderHonorsFalling verifies downward step rules.
func TestFinderHonorsFalling(t *testing.T) {
	finder := finderForTest(t, "20", nil, DefaultRules())

	_, err := finder.Find(Position{Point: grid.MustPoint(0, 0), Z: 2}, grid.MustPoint(1, 0))
	if !errors.Is(err, ErrNoPath) {
		t.Fatalf("expected no path without falling, got %v", err)
	}

	rules := DefaultRules()
	rules.AllowFalling = true
	finder = finderForTest(t, "20", nil, rules)
	roomPath, err := finder.Find(Position{Point: grid.MustPoint(0, 0), Z: 2}, grid.MustPoint(1, 0))
	if err != nil {
		t.Fatalf("find falling path: %v", err)
	}
	assertSteps(t, roomPath, []Position{{Point: grid.MustPoint(1, 0), Z: 0}})
}

// TestFinderHonorsDiagonalCornerRule verifies diagonal blocking.
func TestFinderHonorsDiagonalCornerRule(t *testing.T) {
	finder := finderForTest(t, "0x\rx0", nil, DefaultRules())

	_, err := finder.Find(Position{Point: grid.MustPoint(0, 0), Z: 0}, grid.MustPoint(1, 1))
	if !errors.Is(err, ErrNoPath) {
		t.Fatalf("expected blocked diagonal, got %v", err)
	}

	finder = finderForTest(t, "00\rx0", nil, DefaultRules())
	roomPath, err := finder.Find(Position{Point: grid.MustPoint(0, 0), Z: 0}, grid.MustPoint(1, 1))
	if err != nil {
		t.Fatalf("find diagonal path: %v", err)
	}
	if roomPath.Len() != 1 || !roomPath.Steps()[0].Diagonal {
		t.Fatalf("expected one diagonal step")
	}
}

// TestFinderUsesVerticalSections verifies section-aware starts and falling.
func TestFinderUsesVerticalSections(t *testing.T) {
	point := grid.MustPoint(0, 0)
	fixture := fixtureForTest(t, surface.FixtureParams{Point: point, Z: 2, Top: 2, State: surface.StateOpen})
	rules := DefaultRules()
	rules.AllowFalling = true
	finder := finderForTest(t, "00", []surface.Fixture{fixture}, rules)

	roomPath, err := finder.Find(Position{Point: point, Z: 2}, grid.MustPoint(1, 0))
	if err != nil {
		t.Fatalf("find section path: %v", err)
	}
	assertSteps(t, roomPath, []Position{{Point: grid.MustPoint(1, 0), Z: 0}})
}

// TestFinderValidatesColumnVersions verifies path version validation.
func TestFinderValidatesColumnVersions(t *testing.T) {
	point := grid.MustPoint(1, 0)
	fixture := fixtureForTest(t, surface.FixtureParams{Point: point, Z: 1, Top: 1, State: surface.StateOpen})
	resolver := resolverForTest(t, "000", []surface.Fixture{fixture})
	finder := NewFinder(resolver, DefaultRules())

	roomPath, err := finder.Find(Position{Point: grid.MustPoint(0, 0), Z: 0}, grid.MustPoint(2, 0))
	if err != nil {
		t.Fatalf("find path: %v", err)
	}
	if err := roomPath.Validate(resolver); err != nil {
		t.Fatalf("validate unchanged path: %v", err)
	}

	second := fixtureForTest(t, surface.FixtureParams{Point: point, Z: 2, Top: 2, State: surface.StateOpen})
	if err := resolver.AddFixture(second); err != nil {
		t.Fatalf("add fixture: %v", err)
	}
	if err := roomPath.Validate(resolver); !errors.Is(err, ErrInvalidPath) {
		t.Fatalf("expected invalid path, got %v", err)
	}
}

// TestFinderHonorsSearchLimit verifies visit limit protection.
func TestFinderHonorsSearchLimit(t *testing.T) {
	rules := DefaultRules()
	rules.MaxVisited = 1
	finder := finderForTest(t, "000\r000\r000", nil, rules)

	_, err := finder.Find(Position{Point: grid.MustPoint(0, 0), Z: 0}, grid.MustPoint(2, 2))
	if !errors.Is(err, ErrSearchLimit) {
		t.Fatalf("expected search limit, got %v", err)
	}
}

// TestFinderCanDisableDiagonals verifies cardinal-only movement.
func TestFinderCanDisableDiagonals(t *testing.T) {
	rules := DefaultRules()
	rules.DisableDiagonal = true
	finder := finderForTest(t, "00\r00", nil, rules)

	roomPath, err := finder.Find(Position{Point: grid.MustPoint(0, 0), Z: 0}, grid.MustPoint(1, 1))
	if err != nil {
		t.Fatalf("find cardinal path: %v", err)
	}
	if roomPath.Len() != 2 {
		t.Fatalf("expected two cardinal steps, got %d", roomPath.Len())
	}
	for _, step := range roomPath.Steps() {
		if step.Diagonal {
			t.Fatal("expected cardinal-only path")
		}
	}
}

// assertSteps verifies path step positions.
func assertSteps(t *testing.T, roomPath Path, expected []Position) {
	t.Helper()

	steps := roomPath.Steps()
	if len(steps) != len(expected) {
		t.Fatalf("expected %d steps, got %d", len(expected), len(steps))
	}
	for index, step := range steps {
		if step.Position != expected[index] {
			t.Fatalf("expected step %d %#v, got %#v", index, expected[index], step.Position)
		}
	}
}

// finderForTest creates a finder from heightmap and fixtures.
func finderForTest(t *testing.T, heightmap string, fixtures []surface.Fixture, rules Rules) *Finder {
	t.Helper()

	return NewFinder(resolverForTest(t, heightmap, fixtures), rules)
}

// resolverForTest creates a surface resolver from heightmap and fixtures.
func resolverForTest(t *testing.T, heightmap string, fixtures []surface.Fixture) *surface.Resolver {
	t.Helper()

	roomGrid, err := grid.Parse(heightmap)
	if err != nil {
		t.Fatalf("parse grid: %v", err)
	}
	resolver, err := surface.NewResolver(roomGrid, fixtures)
	if err != nil {
		t.Fatalf("create resolver: %v", err)
	}

	return resolver
}

// fixtureForTest creates a surface fixture for tests.
func fixtureForTest(t *testing.T, params surface.FixtureParams) surface.Fixture {
	t.Helper()

	fixture, err := surface.NewFixture(params)
	if err != nil {
		t.Fatalf("create fixture: %v", err)
	}

	return fixture
}
