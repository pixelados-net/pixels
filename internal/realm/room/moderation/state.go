package moderation

import (
	"context"

	unbannedevent "github.com/niflaot/pixels/internal/realm/room/events/unbanned"
	unmutedevent "github.com/niflaot/pixels/internal/realm/room/events/unmuted"
	moderationmodel "github.com/niflaot/pixels/internal/realm/room/moderation/model"
	"github.com/niflaot/pixels/pkg/bus"
)

// IsBanned reports whether a room ban is active.
func (service *Service) IsBanned(ctx context.Context, roomID int64, playerID int64) (bool, error) {
	if roomID <= 0 || playerID <= 0 {
		return false, nil
	}

	return service.store.IsBanned(ctx, roomID, playerID, service.now())
}

// IsMuted reports whether a room mute is active.
func (service *Service) IsMuted(ctx context.Context, roomID int64, playerID int64) (bool, error) {
	if roomID <= 0 || playerID <= 0 {
		return false, nil
	}

	return service.store.IsMuted(ctx, roomID, playerID, service.now())
}

// ListBans lists active room bans.
func (service *Service) ListBans(ctx context.Context, roomID int64) ([]moderationmodel.Sanction, error) {
	if roomID <= 0 {
		return nil, ErrInvalidIdentity
	}

	return service.store.ListBans(ctx, roomID, service.now())
}

// ListMutes lists active room mutes.
func (service *Service) ListMutes(ctx context.Context, roomID int64) ([]moderationmodel.Sanction, error) {
	if roomID <= 0 {
		return nil, ErrInvalidIdentity
	}

	return service.store.ListMutes(ctx, roomID, service.now())
}

// end expires one current sanction and publishes its explicit reversal.
func (service *Service) end(ctx context.Context, roomID int64, actorID int64, targetID int64, action moderationmodel.Action) error {
	return service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		now := service.now()
		var removed bool
		var err error
		if action == moderationmodel.ActionUnmute {
			removed, err = service.store.Unmute(txCtx, roomID, targetID, now)
		} else {
			removed, err = service.store.Unban(txCtx, roomID, targetID, now)
		}
		if err != nil || !removed {
			return err
		}
		if action == moderationmodel.ActionUnmute {
			return service.publish(txCtx, unmutedevent.Name, unmutedevent.Payload{RoomID: roomID, TargetPlayerID: targetID, ActorID: actorID})
		}

		return service.publish(txCtx, unbannedevent.Name, unbannedevent.Payload{RoomID: roomID, TargetPlayerID: targetID, ActorID: actorID})
	})
}

// publish emits one moderation mutation event.
func (service *Service) publish(ctx context.Context, name bus.Name, payload any) error {
	if service.events == nil {
		return nil
	}

	return service.events.Publish(ctx, bus.Event{Name: name, Payload: payload})
}
