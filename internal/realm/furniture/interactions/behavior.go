// Package interactions resolves shared furniture interaction behavior.
package interactions

import (
	gatebehavior "github.com/niflaot/pixels/internal/realm/furniture/interactions/gate"
	togglebehavior "github.com/niflaot/pixels/internal/realm/furniture/interactions/toggle"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
)

// Behavior computes one state transition against active room state.
type Behavior interface {
	// Next returns the next state, whether fixtures change, and whether to commit.
	Next(*roomlive.Room, worldfurniture.Item) (string, bool, bool)
}

// Registry resolves the bounded set of generic state behaviors without a map lookup.
type Registry struct {
	// toggle cycles ordinary multi-state furniture.
	toggle togglebehavior.Behavior
	// gate changes binary gate state and walkability.
	gate gatebehavior.Behavior
}

// NewRegistry creates the generic furniture interaction registry.
func NewRegistry() *Registry {
	return &Registry{}
}

// Resolve returns behavior for one definition interaction type.
func (registry *Registry) Resolve(interactionType string) (Behavior, bool) {
	if registry == nil {
		return nil, false
	}
	switch interactionType {
	case "default", "toggle":
		return registry.toggle, true
	case "gate":
		return registry.gate, true
	default:
		return nil, false
	}
}
