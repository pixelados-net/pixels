package path

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	"github.com/niflaot/pixels/internal/realm/room/world/surface"
)

// TestFinderRoutesAroundOccupiedPositions verifies occupied path blocking.
func TestFinderRoutesAroundOccupiedPositions(t *testing.T) {
	resolver := resolverForTest(t, "00000\r00000\r00000", nil)
	occupancy := NewOccupancy([]Position{
		{Point: grid.MustPoint(1, 1), Z: 0},
		{Point: grid.MustPoint(2, 1), Z: 0},
		{Point: grid.MustPoint(3, 1), Z: 0},
	})
	finder := NewFinderWithOccupancy(resolver, DefaultRules(), occupancy)

	roomPath, err := finder.Find(Position{Point: grid.MustPoint(0, 1), Z: 0}, grid.MustPoint(4, 1))
	if err != nil {
		t.Fatalf("find occupied path: %v", err)
	}
	for _, step := range roomPath.Steps() {
		if occupancy.Occupied(step.Position) {
			t.Fatalf("path crossed occupied position %#v", step.Position)
		}
	}
}

// TestFinderOccupancyIsSectionSpecific verifies z-specific occupancy.
func TestFinderOccupancyIsSectionSpecific(t *testing.T) {
	point := grid.MustPoint(1, 0)
	fixture := fixtureForTest(t, surface.FixtureParams{Point: point, Z: 1, Top: 1, State: surface.StateOpen})
	resolver := resolverForTest(t, "00", []surface.Fixture{fixture})
	occupancy := NewOccupancy([]Position{{Point: point, Z: 1}})
	finder := NewFinderWithOccupancy(resolver, DefaultRules(), occupancy)

	roomPath, err := finder.Find(Position{Point: grid.MustPoint(0, 0), Z: 0}, point)
	if err != nil {
		t.Fatalf("find section-specific occupied path: %v", err)
	}
	assertSteps(t, roomPath, []Position{{Point: point, Z: 0}})
	if occupancy.Len() != 1 || occupancy.Empty() {
		t.Fatal("expected one occupied position")
	}
}

// TestFinderSlotSectionsAreDestinationOnly verifies sit/lay tiles accept a unit as the path goal
// but never as a transit tile toward a further goal.
func TestFinderSlotSectionsAreDestinationOnly(t *testing.T) {
	point := grid.MustPoint(1, 0)
	fixture := fixtureForTest(t, surface.FixtureParams{Point: point, Z: 0, Top: 1, State: surface.StateSit})
	resolver := resolverForTest(t, "000", []surface.Fixture{fixture})
	finder := NewFinder(resolver, DefaultRules())

	if _, err := finder.Find(Position{Point: grid.MustPoint(0, 0), Z: 0}, point); err != nil {
		t.Fatalf("expected seat reachable as goal, got %v", err)
	}
	if _, err := finder.Find(Position{Point: grid.MustPoint(0, 0), Z: 0}, grid.MustPoint(2, 0)); !errors.Is(err, ErrNoPath) {
		t.Fatalf("expected no path through the seat, got %v", err)
	}
}

// TestFinderGoalPrefersSitSectionOverBaseFloor verifies a goal resolves to a sit slot above the
// base floor even when the slot's height does not tie with it (e.g. a chair with stack_height > 0).
func TestFinderGoalPrefersSitSectionOverBaseFloor(t *testing.T) {
	point := grid.MustPoint(1, 0)
	fixture := fixtureForTest(t, surface.FixtureParams{Point: point, Z: 1, Top: 1, State: surface.StateSit})
	resolver := resolverForTest(t, "00", []surface.Fixture{fixture})
	finder := NewFinder(resolver, DefaultRules())

	roomPath, err := finder.Find(Position{Point: grid.MustPoint(0, 0), Z: 0}, point)
	if err != nil {
		t.Fatalf("find sit path: %v", err)
	}
	assertSteps(t, roomPath, []Position{{Point: point, Z: 1}})
}

// TestFinderDiagonalBlockedByOccupiedCorners verifies occupancy diagonal safety.
func TestFinderDiagonalBlockedByOccupiedCorners(t *testing.T) {
	resolver := resolverForTest(t, "00\r00", nil)
	occupancy := NewOccupancy([]Position{
		{Point: grid.MustPoint(1, 0), Z: 0},
		{Point: grid.MustPoint(0, 1), Z: 0},
	})
	finder := NewFinderWithOccupancy(resolver, DefaultRules(), occupancy)

	_, err := finder.Find(Position{Point: grid.MustPoint(0, 0), Z: 0}, grid.MustPoint(1, 1))
	if err == nil {
		t.Fatal("expected occupied diagonal to fail")
	}
}

// TestOpenHeapOrdersNodes verifies typed heap ordering.
func TestOpenHeapOrdersNodes(t *testing.T) {
	heap := openHeap{}
	heap.Push(openNode{priority: 20, cost: 20})
	heap.Push(openNode{priority: 10, cost: 30})
	heap.Push(openNode{priority: 10, cost: 10})

	first, ok := heap.Pop()
	if !ok || first.priority != 10 || first.cost != 10 {
		t.Fatalf("unexpected first node %#v found=%v", first, ok)
	}
	second, ok := heap.Pop()
	if !ok || second.priority != 10 || second.cost != 30 {
		t.Fatalf("unexpected second node %#v found=%v", second, ok)
	}
	third, ok := heap.Pop()
	if !ok || third.priority != 20 {
		t.Fatalf("unexpected third node %#v found=%v", third, ok)
	}
	_, ok = heap.Pop()
	if ok {
		t.Fatal("expected empty heap")
	}
}
