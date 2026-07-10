package moderation

import (
	"context"
	"time"

	"github.com/niflaot/pixels/internal/permission"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	bannedevent "github.com/niflaot/pixels/internal/realm/room/events/banned"
	kickedevent "github.com/niflaot/pixels/internal/realm/room/events/kicked"
	mutedevent "github.com/niflaot/pixels/internal/realm/room/events/muted"
	moderationmodel "github.com/niflaot/pixels/internal/realm/room/moderation/model"
	moderationrepo "github.com/niflaot/pixels/internal/realm/room/moderation/repository"
	"github.com/niflaot/pixels/pkg/bus"
)

// Nodes stores room moderation permission nodes.
type Nodes struct {
	// OwnKick allows locally authorized kicking.
	OwnKick permission.Node
	// OwnMute allows locally authorized muting.
	OwnMute permission.Node
	// OwnBan allows locally authorized banning.
	OwnBan permission.Node
	// AnyKick allows staff kicking in any room.
	AnyKick permission.Node
	// AnyMute allows staff muting in any room.
	AnyMute permission.Node
	// AnyBan allows staff banning in any room.
	AnyBan permission.Node
	// Unkickable protects moderation targets.
	Unkickable permission.Node
}

// Service coordinates room moderation policy and state.
type Service struct {
	// config stores normalized moderation limits.
	config Config
	// store persists current moderation state.
	store moderationrepo.Store
	// rooms resolves durable room policy.
	rooms RoomFinder
	// rights resolves room-scoped capabilities.
	rights RightsChecker
	// permissions resolves global capability nodes.
	permissions permissionservice.Checker
	// events publishes committed moderation changes.
	events bus.Publisher
	// nodes stores moderation capability nodes.
	nodes Nodes
	// now returns current time for deterministic behavior.
	now func() time.Time
}

// New creates a room moderation service.
func New(config Config, store moderationrepo.Store, rooms RoomFinder, rights RightsChecker, permissions permissionservice.Checker, events bus.Publisher, nodes Nodes) *Service {
	return &Service{config: config.Normalize(), store: store, rooms: rooms, rights: rights, permissions: permissions, events: events, nodes: nodes, now: time.Now}
}

// Kick commits an immediate room kick action.
func (service *Service) Kick(ctx context.Context, roomID int64, actorID int64, targetID int64) error {
	if _, err := service.authorize(ctx, roomID, actorID, targetID, moderationmodel.ActionKick); err != nil {
		return err
	}

	return service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		return service.publish(txCtx, kickedevent.Name, kickedevent.Payload{RoomID: roomID, TargetPlayerID: targetID, ActorID: actorID})
	})
}

// Mute creates or replaces a room mute.
func (service *Service) Mute(ctx context.Context, roomID int64, actorID int64, targetID int64, minutes int32) error {
	if minutes < service.config.MinMuteMinutes || minutes > service.config.MaxMuteMinutes {
		return ErrInvalidMuteDuration
	}
	if _, err := service.authorize(ctx, roomID, actorID, targetID, moderationmodel.ActionMute); err != nil {
		return err
	}
	duration := time.Duration(minutes) * time.Minute
	expiresAt := service.now().Add(duration)

	return service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := service.store.Mute(txCtx, roomID, targetID, expiresAt); err != nil {
			return err
		}
		payload := mutedevent.Payload{RoomID: roomID, TargetPlayerID: targetID, ActorID: actorID, DurationSeconds: int64(duration / time.Second), ExpiresAt: expiresAt}

		return service.publish(txCtx, mutedevent.Name, payload)
	})
}

// Unmute ends an active room mute.
func (service *Service) Unmute(ctx context.Context, roomID int64, actorID int64, targetID int64) error {
	if _, err := service.authorize(ctx, roomID, actorID, targetID, moderationmodel.ActionMute); err != nil {
		return err
	}

	return service.end(ctx, roomID, actorID, targetID, moderationmodel.ActionUnmute)
}

// Ban creates or replaces a room ban.
func (service *Service) Ban(ctx context.Context, roomID int64, actorID int64, targetID int64, banDuration moderationmodel.BanDuration) error {
	duration, valid := banDuration.Duration()
	if !valid {
		return ErrInvalidBanDuration
	}
	if _, err := service.authorize(ctx, roomID, actorID, targetID, moderationmodel.ActionBan); err != nil {
		return err
	}
	expiresAt := service.now().Add(duration)

	return service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := service.store.Ban(txCtx, roomID, targetID, expiresAt); err != nil {
			return err
		}
		payload := bannedevent.Payload{RoomID: roomID, TargetPlayerID: targetID, ActorID: actorID, DurationSeconds: int64(duration / time.Second), ExpiresAt: expiresAt}

		return service.publish(txCtx, bannedevent.Name, payload)
	})
}

// Unban ends an active room ban.
func (service *Service) Unban(ctx context.Context, roomID int64, actorID int64, targetID int64) error {
	if _, err := service.authorize(ctx, roomID, actorID, targetID, moderationmodel.ActionBan); err != nil {
		return err
	}

	return service.end(ctx, roomID, actorID, targetID, moderationmodel.ActionUnban)
}
