package surface

import (
	"testing"

	"github.com/niflaot/pixels/internal/realm/room/world/grid"
)

// TestColumnStoresSectionsInlineAndOverflow verifies compact column storage.
func TestColumnStoresSectionsInlineAndOverflow(t *testing.T) {
	point := grid.MustPoint(1, 1)
	column := NewColumn(point, 3)

	for height := 9; height >= 0; height-- {
		column.AddSection(NewSection(SectionParams{
			Point:  point,
			Z:      grid.Height(height),
			Top:    grid.Height(height),
			State:  StateOpen,
			Source: SourceFixture,
		}))
	}

	if column.Point() != point || column.Version() != 3 || column.Len() != 10 {
		t.Fatalf("unexpected column metadata")
	}

	sections := column.Sections()
	for index, section := range sections {
		if section.Z() != grid.Height(index) {
			t.Fatalf("expected sorted height %d, got %d", index, section.Z())
		}
		sectionAt, ok := column.Section(index)
		if !ok || sectionAt.Z() != grid.Height(index) {
			t.Fatalf("expected section at %d, got %d found=%v", index, sectionAt.Z(), ok)
		}
	}

	top, ok := column.TopSection()
	if !ok || top.Z() != 9 {
		t.Fatalf("unexpected top section %d found=%v", top.Z(), ok)
	}
}

// TestColumnBlockedSectionReplacesTiedHeight verifies blocking sections win height ties.
func TestColumnBlockedSectionReplacesTiedHeight(t *testing.T) {
	point := grid.MustPoint(1, 1)
	column := NewColumn(point, 1)
	column.AddSection(NewSection(SectionParams{Point: point, Z: 0, Top: 0, State: StateOpen, Source: SourceBase}))
	column.AddSection(NewSection(SectionParams{Point: point, Z: 5, Top: 5, State: StateOpen, Source: SourceFixture}))
	column.AddSection(NewSection(SectionParams{Point: point, Z: 0, Top: 2, State: StateBlocked, Source: SourceFixture, SourceID: 9}))

	if column.Len() != 2 {
		t.Fatalf("expected tied section replaced, got %d sections: %#v", column.Len(), column.Sections())
	}
	section, ok := column.SectionAt(0)
	if !ok || section.State() != StateBlocked || section.Top() != 2 || section.SourceID() != 9 {
		t.Fatalf("expected replaced blocked section, got %#v found=%v", section, ok)
	}
	other, ok := column.SectionAt(5)
	if !ok || other.State() != StateOpen {
		t.Fatalf("expected untouched section at different height, got %#v found=%v", other, ok)
	}
}

// TestColumnReportsEmptyTop verifies empty column top lookup.
func TestColumnReportsEmptyTop(t *testing.T) {
	column := NewColumn(grid.MustPoint(1, 1), 0)

	_, ok := column.TopSection()
	if ok {
		t.Fatal("expected missing top section")
	}
	_, ok = column.Section(0)
	if ok {
		t.Fatal("expected missing indexed section")
	}
}
