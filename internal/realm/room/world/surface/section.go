package surface

import "github.com/niflaot/pixels/internal/realm/room/world/grid"

// State stores the movement state of a tile section.
type State uint8

const (
	// StateInvalid marks a section that cannot exist.
	StateInvalid State = iota

	// StateOpen marks a walkable section.
	StateOpen

	// StateBlocked marks a blocking section.
	StateBlocked

	// StateSit marks a section that can be used as a seat target.
	StateSit

	// StateLay marks a section that can be used as a lay target.
	StateLay
)

// Walkable reports whether the state can accept normal movement.
func (state State) Walkable() bool {
	return state == StateOpen || state == StateSit || state == StateLay
}

// replacesTiedSection reports whether the state is a terminal, mutually-exclusive state that
// must replace rather than duplicate an existing section tied at the same height.
func (state State) replacesTiedSection() bool {
	return state == StateBlocked || state == StateSit || state == StateLay
}

// Source stores the origin of a resolved tile section.
type Source uint8

const (
	// SourceBase marks a section produced by the base grid.
	SourceBase Source = iota

	// SourceFixture marks a section produced by a generic future fixture.
	SourceFixture

	// SourceStack marks a section produced by a stack helper.
	SourceStack

	// SourceGate marks a section produced by a gate.
	SourceGate
)

// Section stores one resolved vertical tile section.
type Section struct {
	// point stores the owning tile coordinate.
	point grid.Point

	// z stores the walkable section height.
	z grid.Height

	// bottom stores the occupied volume bottom height.
	bottom grid.Height

	// top stores the occupied top height of the section source.
	top grid.Height

	// clearance stores the required free height above the section.
	clearance grid.Height

	// state stores the movement state.
	state State

	// stacking reports whether another item can stack over the section.
	stacking bool

	// source stores the section source type.
	source Source

	// sourceID stores the optional source record id.
	sourceID int64
}

// NewSection creates a resolved tile section.
func NewSection(params SectionParams) Section {
	return Section{
		point:     params.Point,
		z:         params.Z,
		bottom:    params.Bottom,
		top:       params.Top,
		clearance: params.Clearance,
		state:     params.State,
		stacking:  params.Stacking,
		source:    params.Source,
		sourceID:  params.SourceID,
	}
}

// SectionParams stores section creation input.
type SectionParams struct {
	// Point stores the owning tile coordinate.
	Point grid.Point

	// Z stores the walkable section height.
	Z grid.Height

	// Bottom stores the occupied volume bottom height.
	Bottom grid.Height

	// Top stores the occupied top height of the section source.
	Top grid.Height

	// Clearance stores the required free height above the section.
	Clearance grid.Height

	// State stores the movement state.
	State State

	// Stacking reports whether another item can stack over the section.
	Stacking bool

	// Source stores the section source type.
	Source Source

	// SourceID stores the optional source record id.
	SourceID int64
}

// Point returns the owning tile coordinate.
func (section Section) Point() grid.Point {
	return section.point
}

// Z returns the walkable section height.
func (section Section) Z() grid.Height {
	return section.z
}

// Bottom returns the occupied volume bottom height.
func (section Section) Bottom() grid.Height {
	return section.bottom
}

// Top returns the occupied top height.
func (section Section) Top() grid.Height {
	return section.top
}

// Clearance returns the required free height above the section.
func (section Section) Clearance() grid.Height {
	return section.clearance
}

// State returns the movement state.
func (section Section) State() State {
	return section.state
}

// Stacking reports whether the section accepts stacking.
func (section Section) Stacking() bool {
	return section.stacking
}

// Source returns the section source type.
func (section Section) Source() Source {
	return section.source
}

// SourceID returns the optional source record id.
func (section Section) SourceID() int64 {
	return section.sourceID
}

// Walkable reports whether the section can accept normal movement.
func (section Section) Walkable() bool {
	return section.state.Walkable()
}
