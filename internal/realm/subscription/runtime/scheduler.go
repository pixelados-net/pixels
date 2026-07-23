// Package runtime owns subscription process lifecycle.
package runtime

import (
	"context"
	"time"

	playerconnected "github.com/niflaot/pixels/internal/realm/player/events/connected"
	"github.com/niflaot/pixels/internal/realm/subscription/core"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Scheduler runs global subscription lifecycle work.
type Scheduler struct {
	// interval stores cycle frequency.
	interval time.Duration
	// service runs subscription calculations.
	service cycleRunner
	// log records cycle failures.
	log *zap.Logger
	// cancel stops the active loop.
	cancel context.CancelFunc
}

// cycleRunner executes one durable subscription reconciliation pass.
type cycleRunner interface {
	// RunCycle reconciles all memberships due at the current instant.
	RunCycle(context.Context) error
}

// NewScheduler creates a subscription scheduler.
func NewScheduler(interval time.Duration, service *core.Service, log *zap.Logger) *Scheduler {
	if log == nil {
		log = zap.NewNop()
	}
	return &Scheduler{interval: interval, service: service, log: log}
}

// RegisterScheduler starts and stops the subscription scheduler with Fx.
func RegisterScheduler(lifecycle fx.Lifecycle, scheduler *Scheduler) {
	lifecycle.Append(fx.Hook{OnStart: scheduler.Start, OnStop: scheduler.Stop})
}

// Start launches the subscription cycle loop.
func (scheduler *Scheduler) Start(ctx context.Context) error {
	if err := scheduler.service.RunCycle(ctx); err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	scheduler.cancel = cancel
	go scheduler.run(ctx)
	return nil
}

// Stop stops the subscription cycle loop.
func (scheduler *Scheduler) Stop(context.Context) error {
	if scheduler.cancel != nil {
		scheduler.cancel()
	}
	return nil
}

// run executes subscription cycles until cancellation.
func (scheduler *Scheduler) run(ctx context.Context) {
	ticker := time.NewTicker(scheduler.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := scheduler.service.RunCycle(ctx); err != nil {
				scheduler.log.Error("subscription scheduler cycle failed", zap.Error(err))
			}
		}
	}
}

// RegisterPaydayClaims subscribes login-time pending reward delivery.
func RegisterPaydayClaims(lifecycle fx.Lifecycle, subscriber bus.Subscriber, service *core.Service) error {
	subscription, err := subscriber.Subscribe(playerconnected.Name, bus.PriorityNormal, func(ctx context.Context, event bus.Event) error {
		payload, ok := event.Payload.(playerconnected.Payload)
		if !ok || payload.PlayerID <= 0 {
			return nil
		}
		return service.ClaimPaydays(ctx, payload.PlayerID)
	})
	if err != nil {
		return err
	}
	lifecycle.Append(fx.Hook{OnStop: func(context.Context) error {
		subscription.Unsubscribe()
		return nil
	}})
	return nil
}
