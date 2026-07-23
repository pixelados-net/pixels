package pickup

import (
	"context"

	pickedupevent "github.com/niflaot/pixels/internal/realm/furniture/events/pickedup"
	"github.com/niflaot/pixels/pkg/bus"
)

// publish emits furniture pickup completion.
func (handler Handler) publish(ctx context.Context, playerID int64, itemID int64, roomID int64) error {
	if handler.Events == nil {
		return nil
	}

	return handler.Events.Publish(ctx, bus.Event{
		Name:    pickedupevent.Name,
		Payload: pickedupevent.Payload{PlayerID: playerID, ItemID: itemID, RoomID: roomID},
	})
}
