package surface

import "github.com/niflaot/pixels/internal/realm/room/world/grid"

// Fixture stores a compact future item or dynamic tile contribution.
type Fixture struct {
	// point stores the affected tile coordinate.
	point grid.Point

	// z stores the fixture walkable height.
	z grid.Height

	// bottom stores the occupied volume bottom height.
	bottom grid.Height

	// top stores the fixture occupied top height.
	top grid.Height

	// clearance stores the required free height above the fixture.
	clearance grid.Height

	// state stores the fixture movement state.
	state State

	// stacking reports whether another item can stack over the fixture.
	stacking bool

	// source stores the fixture source type.
	source Source

	// sourceID stores the optional source record id.
	sourceID int64
}

// NewFixture creates a room surface fixture.
func NewFixture(params FixtureParams) (Fixture, error) {
	bottom := params.Bottom
	if !params.HasBottom {
		bottom = params.Z
	}
	if bottom > params.Z || params.Top < params.Z || params.State == StateInvalid {
		return Fixture{}, ErrInvalidFixture
	}
	source := params.Source
	if source == SourceBase {
		source = SourceFixture
	}

	return Fixture{
		point:     params.Point,
		z:         params.Z,
		bottom:    bottom,
		top:       params.Top,
		clearance: params.Clearance,
		state:     params.State,
		stacking:  params.Stacking,
		source:    source,
		sourceID:  params.SourceID,
	}, nil
}

// FixtureParams stores fixture creation input.
type FixtureParams struct {
	// Point stores the affected tile coordinate.
	Point grid.Point

	// Z stores the fixture walkable height.
	Z grid.Height

	// Bottom stores the occupied volume bottom height.
	Bottom grid.Height

	// HasBottom reports whether Bottom was explicitly supplied.
	HasBottom bool

	// Top stores the fixture occupied top height.
	Top grid.Height

	// Clearance stores the required free height above the fixture.
	Clearance grid.Height

	// State stores the fixture movement state.
	State State

	// Stacking reports whether another item can stack over the fixture.
	Stacking bool

	// Source stores the fixture source type.
	Source Source

	// SourceID stores the optional source record id.
	SourceID int64
}

// Point returns the affected tile coordinate.
func (fixture Fixture) Point() grid.Point {
	return fixture.point
}

// SourceID returns the optional source record id.
func (fixture Fixture) SourceID() int64 {
	return fixture.sourceID
}

// Section converts the fixture to a resolved section.
func (fixture Fixture) Section() Section {
	return NewSection(SectionParams{
		Point:     fixture.point,
		Z:         fixture.z,
		Bottom:    fixture.bottom,
		Top:       fixture.top,
		Clearance: fixture.clearance,
		State:     fixture.state,
		Stacking:  fixture.stacking,
		Source:    fixture.source,
		SourceID:  fixture.sourceID,
	})
}
