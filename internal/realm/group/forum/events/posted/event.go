// Package posted contains the group-forum post committed event.
package posted

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies a committed forum post.
const Name bus.Name = "group.forum.posted"

// Payload identifies the committed post generation.
type Payload struct {
	// GroupID identifies the forum group.
	GroupID int64
	// ThreadID identifies the parent thread.
	ThreadID int64
	// PostID identifies the committed post.
	PostID int64
	// Version stores the committed post version.
	Version int64
}
