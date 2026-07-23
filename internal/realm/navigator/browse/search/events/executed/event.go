// Package searchexecuted contains the navigator search executed event.
package searchexecuted

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies the navigator search executed event.
const Name bus.Name = "navigator.search_executed"

// Payload describes a navigator search event.
type Payload struct {
	// PlayerID identifies the player.
	PlayerID int64

	// Code stores the search context or result code.
	Code string

	// Query stores the search query.
	Query string

	// Count stores the result count.
	Count int
}
