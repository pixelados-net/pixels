package history

import (
	"context"
	"sync"
	"time"

	chatconfig "github.com/niflaot/pixels/internal/realm/chat/config"
	shoutedevent "github.com/niflaot/pixels/internal/realm/chat/events/shouted"
	talkedevent "github.com/niflaot/pixels/internal/realm/chat/events/talked"
	whisperedevent "github.com/niflaot/pixels/internal/realm/chat/events/whispered"
	historymodel "github.com/niflaot/pixels/internal/realm/chat/history/model"
	historyrepo "github.com/niflaot/pixels/internal/realm/chat/history/repository"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Register subscribes message history and manages writer and partition lifecycles.
func Register(lifecycle fx.Lifecycle, subscriber bus.Subscriber, writer *Writer, store historyrepo.Store, config chatconfig.Config, log *zap.Logger) error {
	config = config.Normalize()
	subscriptions := make([]*bus.Subscription, 0, 3)
	register := func(name bus.Name, handler bus.Handler) error {
		subscription, err := subscriber.Subscribe(name, bus.PriorityLow, handler)
		if err == nil {
			subscriptions = append(subscriptions, subscription)
		}
		return err
	}
	if err := register(talkedevent.Name, talkHandler(writer)); err != nil {
		return err
	}
	if err := register(shoutedevent.Name, shoutHandler(writer)); err != nil {
		return err
	}
	if config.LogWhispers {
		if err := register(whisperedevent.Name, whisperHandler(writer)); err != nil {
			return err
		}
	}
	var cancel context.CancelFunc
	var group sync.WaitGroup
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			now := time.Now().UTC()
			if err := maintain(ctx, store, config, now); err != nil {
				return err
			}
			workerCtx, workerCancel := context.WithCancel(context.Background())
			cancel = workerCancel
			group.Add(2)
			go func() { defer group.Done(); writer.run(workerCtx) }()
			go func() { defer group.Done(); maintenance(workerCtx, store, config, log) }()
			return nil
		},
		OnStop: func(context.Context) error {
			for _, subscription := range subscriptions {
				subscription.Unsubscribe()
			}
			if cancel != nil {
				cancel()
				group.Wait()
			}
			return nil
		},
	})

	return nil
}

// talkHandler converts delivered talks to history entries.
func talkHandler(writer *Writer) bus.Handler {
	return func(_ context.Context, event bus.Event) error {
		payload, ok := event.Payload.(talkedevent.Payload)
		if ok {
			writer.Enqueue(historymodel.Entry{RoomID: payload.RoomID, PlayerID: payload.PlayerID, Kind: "talk", Message: payload.Message, Censored: payload.Censored, CreatedAt: payload.CreatedAt})
		}
		return nil
	}
}

// shoutHandler converts delivered shouts to history entries.
func shoutHandler(writer *Writer) bus.Handler {
	return func(_ context.Context, event bus.Event) error {
		payload, ok := event.Payload.(shoutedevent.Payload)
		if ok {
			writer.Enqueue(historymodel.Entry{RoomID: payload.RoomID, PlayerID: payload.PlayerID, Kind: "shout", Message: payload.Message, Censored: payload.Censored, CreatedAt: payload.CreatedAt})
		}
		return nil
	}
}

// whisperHandler converts opted-in whispers to history entries.
func whisperHandler(writer *Writer) bus.Handler {
	return func(_ context.Context, event bus.Event) error {
		payload, ok := event.Payload.(whisperedevent.Payload)
		if ok {
			target := payload.TargetPlayerID
			writer.Enqueue(historymodel.Entry{RoomID: payload.RoomID, PlayerID: payload.PlayerID, TargetPlayerID: &target, Kind: "whisper", Message: payload.Message, Censored: payload.Censored, CreatedAt: payload.CreatedAt})
		}
		return nil
	}
}

// maintenance periodically creates and retires daily partitions.
func maintenance(ctx context.Context, store historyrepo.Store, config chatconfig.Config, log *zap.Logger) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		select {
		case now := <-ticker.C:
			if err := maintain(ctx, store, config, now); err != nil {
				log.Error("chat history partition maintenance failed", zap.Error(err), zap.Time("at", now))
			}
		case <-ctx.Done():
			return
		}
	}
}

// maintain ensures future partitions and drops expired history partitions.
func maintain(ctx context.Context, store historyrepo.Store, config chatconfig.Config, now time.Time) error {
	if err := store.EnsurePartitions(ctx, now, now.AddDate(0, 0, 2)); err != nil {
		return err
	}

	return store.DropBefore(ctx, now.AddDate(0, 0, -config.HistoryRetentionDays))
}
