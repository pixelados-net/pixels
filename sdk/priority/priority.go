// Package priority defines shared ordering for plugin callbacks.
package priority

// Priority controls interceptor and listener execution order.
type Priority int

const (
	// Lowest runs after ordinary low-priority callbacks.
	Lowest Priority = -200
	// Low runs below normal callbacks.
	Low Priority = -100
	// Normal is the default callback priority.
	Normal Priority = 0
	// High runs above normal callbacks.
	High Priority = 100
	// Highest runs before ordinary high-priority callbacks.
	Highest Priority = 200
	// Monitor is the conventional final observation-only priority.
	Monitor Priority = -1000
)
