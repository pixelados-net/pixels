// Package configuration validates and compiles durable WIRED settings.
package configuration

import (
	"time"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/registry"
)

// Point identifies one room stack tile.
type Point struct {
	// X stores the horizontal tile coordinate.
	X int
	// Y stores the vertical tile coordinate.
	Y int
}

// Parameters stores settings parsed once outside execution hot paths.
type Parameters struct {
	// Values stores descriptor-specific integer values.
	Values []int32
	// Text stores normalized descriptor text.
	Text string
	// Name stores a parsed bot or player name.
	Name string
	// Message stores a parsed bot message.
	Message string
	// Number stores a strict integer parsed from compatibility text.
	Number int32
	// Duration stores a validated trigger or effect duration.
	Duration time.Duration
}

// Node stores one compiled immutable WIRED node.
type Node struct {
	// ItemID identifies the configured furniture item.
	ItemID int64
	// RoomID identifies the containing room.
	RoomID int64
	// SpriteID identifies the client furniture definition.
	SpriteID int32
	// Point stores the stack tile.
	Point Point
	// Descriptor stores canonical behavior metadata.
	Descriptor registry.Descriptor
	// Parameters stores parsed behavior settings.
	Parameters Parameters
	// SelectionMode stores Nitro's target matching mode.
	SelectionMode int32
	// Delay stores the compiled action delay.
	Delay time.Duration
	// Version stores durable configuration version.
	Version int64
	// Targets stores immutable selected furniture records.
	Targets []record.Target
}

// Stack stores compiled nodes occupying one room tile.
type Stack struct {
	// Point identifies the stack tile.
	Point Point
	// Triggers stores stable trigger order.
	Triggers []*Node
	// Conditions stores stable condition order.
	Conditions []*Node
	// Effects stores stable effect order.
	Effects []*Node
	// Extras stores stable stack-selector and evaluation add-ons.
	Extras []*Node
	// Random selects one eligible effect.
	Random bool
	// Unseen selects effects in runtime round-robin order.
	Unseen bool
	// Or evaluates conditions with OR semantics.
	Or bool
}

// Generation stores an immutable compiled room graph.
type Generation struct {
	// ID identifies this graph revision.
	ID uint64
	// RoomID identifies the containing room.
	RoomID int64
	// Nodes stores compiled nodes by item id.
	Nodes map[int64]*Node
	// Stacks stores compiled stacks by tile.
	Stacks map[Point]*Stack
	// Triggers stores all triggers for indexed event matching.
	Triggers []*Node
}
