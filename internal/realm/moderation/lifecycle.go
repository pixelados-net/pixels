package moderation

import (
	"context"
	"sync"
	"time"

	"github.com/niflaot/pixels/internal/realm/moderation/cfh"
	moderationcore "github.com/niflaot/pixels/internal/realm/moderation/core"
	issueclosed "github.com/niflaot/pixels/internal/realm/moderation/events/issueclosed"
	issuecreated "github.com/niflaot/pixels/internal/realm/moderation/events/issuecreated"
	"github.com/niflaot/pixels/internal/realm/moderation/guardian"
	guardianhandlers "github.com/niflaot/pixels/internal/realm/moderation/guardian/handlers"
	guidehandlers "github.com/niflaot/pixels/internal/realm/moderation/guide/handlers"
	moderationruntime "github.com/niflaot/pixels/internal/realm/moderation/runtime"
	"github.com/niflaot/pixels/internal/realm/moderation/staff"
	playerconnected "github.com/niflaot/pixels/internal/realm/player/events/connected"
	playerdisconnected "github.com/niflaot/pixels/internal/realm/player/events/disconnected"
	sanctionapplied "github.com/niflaot/pixels/internal/realm/sanction/events/applied"
	sanctionrevoked "github.com/niflaot/pixels/internal/realm/sanction/events/revoked"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/fx"
)

// RegisterLifecycle warms topics, projects login state, and owns one guardian ticker.
func RegisterLifecycle(lifecycle fx.Lifecycle, subscriber bus.Subscriber, service *moderationcore.Service, manager *guardian.Manager, runtime *moderationruntime.Context) error {
	connected, err := subscriber.Subscribe(playerconnected.Name, bus.PriorityNormal, func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(playerconnected.Payload)
		if !ok {
			return nil
		}
		if err := cfh.Bootstrap(ctx, runtime, payload.PlayerID); err != nil {
			return err
		}
		return staff.Bootstrap(ctx, runtime, payload.PlayerID)
	})
	if err != nil {
		return err
	}
	disconnected, err := subscriber.Subscribe(playerdisconnected.Name, bus.PriorityNormal, func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(playerdisconnected.Payload)
		if !ok {
			return nil
		}
		if err := guidehandlers.Disconnected(ctx, runtime, payload.PlayerID); err != nil {
			return err
		}
		return guardianhandlers.Disconnected(ctx, runtime, payload.PlayerID)
	})
	if err != nil {
		connected.Unsubscribe()
		return err
	}
	applied, err := subscriber.Subscribe(sanctionapplied.Name, bus.PriorityNormal, func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(sanctionapplied.Payload)
		if !ok {
			return nil
		}
		return cfh.RefreshStatus(ctx, runtime, payload.ReceiverID)
	})
	if err != nil {
		connected.Unsubscribe()
		disconnected.Unsubscribe()
		return err
	}
	revoked, err := subscriber.Subscribe(sanctionrevoked.Name, bus.PriorityNormal, func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(sanctionrevoked.Payload)
		if !ok {
			return nil
		}
		return cfh.RefreshStatus(ctx, runtime, payload.ReceiverID)
	})
	if err != nil {
		connected.Unsubscribe()
		disconnected.Unsubscribe()
		applied.Unsubscribe()
		return err
	}
	created, err := subscriber.Subscribe(issuecreated.Name, bus.PriorityNormal, func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(issuecreated.Payload)
		if !ok {
			return nil
		}
		return (staff.Handler{Context: runtime}).BroadcastIssue(ctx, payload.IssueID)
	})
	if err != nil {
		connected.Unsubscribe()
		disconnected.Unsubscribe()
		applied.Unsubscribe()
		revoked.Unsubscribe()
		return err
	}
	closed, err := subscriber.Subscribe(issueclosed.Name, bus.PriorityNormal, func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(issueclosed.Payload)
		if !ok {
			return nil
		}
		return (staff.Handler{Context: runtime}).BroadcastIssueDeleted(ctx, payload.IssueID)
	})
	if err != nil {
		connected.Unsubscribe()
		disconnected.Unsubscribe()
		applied.Unsubscribe()
		revoked.Unsubscribe()
		created.Unsubscribe()
		return err
	}
	var cancel context.CancelFunc
	var workers sync.WaitGroup
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := service.RefreshTopics(ctx); err != nil {
				return err
			}
			workerContext, stop := context.WithCancel(context.Background())
			cancel = stop
			workers.Add(1)
			go tickGuardians(workerContext, &workers, manager, runtime)
			return nil
		},
		OnStop: func(context.Context) error {
			connected.Unsubscribe()
			disconnected.Unsubscribe()
			applied.Unsubscribe()
			revoked.Unsubscribe()
			created.Unsubscribe()
			closed.Unsubscribe()
			if cancel != nil {
				cancel()
			}
			workers.Wait()
			return nil
		},
	})
	return nil
}

// tickGuardians closes expired peer reviews from one realm-owned worker.
func tickGuardians(ctx context.Context, workers *sync.WaitGroup, manager *guardian.Manager, runtime *moderationruntime.Context) {
	defer workers.Done()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			for _, ticket := range manager.Tick(ctx, now) {
				_ = (guardianhandlers.Handler{Context: runtime}).Finish(ctx, ticket)
			}
		}
	}
}
