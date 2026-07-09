package path

import (
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	"github.com/niflaot/pixels/internal/realm/room/world/surface"
)

// Finder calculates paths over a room world snapshot.
type Finder struct {
	// world resolves room columns.
	world World

	// rules stores movement rules.
	rules Rules

	// occupancy stores blocked path positions.
	occupancy Occupancy
}

// NewFinder creates a path finder.
func NewFinder(world World, rules Rules) *Finder {
	return &Finder{world: world, rules: rules.Normalize()}
}

// NewFinderWithOccupancy creates a path finder with blocked positions.
func NewFinderWithOccupancy(world World, rules Rules, occupancy Occupancy) *Finder {
	return &Finder{world: world, rules: rules.Normalize(), occupancy: occupancy}
}

// Find calculates a path from start to goal.
func (finder *Finder) Find(start Position, goal grid.Point) (Path, error) {
	search := newSearch(finder.world, finder.rules, finder.occupancy, goal)
	if err := search.validateStart(start); err != nil {
		return Path{}, err
	}
	if _, err := search.column(goal); err != nil {
		return Path{}, ErrInvalidGoal
	}

	return search.run(nodeFromPosition(start))
}

// search stores one A* search state.
type search struct {
	// world resolves room columns.
	world World

	// rules stores movement rules.
	rules Rules

	// occupancy stores blocked path positions.
	occupancy Occupancy

	// goal stores the target tile.
	goal grid.Point

	// open stores the open node heap.
	open openHeap

	// cameFrom stores previous nodes by node.
	cameFrom map[nodeKey]nodeKey

	// gScore stores accumulated movement cost by node.
	gScore map[nodeKey]int

	// versions stores observed column versions by point.
	versions map[grid.Point]uint32

	// visited stores the number of popped nodes.
	visited int
}

// newSearch creates a path search state.
func newSearch(world World, rules Rules, occupancy Occupancy, goal grid.Point) *search {
	return &search{
		world:     world,
		rules:     rules,
		occupancy: occupancy,
		goal:      goal,
		open:      openHeap{nodes: make([]openNode, 0, 128)},
		cameFrom:  make(map[nodeKey]nodeKey, 64),
		gScore:    make(map[nodeKey]int, 64),
		versions:  make(map[grid.Point]uint32, 64),
	}
}

// validateStart verifies that the start position exists.
func (search *search) validateStart(start Position) error {
	column, err := search.column(start.Point)
	if err != nil {
		return ErrInvalidStart
	}
	section, ok := column.SectionAt(start.Z)
	if !ok || !search.rules.AllowsSection(section) {
		return ErrInvalidStart
	}

	return nil
}

// run executes A*.
func (search *search) run(start nodeKey) (Path, error) {
	search.gScore[start] = 0
	search.open.Push(openNode{key: start, priority: heuristic(start, search.goal)})
	for search.open.Len() > 0 {
		current, _ := search.open.Pop()
		if search.isStale(current) {
			continue
		}
		if current.key.point() == search.goal {
			return search.pathTo(current.key), nil
		}
		if err := search.visit(current); err != nil {
			return Path{}, err
		}
	}

	return Path{}, ErrNoPath
}

// visit expands one open node.
func (search *search) visit(current openNode) error {
	search.visited++
	if search.visited > search.rules.MaxVisited {
		return ErrSearchLimit
	}
	for _, direction := range directions {
		if search.skipDirection(current.key, direction) {
			continue
		}
		point, ok := neighborPoint(current.key, direction)
		if !ok {
			continue
		}
		search.visitNeighbor(current, point, direction.diagonal)
	}

	return nil
}

// visitNeighbor evaluates one neighbor column. Sit and lay sections are destination-only: a unit
// can enter them when the tile is the search goal but never route through them, and a goal whose
// top section is a slot resolves to that slot alone so reaching a chair or bed tile always means
// using it rather than whichever tied-cost section a search happens to pop first.
func (search *search) visitNeighbor(current openNode, point grid.Point, diagonal bool) {
	column, err := search.column(point)
	if err != nil {
		return
	}
	if point == search.goal {
		if top, ok := column.TopSection(); ok && isSlotState(top.State()) {
			if search.acceptsSection(current, top) {
				search.accept(current, top, diagonal)
			}

			return
		}
	}
	for index := 0; index < column.Len(); index++ {
		section, ok := column.Section(index)
		if !ok || isSlotState(section.State()) || !search.acceptsSection(current, section) {
			continue
		}
		search.accept(current, section, diagonal)
	}
}

// isSlotState reports whether a section state represents an interactive sit or lay slot.
func isSlotState(state surface.State) bool {
	return state == surface.StateSit || state == surface.StateLay
}

// acceptsSection reports whether a section can be entered from current.
func (search *search) acceptsSection(current openNode, section surface.Section) bool {
	next := nodeKey{X: section.Point().X, Y: section.Point().Y, Z: section.Z()}

	return search.rules.AllowsSection(section) &&
		search.rules.AllowsStep(current.key.Z, section.Z()) &&
		!search.occupancy.blocks(next)
}

// accept records an accepted neighbor section.
func (search *search) accept(current openNode, section surface.Section, diagonal bool) {
	next := nodeKey{X: section.Point().X, Y: section.Point().Y, Z: section.Z()}
	cost := current.cost + movementCost(current.key, next, diagonal)
	best, seen := search.gScore[next]
	if seen && cost >= best {
		return
	}
	search.cameFrom[next] = current.key
	search.gScore[next] = cost
	search.open.Push(openNode{
		key:      next,
		cost:     cost,
		priority: cost + heuristic(next, search.goal),
	})
}

// isStale reports whether an open node has an obsolete cost.
func (search *search) isStale(node openNode) bool {
	best, ok := search.gScore[node.key]

	return ok && node.cost > best
}

// column returns a cached observed column.
func (search *search) column(point grid.Point) (surface.Column, error) {
	column, err := search.world.Column(point)
	if err != nil {
		return surface.Column{}, err
	}
	search.versions[point] = column.Version()

	return column, nil
}

// pathTo reconstructs a path to the goal node.
func (search *search) pathTo(goal nodeKey) Path {
	keys := make([]nodeKey, 0, 16)
	for current := goal; ; current = search.cameFrom[current] {
		keys = append(keys, current)
		_, ok := search.cameFrom[current]
		if !ok {
			break
		}
	}
	steps := make([]Step, 0, len(keys)-1)
	for index := len(keys) - 2; index >= 0; index-- {
		current := keys[index]
		previous := keys[index+1]
		steps = append(steps, Step{
			Position: current.position(),
			Diagonal: current.X != previous.X &&
				current.Y != previous.Y,
		})
	}

	return Path{steps: steps, versions: search.versions}
}
