// Package sessioncompleted defines completed guide feedback events.
package sessioncompleted

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies one completed guide session with feedback.
const Name bus.Name = "guide.session.completed"

// Payload describes completed guide feedback.
type Payload struct {
	// GuideID identifies the assisting guide.
	GuideID int64
	// RequesterID identifies the player receiving help.
	RequesterID int64
	// Feedback reports whether the guide was recommended.
	Feedback bool
}
