package trigger

import (
	"context"

	roomstaffpicked "github.com/niflaot/pixels/internal/realm/room/record/events/staffpicked"
	subscriptionpayday "github.com/niflaot/pixels/internal/realm/subscription/events/payday"
	"github.com/niflaot/pixels/pkg/bus"
)

// subscriptionPayday advances one durable club month boundary.
func (subscriber *Subscriber) subscriptionPayday(ctx context.Context, event bus.Event) error {
	if payload, ok := event.Payload.(subscriptionpayday.Payload); ok {
		subscriber.progress(ctx, payload.PlayerID, "subscription.hc.month", 1, false)
	}
	return nil
}

// staffPicked advances the selected room owner's recognition counter.
func (subscriber *Subscriber) staffPicked(ctx context.Context, event bus.Event) error {
	if payload, ok := event.Payload.(roomstaffpicked.Payload); ok {
		subscriber.progress(ctx, payload.OwnerPlayerID, "staffpick.received", 1, false)
	}
	return nil
}
