package action

import (
	"context"
	"testing"
	"time"

	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
)

// TestIdleState verifies timeout entry and activity-driven exit without sleeps.
func TestIdleState(t *testing.T) {
	timeout := 5 * time.Minute
	last := time.Unix(100, 0)
	idleSince := time.Unix(500, 0)
	tests := []struct {
		name string
		unit roomlive.UnitSnapshot
		last time.Time
		now  time.Time
		want bool
	}{
		{name: "active", last: last, now: last.Add(timeout - time.Second)},
		{name: "timed out", last: last, now: last.Add(timeout), want: true},
		{name: "automatic idle remains", unit: roomlive.UnitSnapshot{Idle: true, IdleSince: idleSince}, last: idleSince.Add(-time.Nanosecond), now: idleSince.Add(time.Second), want: true},
		{name: "new input exits", unit: roomlive.UnitSnapshot{Idle: true, IdleSince: idleSince}, last: idleSince.Add(time.Nanosecond), now: idleSince.Add(time.Second)},
		{name: "manual ignores unrelated input", unit: roomlive.UnitSnapshot{Idle: true, IdleSince: idleSince, ManualIdle: true}, last: idleSince.Add(time.Second), now: idleSince.Add(2 * time.Second), want: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := idleState(test.unit, test.last, test.now, timeout); got != test.want {
				t.Fatalf("expected %t, got %t", test.want, got)
			}
		})
	}
}

// TestSchedulerSweepMarksAndClearsIdle verifies one aggregate reconciliation pass.
func TestSchedulerSweepMarksAndClearsIdle(t *testing.T) {
	connections := netconn.NewRegistry()
	inbound := netconn.NewHandlerRegistry()
	inbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	session, err := netconn.NewSession(netconn.SessionConfig{ID: "action", Kind: "websocket", StartedAt: time.Unix(100, 0), Inbound: inbound, Outbound: outbound,
		Sender: func(context.Context, codec.Packet) error { return nil }, Disposer: func(context.Context, netconn.Reason) error { return nil }})
	if err != nil {
		t.Fatal(err)
	}
	if err = connections.Register(session); err != nil {
		t.Fatal(err)
	}
	runtime := roomlive.NewRegistry(nil)
	active, err := runtime.Activate(roomlive.Snapshot{ID: 9, MaxUsers: 5})
	if err != nil {
		t.Fatal(err)
	}
	roomGrid, err := grid.Parse("00", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if err = active.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}, Body: worldunit.RotationSouth, Head: worldunit.RotationSouth}); err != nil {
		t.Fatal(err)
	}
	if _, err = runtime.Join(context.Background(), 9, roomlive.Occupant{PlayerID: 7, Username: "demo", ConnectionID: "action", ConnectionKind: "websocket"}); err != nil {
		t.Fatal(err)
	}
	scheduler := NewScheduler(Config{IdleTimeout: 5 * time.Minute}, runtime, connections, New(Config{TransitionDelay: time.Nanosecond}, nil, nil), zap.NewNop())
	if err = scheduler.Sweep(context.Background(), time.Unix(400, 0)); err != nil {
		t.Fatal(err)
	}
	unit, _ := active.Unit(7)
	if !unit.Idle {
		t.Fatal("expected timed-out unit to become idle")
	}
	if err = session.Receive(context.Background(), codec.Packet{Header: 1}); err != nil {
		t.Fatal(err)
	}
	if err = scheduler.Sweep(context.Background(), time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	unit, _ = active.Unit(7)
	if unit.Idle {
		t.Fatal("expected new inbound activity to clear idle")
	}
	if _, closed, closeErr := runtime.Close(context.Background(), 9); closeErr != nil || !closed {
		t.Fatalf("close room: closed=%t err=%v", closed, closeErr)
	}
}

// TestConfigDefaultsAndEnvironment verifies bounded idle configuration.
func TestConfigDefaultsAndEnvironment(t *testing.T) {
	defaults := (Config{}).Normalize()
	if defaults.IdleTimeout != 5*time.Minute || defaults.SweepInterval != time.Second || defaults.TransitionDelay != 100*time.Millisecond {
		t.Fatalf("unexpected defaults %#v", defaults)
	}
	t.Setenv("PIXELS_ROOM_IDLE_TIMEOUT", "7m")
	t.Setenv("PIXELS_ROOM_IDLE_SWEEP_INTERVAL", "2s")
	t.Setenv("PIXELS_ROOM_ACTION_TRANSITION_DELAY", "75ms")
	configured, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if configured.IdleTimeout != 7*time.Minute || configured.SweepInterval != 2*time.Second || configured.TransitionDelay != 75*time.Millisecond {
		t.Fatalf("unexpected environment config %#v", configured)
	}
}

// TestSchedulerLifecycle verifies loop startup and idempotent shutdown.
func TestSchedulerLifecycle(t *testing.T) {
	scheduler := NewScheduler(Config{SweepInterval: time.Millisecond}, roomlive.NewRegistry(nil), netconn.NewRegistry(), New(Config{TransitionDelay: time.Nanosecond}, nil, nil), zap.NewNop())
	lifecycle := fxtest.NewLifecycle(t)
	RegisterScheduler(lifecycle, scheduler)
	lifecycle.RequireStart().RequireStop()
}

// BenchmarkIdleState measures the per-presence scheduler decision.
func BenchmarkIdleState(b *testing.B) {
	unit := roomlive.UnitSnapshot{Idle: true, IdleSince: time.Unix(100, 0)}
	last := time.Unix(99, 0)
	now := time.Unix(101, 0)
	b.ReportAllocs()
	for b.Loop() {
		_ = idleState(unit, last, now, 5*time.Minute)
	}
}
