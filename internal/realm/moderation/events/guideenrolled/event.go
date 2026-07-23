// Package guideenrolled defines authorized guide duty enrollment events.
package guideenrolled

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies one authorized guide duty enrollment.
const Name bus.Name = "guide.enrolled"

// Payload describes one guide entering a helper pool.
type Payload struct {
	// PlayerID identifies the enrolled guide.
	PlayerID int64
}
