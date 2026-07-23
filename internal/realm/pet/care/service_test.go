package care

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/permission"
	petpolicy "github.com/niflaot/pixels/internal/realm/pet/policy"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	petreference "github.com/niflaot/pixels/internal/realm/pet/reference"
	petruntime "github.com/niflaot/pixels/internal/realm/pet/runtime"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outtraining "github.com/niflaot/pixels/networking/outbound/room/pet/training"
)

// careReferences supplies one immutable training fixture.
type careReferences struct {
	// snapshot stores the fixture generation.
	snapshot *petreference.Snapshot
}

// Current returns the fixture generation.
func (references careReferences) Current(context.Context) (*petreference.Snapshot, error) {
	return references.snapshot, nil
}

// Refresh leaves the immutable fixture unchanged.
func (references careReferences) Refresh(context.Context) error { return nil }

// careChecker supplies one permission decision.
type careChecker struct {
	// allowed stores the returned decision.
	allowed bool
}

// HasPermission returns the configured decision.
func (checker careChecker) HasPermission(context.Context, int64, permission.Node) (bool, error) {
	return checker.allowed, nil
}

// careClock returns one fixed wall time.
type careClock struct {
	// now stores the configured wall time.
	now time.Time
}

// Now returns the configured wall time.
func (clock careClock) Now() time.Time { return clock.now }

// careStore records one respect mutation.
type careStore struct {
	petrecord.Store
	// pet stores the visible aggregate.
	pet petrecord.Pet
	// applied controls the quota outcome.
	applied bool
	// experience stores the observed experience delta.
	experience int32
	// dailyLimit stores the observed quota.
	dailyLimit int
}

// Room returns the visible pet for room activation.
func (store *careStore) Room(_ context.Context, roomID int64) ([]petrecord.Pet, error) {
	if store.pet.RoomID == nil || *store.pet.RoomID != roomID {
		return nil, nil
	}
	return []petrecord.Pet{store.pet}, nil
}

// Respect applies the configured quota outcome.
func (store *careStore) Respect(_ context.Context, petID int64, _ int64, experience int32, dailyLimit int) (petrecord.Pet, bool, error) {
	store.experience, store.dailyLimit = experience, dailyLimit
	if petID != store.pet.ID || !store.applied {
		return store.pet, false, nil
	}
	store.pet.Respect++
	store.pet.Experience += experience
	store.pet.Version++
	return store.pet, true, nil
}

// TestRespectUsesInjectedAge verifies minimum age without real-time sleeps.
func TestRespectUsesInjectedAge(t *testing.T) {
	now := time.Date(2026, time.July, 17, 12, 0, 0, 0, time.UTC)
	service, _, rooms, active, _ := careFixture(t, now, now.Add(-24*time.Hour), false)
	result, err := service.Respect(context.Background(), 9, 50, 8)
	if err != nil {
		t.Fatal(err)
	}
	if !result.TooYoung || result.AgeDays != 1 || result.RequiredAgeDays != 3 {
		t.Fatalf("result=%+v", result)
	}
	_, _, _ = rooms.Close(context.Background(), active.ID())
}

// TestRespectRejectsOwnerAndQuota verifies policy failures stay explicit.
func TestRespectRejectsOwnerAndQuota(t *testing.T) {
	now := time.Date(2026, time.July, 17, 12, 0, 0, 0, time.UTC)
	service, _, rooms, active, store := careFixture(t, now, now.Add(-10*24*time.Hour), false)
	if _, err := service.Respect(context.Background(), 9, 50, 7); !errors.Is(err, petrecord.ErrNoRights) {
		t.Fatalf("expected owner rejection, got %v", err)
	}
	if _, err := service.Respect(context.Background(), 9, 50, 8); !errors.Is(err, petrecord.ErrRespectQuota) {
		t.Fatalf("expected quota rejection, got %v", err)
	}
	if store.dailyLimit != 3 {
		t.Fatalf("daily limit=%d", store.dailyLimit)
	}
	_, _, _ = rooms.Close(context.Background(), active.ID())
}

// TestRespectProjectsCommittedState verifies the runtime snapshot follows persistence.
func TestRespectProjectsCommittedState(t *testing.T) {
	now := time.Date(2026, time.July, 17, 12, 0, 0, 0, time.UTC)
	service, runtimeService, rooms, active, store := careFixture(t, now, now.Add(-10*24*time.Hour), true)
	result, err := service.Respect(context.Background(), 9, 50, 8)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Applied || result.Pet.Respect != 1 || store.experience != 10 {
		t.Fatalf("result=%+v experience=%d", result, store.experience)
	}
	current, found := runtimeService.Snapshot(9, 50)
	if !found || current.Version != 2 || current.Respect != 1 {
		t.Fatalf("runtime=%+v found=%v", current, found)
	}
	_, _, _ = rooms.Close(context.Background(), active.ID())
}

// TestRespectBypassRemovesDailyLimit verifies the explicit permission override.
func TestRespectBypassRemovesDailyLimit(t *testing.T) {
	now := time.Date(2026, time.July, 17, 12, 0, 0, 0, time.UTC)
	service, _, rooms, active, store := careFixture(t, now, now.Add(-10*24*time.Hour), true)
	service.permissions = careChecker{allowed: true}
	if _, err := service.Respect(context.Background(), 9, 50, 8); err != nil {
		t.Fatal(err)
	}
	if store.dailyLimit != 0 {
		t.Fatalf("expected bypassed quota, got %d", store.dailyLimit)
	}
	_, _, _ = rooms.Close(context.Background(), active.ID())
}

// TestTrainingSendsLearnedCommands verifies the visible-pet protocol projection.
func TestTrainingSendsLearnedCommands(t *testing.T) {
	now := time.Date(2026, time.July, 17, 12, 0, 0, 0, time.UTC)
	service, _, rooms, active, _ := careFixture(t, now, now.Add(-10*24*time.Hour), true)
	target, packets := careConnection(t)
	if err := service.Training(context.Background(), target, 9, 50); err != nil {
		t.Fatal(err)
	}
	if len(*packets) != 1 || (*packets)[0].Header != outtraining.Header {
		t.Fatalf("packets=%+v", *packets)
	}
	_, _, _ = rooms.Close(context.Background(), active.ID())
}

// careFixture creates one loaded room and deterministic visible pet.
func careFixture(t testing.TB, now time.Time, createdAt time.Time, applied bool) (*Service, *petruntime.Service, *roomlive.Registry, *roomlive.Room, *careStore) {
	t.Helper()
	roomID, x, y, z, rotation := int64(9), 1, 0, 0.0, int16(2)
	store := &careStore{pet: petrecord.Pet{ID: 50, OwnerPlayerID: 7, Name: "Pixel", RoomID: &roomID, X: &x, Y: &y, Z: &z, Rotation: &rotation, State: petrecord.StateRoom, CreatedAt: createdAt, StatsAt: now, Energy: 100, Happiness: 100, Version: 1}, applied: applied}
	rooms := roomlive.NewRegistry(nil)
	active, err := rooms.Activate(roomlive.Snapshot{ID: roomID, OwnerPlayerID: 7, MaxUsers: 25, AllowPets: true})
	if err != nil {
		t.Fatal(err)
	}
	roomGrid, err := grid.Parse("00", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if err = active.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}}); err != nil {
		t.Fatal(err)
	}
	config := petpolicy.Config{Enabled: true, RespectMinimumAge: 3 * 24 * time.Hour, RespectDailyLimit: 3, RespectExperience: 10}
	references := &petreference.Snapshot{}
	references.SpeciesCommands[0] = []int32{1, 2}
	references.Commands[1], references.Commands[2] = petrecord.Command{ID: 1, RequiredLevel: 1, Enabled: true}, petrecord.Command{ID: 2, RequiredLevel: 2, Enabled: true}
	references.CommandPresent[1], references.CommandPresent[2] = true, true
	runtimeService := petruntime.New(config, store, careReferences{snapshot: references}, rooms, nil, nil, nil, nil, nil, nil, nil)
	runtimeService.SetClock(careClock{now: now})
	if err = runtimeService.EnsureRoom(context.Background(), active); err != nil {
		t.Fatal(err)
	}
	return New(config, store, nil, rooms, runtimeService), runtimeService, rooms, active, store
}

// careConnection creates a transport-backed handler context and packet sink.
func careConnection(t testing.TB) (netconn.Context, *[]codec.Packet) {
	t.Helper()
	outbound := netconn.NewHandlerRegistry()
	var target netconn.Context
	packets := make([]codec.Packet, 0, 1)
	outbound.SetFallback(func(current netconn.Context, _ codec.Packet) error {
		target = current
		return nil
	}, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	session, err := netconn.NewSession(netconn.SessionConfig{ID: "care-test", Kind: "test", Outbound: outbound, Sender: func(_ context.Context, packet codec.Packet) error {
		packets = append(packets, packet)
		return nil
	}, Disposer: func(context.Context, netconn.Reason) error { return nil }})
	if err != nil {
		t.Fatal(err)
	}
	if err = session.Send(context.Background(), codec.Packet{Header: 1}); err != nil {
		t.Fatal(err)
	}
	packets = packets[:0]
	return target, &packets
}
