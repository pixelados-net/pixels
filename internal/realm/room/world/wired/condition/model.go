// Package condition evaluates compiled WIRED conditions against room state.
package condition

import (
	"time"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
)

// Context stores condition evaluation inputs.
type Context struct {
	// Event stores the trigger event.
	Event trigger.Event
	// Now stores the injected wall and monotonic clock value.
	Now time.Time
	// ResetAt stores the most recent room timer reset.
	ResetAt time.Time
	// Effects stores the current stack's immutable effects for compatibility simulation.
	Effects []*configuration.Node
}

// View exposes read-only room facts required by canonical conditions.
type View interface {
	// UserCount returns player occupancy excluding bots and pets.
	UserCount() int
	// UnitOn reports whether any allowed room unit occupies furniture.
	UnitOn(int64) (bool, error)
	// ActorOn reports whether the event actor occupies furniture.
	ActorOn(trigger.Event, int64) (bool, bool, error)
	// Stacked reports whether furniture is stacked on a base target.
	Stacked(int64) (bool, error)
	// SnapshotMatches compares live furniture with one captured snapshot.
	SnapshotMatches(int64, record.Snapshot, []int32) (bool, error)
	// ActorTeam reports player team membership.
	ActorTeam(int64, int32) (bool, bool, error)
	// ActorGroup reports social group membership in the room group.
	ActorGroup(int64) (bool, bool, error)
	// WearingBadge reports whether a player equips a badge code.
	WearingBadge(int64, string) (bool, bool, error)
	// WearingEffect reports whether a player's active room effect matches.
	WearingEffect(int64, int32) (bool, bool, error)
	// HasHanditem reports whether a player's current hand item matches.
	HasHanditem(int64, int32) (bool, bool, error)
	// ValidMoves simulates the stack's movement effects without durable mutation.
	ValidMoves([]*configuration.Node, trigger.Event) (bool, error)
}

// Result stores a condition result and whether its predicate domain existed.
type Result struct {
	// Pass reports whether the stack may continue.
	Pass bool
	// Valid reports whether the actor/context domain existed.
	Valid bool
}

// Evaluator evaluates all canonical condition descriptors.
type Evaluator struct{}

// New creates a stateless condition evaluator.
func New() Evaluator { return Evaluator{} }
