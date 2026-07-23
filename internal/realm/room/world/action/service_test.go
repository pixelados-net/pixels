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
	outdance "github.com/niflaot/pixels/networking/outbound/room/entities/dance"
	outexpression "github.com/niflaot/pixels/networking/outbound/room/entities/expression"
	"github.com/niflaot/pixels/pkg/bus"
)

// TestServiceCoordinatesAvatarState verifies dance, expression, idle, and posture invariants.
func TestServiceCoordinatesAvatarState(t *testing.T) {
	active := actionRoom(t)
	service := New(Config{TransitionDelay: time.Nanosecond}, nil, bus.New())
	active.SetUnitStatus(7, worldunit.StatusSit, "0.5")
	if err := service.Dance(context.Background(), active, 7, 3); err != nil {
		t.Fatal(err)
	}
	unit, _ := active.Unit(7)
	if hasStatus(unit, worldunit.StatusSit) || !hasStatus(unit, worldunit.StatusDance) {
		t.Fatalf("unexpected dancing unit %#v", unit)
	}
	if err := service.Express(context.Background(), active, 7, 2); err != nil {
		t.Fatal(err)
	}
	unit, _ = active.Unit(7)
	if hasStatus(unit, worldunit.StatusDance) {
		t.Fatalf("expected expression to cancel dance %#v", unit)
	}
	if err := service.SetIdle(context.Background(), active, 7, true); err != nil {
		t.Fatal(err)
	}
	unit, _ = active.Unit(7)
	if !unit.Idle || !unit.ManualIdle || hasStatus(unit, worldunit.StatusDance) {
		t.Fatalf("unexpected idle unit %#v", unit)
	}
	if err := service.Dance(context.Background(), active, 7, 4); err != nil {
		t.Fatal(err)
	}
	unit, _ = active.Unit(7)
	if unit.Idle || !hasStatus(unit, worldunit.StatusDance) {
		t.Fatalf("unexpected resumed unit %#v", unit)
	}
	if err := service.Posture(context.Background(), active, 7, true); err != nil {
		t.Fatal(err)
	}
	unit, _ = active.Unit(7)
	if hasStatus(unit, worldunit.StatusDance) || !hasStatus(unit, worldunit.StatusSit) {
		t.Fatalf("unexpected sitting unit %#v", unit)
	}
	if err := service.Sign(context.Background(), active, 7, 18); err != nil {
		t.Fatal(err)
	}
	unit, _ = active.Unit(7)
	if hasStatus(unit, worldunit.StatusSign) || unit.Idle {
		t.Fatalf("expected transient resumed sign %#v", unit)
	}
}

// TestServiceProjectsActionCancellation verifies replacements explicitly clear the previous client action.
func TestServiceProjectsActionCancellation(t *testing.T) {
	active := actionRoom(t)
	connections, packets := actionConnection(t)
	service := New(Config{TransitionDelay: time.Nanosecond}, connections, nil)
	if err := service.Dance(context.Background(), active, 7, 3); err != nil {
		t.Fatal(err)
	}
	*packets = (*packets)[:0]
	if err := service.Express(context.Background(), active, 7, 1); err != nil {
		t.Fatal(err)
	}
	if !containsActionValue(t, *packets, outdance.Header, outdance.Definition, 0) {
		t.Fatalf("expected dance cancellation packets %#v", *packets)
	}
	*packets = (*packets)[:0]
	if err := service.Dance(context.Background(), active, 7, 4); err != nil {
		t.Fatal(err)
	}
	if !containsActionValue(t, *packets, outexpression.Header, outexpression.Definition, 0) {
		t.Fatalf("expected expression cancellation packets %#v", *packets)
	}
}

// TestServiceMissingUnitReturnsStableErrors verifies stale room commands are harmless.
func TestServiceMissingUnitReturnsStableErrors(t *testing.T) {
	active := actionRoom(t)
	service := New(Config{TransitionDelay: time.Nanosecond}, nil, nil)
	if err := service.Dance(context.Background(), active, 99, 1); err != roomlive.ErrUnitNotFound {
		t.Fatalf("expected missing dance unit, got %v", err)
	}
	if err := service.Express(context.Background(), active, 99, 1); err != roomlive.ErrUnitNotFound {
		t.Fatalf("expected missing expression unit, got %v", err)
	}
	if err := service.SetIdle(context.Background(), active, 99, true); err != roomlive.ErrUnitNotFound {
		t.Fatalf("expected missing idle unit, got %v", err)
	}
}

// TestServiceAppliesTransitionDelay verifies replacement actions are not projected in the same instant.
func TestServiceAppliesTransitionDelay(t *testing.T) {
	active := actionRoom(t)
	delay := 5 * time.Millisecond
	service := New(Config{TransitionDelay: delay}, nil, nil)
	started := time.Now()
	if err := service.Dance(context.Background(), active, 7, 3); err != nil {
		t.Fatal(err)
	}
	if elapsed := time.Since(started); elapsed < delay {
		t.Fatalf("expected at least %s transition delay, got %s", delay, elapsed)
	}
}

// actionRoom creates one loaded room with a player unit.
func actionRoom(t testing.TB) *roomlive.Room {
	t.Helper()
	roomGrid, err := grid.Parse("00", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	active, err := roomlive.NewRoom(roomlive.Snapshot{ID: 9, MaxUsers: 5})
	if err != nil {
		t.Fatal(err)
	}
	if err = active.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}, Body: worldunit.RotationSouth, Head: worldunit.RotationSouth}); err != nil {
		t.Fatal(err)
	}
	if _, err = active.Join(roomlive.Occupant{PlayerID: 7, Username: "demo", ConnectionID: "action", ConnectionKind: "websocket"}); err != nil {
		t.Fatal(err)
	}
	return active
}

// actionConnection creates one packet-capturing room connection.
func actionConnection(t testing.TB) (*netconn.Registry, *[]codec.Packet) {
	t.Helper()
	connections := netconn.NewRegistry()
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	packets := make([]codec.Packet, 0, 8)
	session, err := netconn.NewSession(netconn.SessionConfig{ID: "action", Kind: "websocket", Outbound: outbound,
		Sender:   func(_ context.Context, packet codec.Packet) error { packets = append(packets, packet); return nil },
		Disposer: func(context.Context, netconn.Reason) error { return nil }})
	if err != nil {
		t.Fatal(err)
	}
	if err = connections.Register(session); err != nil {
		t.Fatal(err)
	}
	return connections, &packets
}

// containsActionValue reports whether one two-int action packet carries the requested action id.
func containsActionValue(t testing.TB, packets []codec.Packet, header uint16, definition codec.Definition, value int32) bool {
	t.Helper()
	for _, packet := range packets {
		if packet.Header != header {
			continue
		}
		values, err := codec.DecodePacketExact(packet, definition)
		if err != nil {
			t.Fatal(err)
		}
		if len(values) == 2 && values[1].Int32 == value {
			return true
		}
	}
	return false
}
