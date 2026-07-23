package breeding

import (
	"context"
	"testing"

	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// WithinTransaction runs one in-memory breeding transaction.
func (store *breedingStore) WithinTransaction(ctx context.Context, operation func(context.Context) error) error {
	return operation(ctx)
}

// SaveBreedingSession creates or adds one owner confirmation.
func (store *breedingStore) SaveBreedingSession(_ context.Context, value petrecord.BreedingSession, actorID int64) (petrecord.BreedingSession, bool, error) {
	if store.session.NestItemID == 0 {
		value.State, value.Version = "requested", 1
		store.session = value
	} else {
		store.session.Version++
	}
	if store.pets[0].OwnerPlayerID == actorID {
		store.session.OwnerOneConfirmed = true
	}
	if store.pets[1].OwnerPlayerID == actorID {
		store.session.OwnerTwoConfirmed = true
	}
	if store.session.OwnerOneConfirmed && store.session.OwnerTwoConfirmed {
		store.session.State = "confirmed"
	}
	return store.session, true, nil
}

// FindBreedingSession returns the fixture workflow.
func (store *breedingStore) FindBreedingSession(_ context.Context, nestID int64) (petrecord.BreedingSession, bool, error) {
	return store.session, store.session.NestItemID == nestID, nil
}

// SetBreedingSessionState compare-and-swaps the fixture workflow.
func (store *breedingStore) SetBreedingSessionState(_ context.Context, nestID int64, from string, to string, version int64) (bool, error) {
	if store.session.NestItemID != nestID || store.session.State != from || store.session.Version != version {
		return false, nil
	}
	store.session.State, store.session.Version = to, version+1
	return true, nil
}

// SetBreedingEligibility consumes one parent's eligibility.
func (store *breedingStore) SetBreedingEligibility(_ context.Context, petID int64, eligible bool, version int64) (petrecord.Pet, bool, error) {
	for index := range store.pets {
		if store.pets[index].ID == petID && store.pets[index].Version == version {
			store.pets[index].CanBreed, store.pets[index].Version = eligible, version+1
			return store.pets[index], true, nil
		}
	}
	return petrecord.Pet{}, false, nil
}

// Grant creates one deterministic offspring.
func (store *breedingStore) Grant(_ context.Context, params petrecord.GrantParams) (petrecord.Pet, bool, error) {
	if store.offspring.ID == 0 {
		store.offspring = petrecord.Pet{ID: 99, OwnerPlayerID: params.OwnerPlayerID, Name: params.Name, TypeID: params.TypeID, BreedID: params.BreedID, PaletteID: params.PaletteID, Color: params.Color, State: petrecord.StateInventory, Version: 1}
		return store.offspring, true, nil
	}
	return store.offspring, false, nil
}

// CancelBreedingRoom cancels the active fixture workflow.
func (store *breedingStore) CancelBreedingRoom(_ context.Context, roomID int64) error {
	if store.session.RoomID == roomID && (store.session.State == "requested" || store.session.State == "confirmed") {
		store.session.State = "cancelled"
	}
	return nil
}

// TestBreedingWorkflowConfirmsBothOwnersAndGrantsOnce verifies the complete nest transaction.
func TestBreedingWorkflowConfirmsBothOwnersAndGrantsOnce(t *testing.T) {
	service, runtimeService, rooms, active, _, _ := breedingFixture(t)
	store := service.store.(*breedingStore)
	target := breedingConnection(t)
	if err := service.Start(context.Background(), target, 17, 7, 1, 2); err != nil || store.session.State != "requested" {
		t.Fatalf("first confirmation state=%+v err=%v", store.session, err)
	}
	if err := service.Start(context.Background(), target, 17, 8, 1, 2); err != nil || store.session.State != "confirmed" {
		t.Fatalf("second confirmation state=%+v err=%v", store.session, err)
	}
	if err := service.Confirm(context.Background(), target, 17, 8, 700, "Junior", 1, 2); err != nil {
		t.Fatal(err)
	}
	first, firstFound := runtimeService.Snapshot(17, 1)
	second, secondFound := runtimeService.Snapshot(17, 2)
	if store.session.State != "completed" || store.offspring.ID != 99 || store.offspring.OwnerPlayerID != 8 || !firstFound || !secondFound || first.CanBreed || second.CanBreed {
		t.Fatalf("session=%+v offspring=%+v first=%+v second=%+v", store.session, store.offspring, first, second)
	}
	_, _, _ = rooms.Close(context.Background(), active.ID())
}

// breedingConnection creates one transport-backed handler context.
func breedingConnection(t testing.TB) netconn.Context {
	t.Helper()
	outbound := netconn.NewHandlerRegistry()
	var target netconn.Context
	outbound.SetFallback(func(current netconn.Context, _ codec.Packet) error {
		target = current
		return nil
	}, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	session, err := netconn.NewSession(netconn.SessionConfig{ID: "breeding-test", Kind: "test", Outbound: outbound, Sender: func(context.Context, codec.Packet) error { return nil }, Disposer: func(context.Context, netconn.Reason) error { return nil }})
	if err != nil {
		t.Fatal(err)
	}
	if err = session.Send(context.Background(), codec.Packet{Header: 1}); err != nil {
		t.Fatal(err)
	}
	return target
}
