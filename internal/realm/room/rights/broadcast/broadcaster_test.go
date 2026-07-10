package broadcast

import (
	"context"
	"errors"
	"testing"

	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	rightsgranted "github.com/niflaot/pixels/internal/realm/room/events/rightsgranted"
	rightsrevoked "github.com/niflaot/pixels/internal/realm/room/events/rightsrevoked"
	roomlive "github.com/niflaot/pixels/internal/realm/room/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outstatus "github.com/niflaot/pixels/networking/outbound/room/entities/status"
	"github.com/niflaot/pixels/pkg/bus"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
)

// playerFinderForTest resolves one player username.
type playerFinderForTest struct{}

// FindByID finds one player.
func (playerFinderForTest) FindByID(context.Context, int64) (playerservice.Record, bool, error) {
	return playerservice.Record{Player: playermodel.Player{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 2}}, Username: "Alice"}}, true, nil
}

// FindByUsername finds no player.
func (playerFinderForTest) FindByUsername(context.Context, string) (playerservice.Record, bool, error) {
	return playerservice.Record{}, false, nil
}

// failingPlayerFinderForTest fails player lookup.
type failingPlayerFinderForTest struct{}

// FindByID fails player lookup.
func (failingPlayerFinderForTest) FindByID(context.Context, int64) (playerservice.Record, bool, error) {
	return playerservice.Record{}, false, errors.New("player lookup failed")
}

// FindByUsername fails player lookup.
func (failingPlayerFinderForTest) FindByUsername(context.Context, string) (playerservice.Record, bool, error) {
	return playerservice.Record{}, false, errors.New("player lookup failed")
}

// missingPlayerFinderForTest reports a deleted player projection.
type missingPlayerFinderForTest struct{}

// FindByID reports no player.
func (missingPlayerFinderForTest) FindByID(context.Context, int64) (playerservice.Record, bool, error) {
	return playerservice.Record{}, false, nil
}

// FindByUsername reports no player.
func (missingPlayerFinderForTest) FindByUsername(context.Context, string) (playerservice.Record, bool, error) {
	return playerservice.Record{}, false, nil
}

// TestBroadcasterProjectsRightsIntoActiveRoom verifies cache synchronization.
func TestBroadcasterProjectsRightsIntoActiveRoom(t *testing.T) {
	runtime := roomlive.NewRegistry(nil)
	active, err := runtime.Activate(roomlive.Snapshot{ID: 9, OwnerPlayerID: 1, MaxUsers: 25})
	if err != nil {
		t.Fatalf("activate room: %v", err)
	}
	loadWorldForRightsBroadcastTest(t, active)
	connections, sent := connectionForTest(t)
	if _, err := runtime.Join(context.Background(), 9, roomlive.Occupant{PlayerID: 2, Username: "Alice", ConnectionID: "conn", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("join room: %v", err)
	}
	broadcaster := New(runtime, connections, playerFinderForTest{})
	if err := broadcaster.Granted(context.Background(), 9, 2); err != nil || !active.HasRights(2) {
		t.Fatalf("grant projection rights=%v err=%v", active.HasRights(2), err)
	}
	if err := broadcaster.Revoked(context.Background(), 9, 2); err != nil || active.HasRights(2) {
		t.Fatalf("revoke projection rights=%v err=%v", active.HasRights(2), err)
	}
	if len(*sent) != 6 || (*sent)[1].Header != outstatus.Header || (*sent)[4].Header != outstatus.Header {
		t.Fatalf("expected list, status, and level packets for both mutations, got %#v", *sent)
	}
}

// loadWorldForRightsBroadcastTest loads the minimum world needed for avatar status projection.
func loadWorldForRightsBroadcastTest(t *testing.T, active *roomlive.Room) {
	t.Helper()
	roomGrid, err := grid.Parse("0", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatalf("parse room grid: %v", err)
	}
	err = active.LoadWorld(roomlive.WorldConfig{
		Grid: roomGrid, Door: worldpath.Position{Point: grid.MustPoint(0, 0)},
		Body: worldunit.RotationSouth, Head: worldunit.RotationSouth, Rules: worldpath.DefaultRules(),
	})
	if err != nil {
		t.Fatalf("load room world: %v", err)
	}
}

// TestRegisterProjectsBusEvents verifies subscriber lifecycle registration.
func TestRegisterProjectsBusEvents(t *testing.T) {
	runtime := roomlive.NewRegistry(nil)
	active, err := runtime.Activate(roomlive.Snapshot{ID: 9, OwnerPlayerID: 1, MaxUsers: 25})
	if err != nil {
		t.Fatalf("activate room: %v", err)
	}
	local := bus.New()
	lifecycle := fxtest.NewLifecycle(t)
	if err := Register(lifecycle, local, New(runtime, nil, playerFinderForTest{}), zap.NewNop()); err != nil {
		t.Fatalf("register: %v", err)
	}
	if err := local.Publish(context.Background(), bus.Event{Name: rightsgranted.Name, Payload: rightsgranted.Payload{RoomID: 9, PlayerID: 2, ActorID: 1}}); err != nil {
		t.Fatalf("publish grant: %v", err)
	}
	if !active.HasRights(2) {
		t.Fatal("expected projected rights")
	}
	if err := local.Publish(context.Background(), bus.Event{Name: rightsrevoked.Name, Payload: rightsrevoked.Payload{RoomID: 9, PlayerID: 2, ActorID: 1}}); err != nil {
		t.Fatalf("publish revoke: %v", err)
	}
	if active.HasRights(2) {
		t.Fatal("expected revoked rights projection")
	}
	lifecycle.RequireStop()
}

// TestDeferredGrantRunsWithoutTransaction verifies standalone event projection.
func TestDeferredGrantRunsWithoutTransaction(t *testing.T) {
	runtime := roomlive.NewRegistry(nil)
	active, err := runtime.Activate(roomlive.Snapshot{ID: 9, OwnerPlayerID: 1, MaxUsers: 25})
	if err != nil {
		t.Fatalf("activate room: %v", err)
	}
	handler := deferGranted(New(runtime, nil, playerFinderForTest{}), nilLoggerForTest())
	err = handler(context.Background(), bus.Event{Name: rightsgranted.Name, Payload: rightsgranted.Payload{RoomID: 9, PlayerID: 2, ActorID: 1}})
	if err != nil || !active.HasRights(2) {
		t.Fatalf("deferred grant rights=%v err=%v", active.HasRights(2), err)
	}
}

// TestBroadcasterIgnoresInactiveRoomsAndUnexpectedPayloads verifies no-op boundaries.
func TestBroadcasterIgnoresInactiveRoomsAndUnexpectedPayloads(t *testing.T) {
	broadcaster := New(roomlive.NewRegistry(nil), nil, playerFinderForTest{})
	if err := broadcaster.Granted(context.Background(), 99, 2); err != nil {
		t.Fatalf("inactive grant: %v", err)
	}
	if err := broadcaster.Revoked(context.Background(), 99, 2); err != nil {
		t.Fatalf("inactive revoke: %v", err)
	}
	if err := deferGranted(broadcaster, zap.NewNop())(context.Background(), bus.Event{Payload: "invalid"}); err != nil {
		t.Fatalf("unexpected grant payload: %v", err)
	}
	if err := deferRevoked(broadcaster, zap.NewNop())(context.Background(), bus.Event{Payload: "invalid"}); err != nil {
		t.Fatalf("unexpected revoke payload: %v", err)
	}
	runtime := roomlive.NewRegistry(nil)
	if _, err := runtime.Activate(roomlive.Snapshot{ID: 9, MaxUsers: 25}); err != nil {
		t.Fatalf("activate room: %v", err)
	}
	if _, err := runtime.Join(context.Background(), 9, roomlive.Occupant{PlayerID: 2, Username: "Alice", ConnectionID: "missing", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("join room: %v", err)
	}
	if err := New(runtime, netconn.NewRegistry(), playerFinderForTest{}).Granted(context.Background(), 9, 2); err != nil {
		t.Fatalf("missing connection projection: %v", err)
	}
	if _, err := runtime.Activate(roomlive.Snapshot{ID: 10, MaxUsers: 25}); err != nil {
		t.Fatalf("activate empty room: %v", err)
	}
	if err := New(runtime, netconn.NewRegistry(), playerFinderForTest{}).Granted(context.Background(), 10, 3); err != nil {
		t.Fatalf("empty room projection: %v", err)
	}
}

// TestDeferredGrantContainsProjectionFailures verifies committed mutations do not fail retroactively.
func TestDeferredGrantContainsProjectionFailures(t *testing.T) {
	runtime := roomlive.NewRegistry(nil)
	if _, err := runtime.Activate(roomlive.Snapshot{ID: 9, OwnerPlayerID: 1, MaxUsers: 25}); err != nil {
		t.Fatalf("activate room: %v", err)
	}
	broadcaster := New(runtime, nil, failingPlayerFinderForTest{})
	handler := deferGranted(broadcaster, zap.NewNop())
	err := handler(context.Background(), bus.Event{Name: rightsgranted.Name, Payload: rightsgranted.Payload{RoomID: 9, PlayerID: 2}})
	if err != nil {
		t.Fatalf("expected contained projection error, got %v", err)
	}
	if err := New(runtime, nil, missingPlayerFinderForTest{}).Granted(context.Background(), 9, 3); err != nil {
		t.Fatalf("missing player projection: %v", err)
	}
}

// nilLoggerForTest creates a silent logger.
func nilLoggerForTest() *zap.Logger {
	return zap.NewNop()
}

// connectionForTest creates one registered packet-capturing connection.
func connectionForTest(t *testing.T) (*netconn.Registry, *[]codec.Packet) {
	t.Helper()
	inbound := netconn.NewHandlerRegistry()
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	sent := make([]codec.Packet, 0, 4)
	session, err := netconn.NewSession(netconn.SessionConfig{
		ID: "conn", Kind: "websocket", Inbound: inbound, Outbound: outbound,
		Sender:   func(_ context.Context, packet codec.Packet) error { sent = append(sent, packet); return nil },
		Disposer: func(context.Context, netconn.Reason) error { return nil },
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	registry := netconn.NewRegistry()
	if err := registry.Register(session); err != nil {
		t.Fatalf("register session: %v", err)
	}

	return registry, &sent
}
