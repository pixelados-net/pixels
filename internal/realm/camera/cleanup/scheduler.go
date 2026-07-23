package cleanup

import (
	"context"
	"time"

	cameraconfig "github.com/niflaot/pixels/internal/realm/camera/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Scheduler owns the single aggregate cleanup loop.
type Scheduler struct {
	// config stores the reconciliation cadence.
	config cameraconfig.Config
	// service performs bounded cleanup sweeps.
	service *Service
	// log records aggregate reconciliation failures.
	log *zap.Logger
	// cancel stops the active loop.
	cancel context.CancelFunc
	// done closes after the active loop exits.
	done chan struct{}
}

// NewScheduler creates the aggregate camera cleanup scheduler.
func NewScheduler(config cameraconfig.Config, service *Service, log *zap.Logger) *Scheduler {
	return &Scheduler{config: config, service: service, log: log}
}

// RegisterScheduler binds camera cleanup to application lifecycle.
func RegisterScheduler(lifecycle fx.Lifecycle, scheduler *Scheduler) {
	lifecycle.Append(fx.Hook{OnStart: scheduler.Start, OnStop: scheduler.Stop})
}

// Start launches one process-wide cleanup loop.
func (scheduler *Scheduler) Start(context.Context) error {
	ctx, cancel := context.WithCancel(context.Background())
	scheduler.cancel = cancel
	scheduler.done = make(chan struct{})
	go scheduler.run(ctx)
	return nil
}

// Stop terminates the cleanup loop.
func (scheduler *Scheduler) Stop(ctx context.Context) error {
	if scheduler.cancel != nil {
		scheduler.cancel()
	}
	if scheduler.done == nil {
		return nil
	}
	select {
	case <-scheduler.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// run executes bounded cleanup sweeps until cancellation.
func (scheduler *Scheduler) run(ctx context.Context) {
	defer close(scheduler.done)
	ticker := time.NewTicker(scheduler.config.CleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, err := scheduler.service.Sweep(ctx); err != nil && scheduler.log != nil {
				scheduler.log.Error("camera cleanup sweep failed", zap.Error(err))
			}
		}
	}
}
