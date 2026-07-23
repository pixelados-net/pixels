package trigger

import (
	"context"

	gameprogressed "github.com/niflaot/pixels/internal/realm/room/world/games/events/progressed"
	"github.com/niflaot/pixels/pkg/bus"
)

// gameProgressed maps one committed room game delta.
func (subscriber *Subscriber) gameProgressed(ctx context.Context, event bus.Event) error {
	payload, ok := event.Payload.(gameprogressed.Payload)
	if ok && payload.PlayerID > 0 && payload.Key != "" && payload.Amount > 0 {
		subscriber.progress(ctx, payload.PlayerID, payload.Key, payload.Amount, false)
	}
	return nil
}
