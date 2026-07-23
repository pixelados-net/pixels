// Package badge validates, compiles, and caches social-group badge editor data.
package badge

import (
	"context"
	"sync/atomic"

	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
)

// Source reads enabled durable badge reference data.
type Source interface {
	// BadgeRegistry returns every enabled editor element and color.
	BadgeRegistry(context.Context) ([]grouprecord.BadgeElement, []grouprecord.BadgeColor, error)
}

// Snapshot is one immutable badge editor generation.
type Snapshot struct {
	// Elements stores ordered editor elements.
	Elements []grouprecord.BadgeElement
	// Colors stores ordered editor colors.
	Colors []grouprecord.BadgeColor
	// element stores zero-allocation validation keys.
	element map[elementKey]struct{}
	// color stores zero-allocation validation keys.
	color map[colorKey]string
}

// elementKey identifies one editor element.
type elementKey struct {
	kind grouprecord.BadgeKind
	id   int32
}

// colorKey identifies one editor color.
type colorKey struct {
	family grouprecord.ColorFamily
	id     int32
}

// Registry stores one immutable badge editor snapshot.
type Registry struct {
	// source reads durable reference data.
	source Source
	// current stores the warmed immutable generation.
	current atomic.Pointer[Snapshot]
}

// New creates an empty badge registry.
func New(source Source) *Registry { return &Registry{source: source} }

// Refresh replaces the current generation after a complete read.
func (registry *Registry) Refresh(ctx context.Context) error {
	elements, colors, err := registry.source.BadgeRegistry(ctx)
	if err != nil {
		return err
	}
	snapshot := &Snapshot{Elements: elements, Colors: colors, element: make(map[elementKey]struct{}, len(elements)), color: make(map[colorKey]string, len(colors))}
	for _, element := range elements {
		snapshot.element[elementKey{kind: element.Kind, id: element.ID}] = struct{}{}
	}
	for _, color := range colors {
		snapshot.color[colorKey{family: color.Family, id: color.ID}] = color.Hex
	}
	registry.current.Store(snapshot)
	return nil
}

// Snapshot returns the warmed immutable generation.
func (registry *Registry) Snapshot() (*Snapshot, bool) {
	snapshot := registry.current.Load()
	return snapshot, snapshot != nil
}

// Element reports whether an editor element is enabled.
func (snapshot *Snapshot) Element(kind grouprecord.BadgeKind, id int32) bool {
	if snapshot == nil {
		return false
	}
	_, found := snapshot.element[elementKey{kind: kind, id: id}]
	return found
}

// Color returns one enabled RGB value without allocation.
func (snapshot *Snapshot) Color(family grouprecord.ColorFamily, id int32) (string, bool) {
	if snapshot == nil {
		return "", false
	}
	value, found := snapshot.color[colorKey{family: family, id: id}]
	return value, found
}
