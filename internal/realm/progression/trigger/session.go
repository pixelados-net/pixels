package trigger

import (
	"context"
	"time"

	playerignored "github.com/niflaot/pixels/internal/realm/messenger/session/events/ignored"
	playerauthenticated "github.com/niflaot/pixels/internal/realm/player/events/authenticated"
	playerdisconnected "github.com/niflaot/pixels/internal/realm/player/events/disconnected"
	roomkicked "github.com/niflaot/pixels/internal/realm/room/control/events/kicked"
	roommuted "github.com/niflaot/pixels/internal/realm/room/control/events/muted"
	roomsettings "github.com/niflaot/pixels/internal/realm/room/control/events/settingsupdated"
	roomwordfilter "github.com/niflaot/pixels/internal/realm/room/control/events/wordfiltermodified"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/zap"
)

// authenticated advances login and account-age achievements.
func (subscriber *Subscriber) authenticated(ctx context.Context, event bus.Event) error {
	payload, ok := event.Payload.(playerauthenticated.Payload)
	if !ok {
		return nil
	}
	if err := subscriber.engine.HydratePlayer(ctx, payload.PlayerID); err != nil {
		subscriber.log.Warn("progression hydration failed", zap.Int64("player_id", payload.PlayerID), zap.Error(err))
	}
	subscriber.progress(ctx, payload.PlayerID, "player.login.streak", 1, true)
	if subscriber.players == nil {
		return nil
	}
	player, found := subscriber.players.Find(payload.PlayerID)
	if !found {
		return nil
	}
	created := player.Snapshot().CreatedAt
	if created.IsZero() {
		return nil
	}
	days := int64(time.Since(created).Hours() / 24)
	if days < 0 {
		days = 0
	}
	if err := subscriber.engine.SetTriggerProgress(ctx, payload.PlayerID, "player.registration.days", days); err != nil {
		subscriber.log.Warn("registration progression failed", zap.Int64("player_id", payload.PlayerID), zap.Error(err))
	}
	return nil
}

// playerIgnored records the ignore self-moderation interaction.
func (subscriber *Subscriber) playerIgnored(ctx context.Context, event bus.Event) error {
	if payload, ok := event.Payload.(playerignored.Payload); ok {
		subscriber.progress(ctx, payload.PlayerID, "selfmod.ignore", 1, false)
	}
	return nil
}

// selfModSettings records the five room-settings help surfaces represented by one editor.
func (subscriber *Subscriber) selfModSettings(ctx context.Context, event bus.Event) error {
	if payload, ok := event.Payload.(roomsettings.Payload); ok {
		for _, key := range []string{"selfmod.door_mode", "selfmod.walkthrough", "selfmod.chat_scroll_speed", "selfmod.chat_hear_range", "selfmod.chat_flood_filter"} {
			subscriber.progress(ctx, payload.ActorID, key, 1, false)
		}
	}
	return nil
}

// selfModWordFilter records room word-filter administration.
func (subscriber *Subscriber) selfModWordFilter(ctx context.Context, event bus.Event) error {
	if payload, ok := event.Payload.(roomwordfilter.Payload); ok {
		subscriber.progress(ctx, payload.ActorID, "selfmod.room_filter", 1, false)
	}
	return nil
}

// selfModMuted records room mute administration.
func (subscriber *Subscriber) selfModMuted(ctx context.Context, event bus.Event) error {
	if payload, ok := event.Payload.(roommuted.Payload); ok {
		subscriber.progress(ctx, payload.ActorID, "selfmod.mute", 1, false)
	}
	return nil
}

// selfModKicked records room kick administration.
func (subscriber *Subscriber) selfModKicked(ctx context.Context, event bus.Event) error {
	if payload, ok := event.Payload.(roomkicked.Payload); ok {
		subscriber.progress(ctx, payload.ActorID, "selfmod.kick", 1, false)
	}
	return nil
}

// disconnected flushes one player's pending write-behind deltas.
func (subscriber *Subscriber) disconnected(ctx context.Context, event bus.Event) error {
	if payload, ok := event.Payload.(playerdisconnected.Payload); ok {
		if err := subscriber.engine.FlushPlayer(ctx, payload.PlayerID); err != nil {
			subscriber.log.Warn("progression disconnect flush failed", zap.Int64("player_id", payload.PlayerID), zap.Error(err))
		}
		subscriber.engine.ForgetPlayer(payload.PlayerID)
	}
	return nil
}
