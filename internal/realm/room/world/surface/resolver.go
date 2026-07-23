package surface

import "github.com/niflaot/pixels/internal/realm/room/world/grid"

// Resolver resolves compact tile columns from a base grid and fixtures.
type Resolver struct {
	// grid stores the immutable base room grid.
	grid grid.Grid

	// fixtures stores dynamic fixtures grouped by grid index.
	fixtures map[int][]Fixture

	// versions stores dynamic column versions grouped by grid index.
	versions map[int]uint32
}

// NewResolver creates a room surface resolver.
func NewResolver(roomGrid grid.Grid, fixtures []Fixture) (*Resolver, error) {
	resolver := &Resolver{
		grid:     roomGrid,
		fixtures: make(map[int][]Fixture),
		versions: make(map[int]uint32),
	}
	for _, fixture := range fixtures {
		if err := resolver.AddFixture(fixture); err != nil {
			return nil, err
		}
	}

	return resolver, nil
}

// AddFixture adds a dynamic fixture and bumps the affected column version.
func (resolver *Resolver) AddFixture(fixture Fixture) error {
	index, ok := resolver.grid.Index(fixture.Point())
	if !ok || !resolver.grid.Valid(fixture.Point()) {
		return ErrInvalidTile
	}

	resolver.fixtures[index] = append(resolver.fixtures[index], fixture)
	resolver.versions[index]++

	return nil
}

// RemoveFixtures removes every fixture matching a source id and bumps the version of each affected column.
func (resolver *Resolver) RemoveFixtures(sourceID int64) int {
	removed := 0
	for index, fixtures := range resolver.fixtures {
		remaining, count := withoutSource(fixtures, sourceID)
		if count == 0 {
			continue
		}
		resolver.fixtures[index] = remaining
		resolver.versions[index]++
		removed += count
	}

	return removed
}

// ReplaceFixtures atomically removes fixtures for a source id and adds replacement fixtures.
func (resolver *Resolver) ReplaceFixtures(sourceID int64, fixtures []Fixture) error {
	for _, fixture := range fixtures {
		if !resolver.grid.Valid(fixture.Point()) {
			return ErrInvalidTile
		}
	}

	resolver.RemoveFixtures(sourceID)
	for _, fixture := range fixtures {
		if err := resolver.AddFixture(fixture); err != nil {
			return err
		}
	}

	return nil
}

// Column resolves a tile column.
func (resolver *Resolver) Column(point grid.Point) (Column, error) {
	tile, ok := resolver.grid.Tile(point)
	if !ok || !tile.Valid() {
		return Column{}, ErrInvalidTile
	}

	index, _ := resolver.grid.Index(point)
	column := resolver.columnFor(tile, resolver.versions[index], resolver.fixtures[index])

	return column, nil
}

// SectionAt finds a section at a point and height.
func (resolver *Resolver) SectionAt(point grid.Point, height grid.Height) (Section, error) {
	column, err := resolver.Column(point)
	if err != nil {
		return Section{}, err
	}

	section, ok := column.SectionAt(height)
	if !ok {
		return Section{}, ErrNoSection
	}

	return section, nil
}

// TopSection returns the visible top section for a point.
func (resolver *Resolver) TopSection(point grid.Point) (Section, error) {
	column, err := resolver.Column(point)
	if err != nil {
		return Section{}, err
	}

	section, ok := column.TopSection()
	if !ok {
		return Section{}, ErrNoSection
	}

	return section, nil
}

// columnFor resolves base and fixture sections for a tile.
func (resolver *Resolver) columnFor(tile grid.Tile, version uint32, fixtures []Fixture) Column {
	column := NewColumn(tile.Point(), version)
	column.AddSection(baseSection(tile))
	for _, fixture := range fixtures {
		column.AddSection(fixture.Section())
	}

	return column
}

// baseSection creates the implicit base section for a valid tile.
func baseSection(tile grid.Tile) Section {
	return NewSection(SectionParams{
		Point:     tile.Point(),
		Z:         tile.Height(),
		Bottom:    tile.Height(),
		Top:       tile.Height(),
		Clearance: grid.AvatarClearance,
		State:     StateOpen,
		Stacking:  true,
		Source:    SourceBase,
	})
}

// withoutSource filters fixtures matching a source id, reporting the remaining slice and removed count.
func withoutSource(fixtures []Fixture, sourceID int64) ([]Fixture, int) {
	remaining := make([]Fixture, 0, len(fixtures))
	removed := 0
	for _, fixture := range fixtures {
		if fixture.SourceID() == sourceID {
			removed++

			continue
		}
		remaining = append(remaining, fixture)
	}

	return remaining, removed
}
