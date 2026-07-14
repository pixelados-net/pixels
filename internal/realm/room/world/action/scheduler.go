package action

import (
	"context"
	"time"

	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	netconn "github.com/niflaot/pixels/networking/connection"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// activityConnection exposes connection input activity without widening the stable interface.
type activityConnection interface {
	// LastActivityAt returns the latest inbound packet time.
	LastActivityAt() time.Time
}

// Scheduler reconciles automatic AFK projection in one process loop.
type Scheduler struct {
	// config controls idle timing.
	config Config
	// runtime stores active rooms.
	runtime *roomlive.Registry
	// connections stores active transports.
	connections *netconn.Registry
	// actions changes room projections.
	actions *Service
	// log records reconciliation failures.
	log *zap.Logger
	// cancel stops the active loop.
	cancel context.CancelFunc
}

// NewScheduler creates the automatic idle scheduler.
func NewScheduler(config Config, runtime *roomlive.Registry, connections *netconn.Registry, actions *Service, log *zap.Logger) *Scheduler {
	return &Scheduler{config: config.Normalize(), runtime: runtime, connections: connections, actions: actions, log: log}
}

// RegisterScheduler binds the automatic idle scheduler to Fx lifecycle.
func RegisterScheduler(lifecycle fx.Lifecycle, scheduler *Scheduler) {
	lifecycle.Append(fx.Hook{OnStart: scheduler.Start, OnStop: scheduler.Stop})
}

// Start launches automatic idle reconciliation.
func (scheduler *Scheduler) Start(context.Context) error {
	ctx, cancel := context.WithCancel(context.Background())
	scheduler.cancel = cancel
	go scheduler.run(ctx)
	return nil
}

// Stop stops automatic idle reconciliation.
func (scheduler *Scheduler) Stop(context.Context) error {
	if scheduler.cancel != nil {
		scheduler.cancel()
	}
	return nil
}

// Sweep reconciles every active player against one time instant.
func (scheduler *Scheduler) Sweep(ctx context.Context, now time.Time) error {
	for _, active := range scheduler.runtime.Snapshot() {
		for _, presence := range active.Presences() {
			connection, found := scheduler.connections.Get(presence.Occupant.ConnectionKind, presence.Occupant.ConnectionID)
			activity, ok := connection.(activityConnection)
			if !found || !ok {
				continue
			}
			idle := idleState(presence.Unit, activity.LastActivityAt(), now, scheduler.config.IdleTimeout)
			if presence.Unit.Idle != idle {
				if err := scheduler.actions.setIdleAt(ctx, active, presence.Occupant.PlayerID, idle, now); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// idleState resolves timeout entry and post-idle activity exit.
func idleState(unit roomlive.UnitSnapshot, lastActivity time.Time, now time.Time, timeout time.Duration) bool {
	if unit.Idle {
		return !lastActivity.After(unit.IdleSince)
	}
	return !lastActivity.Add(timeout).After(now)
}

// run executes aggregate idle sweeps until cancellation.
func (scheduler *Scheduler) run(ctx context.Context) {
	ticker := time.NewTicker(scheduler.config.SweepInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			if err := scheduler.Sweep(ctx, now); err != nil && scheduler.log != nil {
				scheduler.log.Error("room idle sweep failed", zap.Error(err))
			}
		}
	}
}
