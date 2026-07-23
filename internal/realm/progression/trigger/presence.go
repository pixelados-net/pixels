package trigger

import (
	"context"
	"time"

	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	progressionconfig "github.com/niflaot/pixels/internal/realm/progression/config"
	"go.uber.org/fx"
)

// RegisterPresence owns periodic online-presence progression.
func RegisterPresence(lifecycle fx.Lifecycle, config progressionconfig.Config, players *playerlive.Registry, subscriber *Subscriber) {
	ctx, cancel := context.WithCancel(context.Background())
	lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go subscriber.runPresence(ctx, config.PresenceInterval, players)
			return nil
		},
		OnStop: func(context.Context) error {
			cancel()
			return nil
		},
	})
}

// runPresence advances bounded wall-clock minutes for currently online players.
func (subscriber *Subscriber) runPresence(ctx context.Context, interval time.Duration, players *playerlive.Registry) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	amount := int64(interval / time.Minute)
	if amount < 1 {
		amount = 1
	}
	for {
		select {
		case <-ticker.C:
			for _, player := range players.Snapshot() {
				subscriber.progress(ctx, player.ID(), "player.presence.minutes", amount, false)
			}
		case <-ctx.Done():
			return
		}
	}
}
