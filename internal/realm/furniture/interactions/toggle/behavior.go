// Package toggle implements generic multi-state furniture behavior.
package toggle

import (
	"strconv"

	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
)

// Behavior cycles generic furniture through its declared visual states.
type Behavior struct{}

// Next resolves the next generic furniture state.
func (Behavior) Next(_ *roomlive.Room, item worldfurniture.Item) (string, bool, bool) {
	modes := item.Definition.InteractionModesCount
	if modes <= 1 {
		return item.ExtraData, false, false
	}
	current, err := strconv.Atoi(item.ExtraData)
	if err != nil || current < 0 || current >= modes {
		current = 0
	}

	return strconv.Itoa((current + 1) % modes), false, true
}
