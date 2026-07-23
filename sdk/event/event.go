// Package event contains typed events exposed to dynamic plugins.
package event

import (
	"context"

	sdkpriority "github.com/niflaot/pixels/sdk/priority"
)

// Event identifies one plugin-facing event.
type Event interface {
	// Name returns the stable event identifier.
	Name() string
}

// Cancellable describes an event whose default action may be vetoed.
type Cancellable interface {
	Event
	// Cancelled reports whether a listener vetoed the default action.
	Cancelled() bool
	// SetCancelled changes whether the default action is vetoed.
	SetCancelled(bool)
}

// Listener reacts to one dispatched plugin-facing event.
type Listener func(context.Context, Event) error

// ListenerOptions configures event listener ordering and cancellation behavior.
type ListenerOptions struct {
	// Priority runs larger values before smaller values.
	Priority sdkpriority.Priority
	// IgnoreCancelled skips this listener after an earlier listener cancels.
	IgnoreCancelled bool
}
