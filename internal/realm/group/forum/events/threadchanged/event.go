// Package threadchanged contains the group-forum thread changed event.
package threadchanged

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies a committed forum-thread change.
const Name bus.Name = "group.forum.thread.changed"

// Payload identifies the committed thread generation.
type Payload struct {
	// GroupID identifies the forum group.
	GroupID int64
	// ThreadID identifies the changed thread.
	ThreadID int64
	// Version stores the committed thread version.
	Version int64
	// Action identifies creation or moderation.
	Action string
}
