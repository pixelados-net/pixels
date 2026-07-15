package surface

import "github.com/niflaot/pixels/internal/realm/room/world/grid"

const (
	// inlineSectionLimit stores the sections kept inside a column value.
	inlineSectionLimit = 8
)

// Column stores resolved sections for a tile that has dynamic state.
type Column struct {
	// point stores the tile coordinate.
	point grid.Point

	// version stores the monotonic column version.
	version uint32

	// count stores the number of inline sections.
	count uint8

	// sections stores common tile sections without heap allocation.
	sections [inlineSectionLimit]Section

	// extra stores rare overflow sections.
	extra []Section
}

// NewColumn creates a resolved tile column.
func NewColumn(point grid.Point, version uint32) Column {
	return Column{point: point, version: version}
}

// Point returns the tile coordinate.
func (column Column) Point() grid.Point {
	return column.point
}

// Version returns the column version.
func (column Column) Version() uint32 {
	return column.version
}

// Sections returns the resolved tile sections.
func (column Column) Sections() []Section {
	sections := make([]Section, 0, column.Len())
	if len(column.extra) > 0 {
		sections = append(sections, column.extra...)

		return sections
	}
	for index := 0; index < int(column.count); index++ {
		sections = append(sections, column.sections[index])
	}

	return sections
}

// Section returns a resolved section by ordered index.
func (column Column) Section(index int) (Section, bool) {
	if index < 0 || index >= column.Len() {
		return Section{}, false
	}
	if len(column.extra) > 0 {
		return column.extra[index], true
	}
	if index < int(column.count) {
		return column.sections[index], true
	}

	return Section{}, false
}

// Len returns the number of resolved sections.
func (column Column) Len() int {
	return int(column.count) + len(column.extra)
}

// AddSection adds a resolved tile section, letting a blocking, sit, or lay section replace a
// tied-height section rather than duplicate it, since a tile can only have one such terminal
// state at a given height.
func (column *Column) AddSection(section Section) {
	column.removeCoveredWalkableSections(section)
	if section.state.replacesTiedSection() && column.replaceTiedSection(section) {
		return
	}
	if len(column.extra) > 0 {
		column.insertExtra(section)

		return
	}
	if int(column.count) < len(column.sections) {
		column.insertInline(section)

		return
	}

	column.promoteToExtra()
	column.insertExtra(section)
}

// WalkableSectionAt finds an exact section with sufficient avatar clearance.
func (column Column) WalkableSectionAt(height grid.Height) (Section, bool) {
	section, found := column.SectionAt(height)
	if !found || !column.Accepts(section) {
		return Section{}, false
	}

	return section, true
}

// NearestWalkableSection finds the closest usable section to a height.
func (column Column) NearestWalkableSection(height grid.Height) (Section, bool) {
	var selected Section
	var distance grid.Height
	found := false
	for index := 0; index < column.Len(); index++ {
		section, ok := column.Section(index)
		if !ok || !column.Accepts(section) {
			continue
		}
		candidateDistance := section.Z() - height
		if candidateDistance < 0 {
			candidateDistance = -candidateDistance
		}
		if !found || candidateDistance < distance {
			selected, distance, found = section, candidateDistance, true
		}
	}

	return selected, found
}

// Accepts reports whether a section is walkable with sufficient free headroom.
func (column Column) Accepts(section Section) bool {
	if !section.Walkable() {
		return false
	}
	requiredTop := section.Z() + section.Clearance()
	for index := 0; index < column.Len(); index++ {
		other, ok := column.Section(index)
		if !ok || sameSection(section, other) || other.Top() <= section.Z() {
			continue
		}
		if other.Bottom() < requiredTop {
			return false
		}
	}

	return true
}

// removeCoveredWalkableSections removes planes hidden inside one occupied volume.
func (column *Column) removeCoveredWalkableSections(section Section) {
	if section.Top() <= section.Bottom() {
		return
	}
	kept := make([]Section, 0, column.Len())
	for index := 0; index < column.Len(); index++ {
		current, ok := column.Section(index)
		if !ok {
			continue
		}
		covered := current.Walkable() && current.Z() >= section.Bottom() && current.Z() < section.Top()
		if !covered {
			kept = append(kept, current)
		}
	}
	column.replaceSections(kept)
}

// replaceSections rebuilds inline or overflow storage from ordered sections.
func (column *Column) replaceSections(sections []Section) {
	column.count = 0
	column.extra = nil
	if len(sections) <= inlineSectionLimit {
		copy(column.sections[:], sections)
		column.count = uint8(len(sections))
		return
	}
	column.extra = append(column.extra, sections...)
}

// sameSection reports whether two values describe the same resolved contribution.
func sameSection(first Section, second Section) bool {
	return first.Source() == second.Source() && first.SourceID() == second.SourceID() &&
		first.Z() == second.Z() && first.Bottom() == second.Bottom() && first.Top() == second.Top()
}

// SectionAt finds a section at the exact walkable height.
func (column Column) SectionAt(height grid.Height) (Section, bool) {
	for index := 0; index < int(column.count); index++ {
		section := column.sections[index]
		if section.Z() == height {
			return section, true
		}
	}
	for _, section := range column.extra {
		if section.Z() == height {
			return section, true
		}
	}

	return Section{}, false
}

// TopSection returns the highest walkable or blocking section.
func (column Column) TopSection() (Section, bool) {
	if column.Len() == 0 {
		return Section{}, false
	}
	if len(column.extra) > 0 {
		return column.extra[len(column.extra)-1], true
	}

	return column.sections[column.count-1], true
}

// Dynamic reports whether the column was materialized from dynamic state.
func (column Column) Dynamic() bool {
	return column.version > 0
}

// replaceTiedSection replaces an existing section at the same height with a blocking section.
func (column *Column) replaceTiedSection(section Section) bool {
	if len(column.extra) > 0 {
		for index := range column.extra {
			if column.extra[index].Z() == section.Z() {
				column.extra[index] = section

				return true
			}
		}

		return false
	}
	for index := 0; index < int(column.count); index++ {
		if column.sections[index].Z() == section.Z() {
			column.sections[index] = section

			return true
		}
	}

	return false
}

// insertInline adds an inline section ordered by height.
func (column *Column) insertInline(section Section) {
	position := int(column.count)
	for position > 0 && column.sections[position-1].Z() > section.Z() {
		column.sections[position] = column.sections[position-1]
		position--
	}
	column.sections[position] = section
	column.count++
}

// promoteToExtra moves inline sections to overflow storage.
func (column *Column) promoteToExtra() {
	column.extra = make([]Section, int(column.count), int(column.count)+1)
	copy(column.extra, column.sections[:column.count])
	column.count = 0
}

// insertExtra adds an overflow section ordered by height.
func (column *Column) insertExtra(section Section) {
	column.extra = append(column.extra, section)
	for index := len(column.extra) - 1; index > 0 && column.extra[index-1].Z() > section.Z(); index-- {
		column.extra[index] = column.extra[index-1]
		column.extra[index-1] = section
	}
}
