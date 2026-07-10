package audit

import (
	"context"
	"time"

	auditmodel "github.com/niflaot/pixels/internal/realm/room/audit/model"
	auditrepo "github.com/niflaot/pixels/internal/realm/room/audit/repository"
	bannedevent "github.com/niflaot/pixels/internal/realm/room/events/banned"
	kickedevent "github.com/niflaot/pixels/internal/realm/room/events/kicked"
	mutedevent "github.com/niflaot/pixels/internal/realm/room/events/muted"
	rightsgranted "github.com/niflaot/pixels/internal/realm/room/events/rightsgranted"
	rightsrevoked "github.com/niflaot/pixels/internal/realm/room/events/rightsrevoked"
	unbannedevent "github.com/niflaot/pixels/internal/realm/room/events/unbanned"
	unmutedevent "github.com/niflaot/pixels/internal/realm/room/events/unmuted"
	moderationmodel "github.com/niflaot/pixels/internal/realm/room/moderation/model"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/fx"
)

const (
	// actorPlayer identifies player-originated room actions.
	actorPlayer = "player"
)

// subscriptionHandler pairs one event name with its persistence handler.
type subscriptionHandler struct {
	// name identifies the subscribed event.
	name bus.Name
	// handler persists the event.
	handler bus.Handler
}

// RegisterSubscriber persists room mutation events inside their transaction scope.
func RegisterSubscriber(lifecycle fx.Lifecycle, subscriber bus.Subscriber, store auditrepo.Store) error {
	handlers := []subscriptionHandler{
		{name: rightsgranted.Name, handler: handleRightsGranted(store)},
		{name: rightsrevoked.Name, handler: handleRightsRevoked(store)},
		{name: kickedevent.Name, handler: handleKicked(store)},
		{name: mutedevent.Name, handler: handleMuted(store)},
		{name: unmutedevent.Name, handler: handleUnmuted(store)},
		{name: bannedevent.Name, handler: handleBanned(store)},
		{name: unbannedevent.Name, handler: handleUnbanned(store)},
	}
	subscriptions := make([]*bus.Subscription, 0, len(handlers))
	for _, item := range handlers {
		subscription, err := subscriber.Subscribe(item.name, bus.PriorityHigh, item.handler)
		if err != nil {
			unsubscribeAll(subscriptions)
			return err
		}
		subscriptions = append(subscriptions, subscription)
	}
	lifecycle.Append(fx.Hook{OnStop: func(context.Context) error {
		unsubscribeAll(subscriptions)
		return nil
	}})

	return nil
}

// handleRightsGranted persists a rights grant.
func handleRightsGranted(store auditrepo.Store) bus.Handler {
	return func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(rightsgranted.Payload)
		if !ok {
			return nil
		}
		actorID := payload.ActorID
		return store.InsertRights(ctx, auditmodel.RightsAudit{RoomID: payload.RoomID, PlayerID: payload.PlayerID, ActorKind: actorPlayer, ActorID: &actorID, Action: auditmodel.RightsGranted, CreatedAt: event.At})
	}
}

// handleRightsRevoked persists a rights revocation.
func handleRightsRevoked(store auditrepo.Store) bus.Handler {
	return func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(rightsrevoked.Payload)
		if !ok {
			return nil
		}
		actorID := payload.ActorID
		return store.InsertRights(ctx, auditmodel.RightsAudit{RoomID: payload.RoomID, PlayerID: payload.PlayerID, ActorKind: actorPlayer, ActorID: &actorID, Action: auditmodel.RightsAction(payload.Action), CreatedAt: event.At})
	}
}

// handleKicked persists a kick action.
func handleKicked(store auditrepo.Store) bus.Handler {
	return func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(kickedevent.Payload)
		if !ok {
			return nil
		}
		return insertModeration(ctx, store, event, payload.RoomID, payload.TargetPlayerID, payload.ActorID, moderationmodel.ActionKick, nil, nil)
	}
}

// handleMuted persists a mute action.
func handleMuted(store auditrepo.Store) bus.Handler {
	return func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(mutedevent.Payload)
		if !ok {
			return nil
		}
		return insertModeration(ctx, store, event, payload.RoomID, payload.TargetPlayerID, payload.ActorID, moderationmodel.ActionMute, &payload.DurationSeconds, &payload.ExpiresAt)
	}
}

// handleUnmuted persists an unmute action.
func handleUnmuted(store auditrepo.Store) bus.Handler {
	return func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(unmutedevent.Payload)
		if !ok {
			return nil
		}
		return insertModeration(ctx, store, event, payload.RoomID, payload.TargetPlayerID, payload.ActorID, moderationmodel.ActionUnmute, nil, nil)
	}
}

// handleBanned persists a ban action.
func handleBanned(store auditrepo.Store) bus.Handler {
	return func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(bannedevent.Payload)
		if !ok {
			return nil
		}
		return insertModeration(ctx, store, event, payload.RoomID, payload.TargetPlayerID, payload.ActorID, moderationmodel.ActionBan, &payload.DurationSeconds, &payload.ExpiresAt)
	}
}

// handleUnbanned persists an unban action.
func handleUnbanned(store auditrepo.Store) bus.Handler {
	return func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(unbannedevent.Payload)
		if !ok {
			return nil
		}
		return insertModeration(ctx, store, event, payload.RoomID, payload.TargetPlayerID, payload.ActorID, moderationmodel.ActionUnban, nil, nil)
	}
}

// insertModeration appends one normalized moderation record.
func insertModeration(ctx context.Context, store auditrepo.Store, event bus.Event, roomID int64, targetID int64, actorID int64, action moderationmodel.Action, duration *int64, expiresAt *time.Time) error {
	return store.InsertModeration(ctx, auditmodel.ModerationAction{RoomID: roomID, TargetPlayerID: targetID, ActorKind: actorPlayer, ActorID: &actorID, Action: action, DurationSeconds: duration, ExpiresAt: expiresAt, CreatedAt: event.At})
}

// unsubscribeAll removes event subscriptions.
func unsubscribeAll(subscriptions []*bus.Subscription) {
	for _, subscription := range subscriptions {
		subscription.Unsubscribe()
	}
}
