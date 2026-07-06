package path

import (
	"errors"
	"sort"

	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	"github.com/niflaot/pixels/internal/realm/room/world/surface"
)

var (
	// ErrInvalidStart reports a start position that cannot be used.
	ErrInvalidStart = errors.New("invalid room path start")

	// ErrInvalidGoal reports a goal point that cannot be used.
	ErrInvalidGoal = errors.New("invalid room path goal")

	// ErrNoPath reports that no valid path could be found.
	ErrNoPath = errors.New("room path not found")

	// ErrSearchLimit reports that a search reached its visit limit.
	ErrSearchLimit = errors.New("room path search limit reached")

	// ErrInvalidPath reports a path that no longer matches the world snapshot.
	ErrInvalidPath = errors.New("invalid room path")
)

const (
	// CardinalCost stores the base cost for cardinal movement.
	CardinalCost = 10

	// DiagonalCost stores the base cost for diagonal movement.
	DiagonalCost = 14
)

// World resolves room columns for pathfinding.
type World interface {
	// Column resolves a room tile column.
	Column(point grid.Point) (surface.Column, error)
}

// Position stores a path point with vertical section height.
type Position struct {
	// Point stores the tile coordinate.
	Point grid.Point

	// Z stores the vertical section height.
	Z grid.Height
}

// Step stores one accepted path step.
type Step struct {
	// Position stores the step destination.
	Position Position

	// Diagonal reports whether the step moved diagonally.
	Diagonal bool
}

// ColumnVersion stores a column version observed during pathfinding.
type ColumnVersion struct {
	// Point stores the observed tile coordinate.
	Point grid.Point

	// Version stores the observed column version.
	Version uint32
}

// Path stores an immutable path result.
type Path struct {
	// steps stores accepted path steps in walking order.
	steps []Step

	// versions stores observed column versions.
	versions map[grid.Point]uint32
}

// NewPath creates a path from steps without observed column versions.
func NewPath(steps []Step) Path {
	pathSteps := make([]Step, len(steps))
	copy(pathSteps, steps)

	return Path{steps: pathSteps}
}

// Rules stores pathfinding movement rules.
type Rules struct {
	// MaxStepUp stores the maximum allowed upward step height.
	MaxStepUp grid.Height

	// MaxStepDown stores the maximum allowed downward step height when falling is disabled.
	MaxStepDown grid.Height

	// AllowFalling reports whether downward movement ignores MaxStepDown.
	AllowFalling bool

	// DisableDiagonal reports whether diagonal movement is disabled.
	DisableDiagonal bool

	// MaxVisited stores the maximum number of visited nodes.
	MaxVisited int
}

// DefaultRules returns conservative room movement rules.
func DefaultRules() Rules {
	return Rules{
		MaxStepUp:   1,
		MaxStepDown: 1,
		MaxVisited:  4096,
	}
}

// Normalize returns rules with defaults for unset limits.
func (rules Rules) Normalize() Rules {
	defaults := DefaultRules()
	if rules.MaxStepUp == 0 {
		rules.MaxStepUp = defaults.MaxStepUp
	}
	if rules.MaxStepDown == 0 {
		rules.MaxStepDown = defaults.MaxStepDown
	}
	if rules.MaxVisited <= 0 {
		rules.MaxVisited = defaults.MaxVisited
	}

	return rules
}

// AllowsStep reports whether a step between sections is valid.
func (rules Rules) AllowsStep(from grid.Height, to grid.Height) bool {
	delta := to - from
	if delta > rules.MaxStepUp {
		return false
	}
	if delta < 0 && !rules.AllowFalling && -delta > rules.MaxStepDown {
		return false
	}

	return true
}

// AllowsSection reports whether a section can be entered.
func (rules Rules) AllowsSection(section surface.Section) bool {
	return section.Walkable()
}

// Len returns the number of path steps.
func (path Path) Len() int {
	return len(path.steps)
}

// Steps returns a copy of accepted path steps.
func (path Path) Steps() []Step {
	steps := make([]Step, len(path.steps))
	copy(steps, path.steps)

	return steps
}

// ColumnVersions returns observed column versions in stable order.
func (path Path) ColumnVersions() []ColumnVersion {
	versions := make([]ColumnVersion, 0, len(path.versions))
	for point, version := range path.versions {
		versions = append(versions, ColumnVersion{Point: point, Version: version})
	}
	sort.Slice(versions, func(left int, right int) bool {
		if versions[left].Point.Y == versions[right].Point.Y {
			return versions[left].Point.X < versions[right].Point.X
		}

		return versions[left].Point.Y < versions[right].Point.Y
	})

	return versions
}

// Validate verifies that observed column versions still match.
func (path Path) Validate(world World) error {
	for point, version := range path.versions {
		column, err := world.Column(point)
		if err != nil {
			return ErrInvalidPath
		}
		if column.Version() != version {
			return ErrInvalidPath
		}
	}

	return nil
}

// movementCost returns the movement cost between nodes.
func movementCost(from nodeKey, to nodeKey, diagonal bool) int {
	cost := CardinalCost
	if diagonal {
		cost = DiagonalCost
	}
	if to.Z > from.Z {
		cost += int(to.Z - from.Z)
	}

	return cost
}

// heuristic returns an octile distance estimate.
func heuristic(from nodeKey, goal grid.Point) int {
	dx := absInt(int(goal.X) - int(from.X))
	dy := absInt(int(goal.Y) - int(from.Y))
	minimum := dx
	maximum := dy
	if minimum > maximum {
		minimum = dy
		maximum = dx
	}

	return DiagonalCost*minimum + CardinalCost*(maximum-minimum)
}

// absInt returns the absolute int value.
func absInt(value int) int {
	if value < 0 {
		return -value
	}

	return value
}
