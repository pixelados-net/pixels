# The 2.5D Heightmap and Stacking

Third page of the Rooms section. A room isn't flat, and it isn't fully 3D either. Every classic client renders a 2.5D world: a grid of tiles, each with its own base height, where furniture can add walkable surfaces above that base. This page covers how Pixels represents that: a fixed-point height unit, a base grid, and a resolver that layers dynamic furniture on top of it to produce the actual walkable surface at any tile.

## Quarter-unit fixed point

Room height, on the wire and in the classic tools, is a decimal room unit (`2.5`, `1.0`). Storing that as a float everywhere invites the usual float-comparison bugs in pathfinding, so the world package converts it once, at the boundary, into a fixed-point integer scaled by four:

```go
type Height int16

const (
	// HeightScale stores the number of fixed-point steps in one room height unit.
	HeightScale Height = 4

	// AvatarClearance stores the standing avatar clearance in fixed-point units.
	AvatarClearance Height = 8
)

func HeightFromUnits(value float64) Height {
	return Height(math.Round(value * float64(HeightScale)))
}

func (height Height) Units() float64 {
	return float64(height) / float64(HeightScale)
}
```

Four steps per unit means quarter-unit precision. The smallest height difference the world can represent is `0.25`, which matches the finest step height real furniture definitions use. `AvatarClearance` of `8` fixed-point units is exactly two room units: how much headroom a standing avatar needs above a surface before something above it counts as blocking. Every height comparison in pathfinding and stacking happens in this integer space; the decimal form only exists at the packet encode/decode boundary (`grid.HeightFromUnits` / `Height.String()`), so nothing downstream ever compares floats.

## The base grid

`grid.Grid` is the immutable per-room heightmap parsed once from the room's floor plan: a flat base height and a couple of flags (invalid tile, door tile) per point, indexed by `y*width+x`. It answers exactly the questions a static floor plan can answer: is this tile part of the room, what's its base height, is it the door, and nothing about what's currently placed on it. That's deliberately the surface resolver's job, not the grid's.

## Columns and sections: layering furniture on top of the floor

A tile in a 2.5D room can have more than one usable height at once: the floor itself, plus the top of a rug, plus the seat of a chair on that rug, each a distinct walkable "section." `surface.Resolver` is what turns the static base grid plus whatever furniture currently occupies a tile into that list of sections:

```go
type Resolver struct {
	grid     grid.Grid
	fixtures map[int][]Fixture // dynamic furniture-derived fixtures, grouped by grid index
	versions map[int]uint32    // per-tile version, bumped on every fixture change
}

func (resolver *Resolver) Column(point grid.Point) (Column, error)
func (resolver *Resolver) SectionAt(point grid.Point, height grid.Height) (Section, error)
```

Each `Section` records a height (`z`), a movement `State`, and where it came from:

```go
type State uint8

const (
	StateInvalid State = iota
	StateOpen          // walkable
	StateBlocked        // occupies the tile but isn't walkable
	StateSit            // walkable and usable as a seat target
	StateLay            // walkable and usable as a lay target
)

type Source uint8

const (
	SourceBase    Source = iota // the static floor plan itself
	SourceFixture               // a generic placed item (its footprint plus StackHeight)
	SourceStack                 // a manual override from the stack helper, see FURNITURE-ADVANCED
	SourceGate                  // a gate interaction's walkable/blocked toggle
)
```

`Column` stores a small inline array of sections (eight, before falling back to a heap slice for the rare tile with more) and a monotonic `version` bumped every time a fixture is added or removed at that tile. The version is how callers can tell a cached column is stale without re-fetching it defensively on every read.

## How stacking actually composes

`AddSection` is where the composition rules live, and its comment states the invariant directly:

```go
// AddSection adds a resolved tile section, letting a blocking, sit, or lay section replace a
// tied-height section rather than duplicate it, since a tile can only have one such terminal
// state at a given height.
func (column *Column) AddSection(section Section) {
	column.removeCoveredWalkableSections(section)
	if section.state.replacesTiedSection() && column.replaceTiedSection(section) {
		return
	}
	...
}
```

Two rules fall out of this. First, a new section covers (removes) any walkable section beneath it that no longer has enough `AvatarClearance` above it to stand on: placing a low-clearance item over an existing open section removes that section from the walkable set rather than leaving a phantom floor a unit could still be routed onto. Second, `StateBlocked`, `StateSit`, and `StateLay` are *terminal* states: a tile can't simultaneously be a sit target and a lay target at the same height, so a new terminal section at a height that's already occupied replaces the old one instead of stacking a duplicate. `StateOpen` sections don't have this restriction; several open sections can coexist at different heights on the same tile, which is exactly what lets a rug on the floor and a chair's seat both be valid, distinct walkable heights on one point.

Reading a column back for pathfinding goes through two helpers rather than a raw section index:

```go
func (column Column) WalkableSectionAt(height grid.Height) (Section, bool)     // exact height match
func (column Column) NearestWalkableSection(height grid.Height) (Section, bool) // closest usable height
func (column Column) Accepts(section Section) bool                              // walkable + enough headroom
```

`Accepts` is where `AvatarClearance` actually gets enforced: a section only counts as usable if nothing else in the column occupies the space between its height and `height + AvatarClearance`.

## Step height and pathfinding

Moving between adjacent tiles isn't free of height constraints. Pathfinding rejects a step whose height delta is too large to be a normal walk-up, forcing stairs or a ramp instead:

```go
type Rules struct {
	MaxStepUp grid.Height
	...
}

var defaults = Rules{
	MaxStepUp: 6, // 1.5 room units, in quarter-unit fixed point
}
```

`6` fixed-point units is `1.5` real room units. A unit can step up onto a low stack or a stair tread but not straight onto a tall platform; reaching a tall surface has to come through several smaller steps or a dedicated ramp-shaped set of sections rather than one large jump.

## Where custom overrides fit in

The stack helper (`SourceStack`, covered in [[FURNITURE-ADVANCED]]) exists precisely because `SourceFixture`'s automatic footprint-plus-`StackHeight` derivation doesn't always match what a room designer wants: it lets a specific placed item's contribution to the column be pinned to an exact height instead of the definition's computed one, without touching the catalog definition itself. Because `AddSection`'s tie-breaking and clearance rules apply uniformly regardless of `Source`, a manually pinned section composes with everything else on the tile exactly the same way an automatic one would. The override changes *what height* gets contributed, not *how* stacking resolves once it's there.
