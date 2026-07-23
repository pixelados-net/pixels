package admin

import (
	"context"
	"errors"
	"testing"

	petpolicy "github.com/niflaot/pixels/internal/realm/pet/policy"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	petruntime "github.com/niflaot/pixels/internal/realm/pet/runtime"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
)

// adminStore persists one pet and transactional audit state.
type adminStore struct {
	petrecord.Store
	// pet stores the current aggregate.
	pet petrecord.Pet
	// auditErr injects audit persistence failure.
	auditErr error
	// auditAction stores the last committed action.
	auditAction string
}

// WithinTransaction rolls pet state back on failure.
func (store *adminStore) WithinTransaction(ctx context.Context, operation func(context.Context) error) error {
	before, action := store.pet, store.auditAction
	if err := operation(ctx); err != nil {
		store.pet, store.auditAction = before, action
		return err
	}
	return nil
}

// Find returns the only pet.
func (store *adminStore) Find(_ context.Context, petID int64) (petrecord.Pet, bool, error) {
	return store.pet, store.pet.ID == petID && store.pet.DeletedAt == nil, nil
}

// ListAdmin returns the only pet.
func (store *adminStore) ListAdmin(context.Context, petrecord.AdminFilter) ([]petrecord.Pet, error) {
	return []petrecord.Pet{store.pet}, nil
}

// UpdateStats applies one optimistic mutation.
func (store *adminStore) UpdateStats(_ context.Context, petID int64, energy int32, happiness int32, experience int32, version int64) (petrecord.Pet, bool, error) {
	if store.pet.ID != petID || store.pet.Version != version {
		return store.pet, false, nil
	}
	store.pet.Energy += energy
	store.pet.Happiness += happiness
	store.pet.Experience += experience
	store.pet.Version++
	return store.pet, true, nil
}

// TransferOwner applies one optimistic inventory transfer.
func (store *adminStore) TransferOwner(_ context.Context, petID int64, ownerID int64, version int64) (petrecord.Pet, bool, error) {
	if store.pet.ID != petID || store.pet.Version != version || !store.pet.Inventory() {
		return store.pet, false, nil
	}
	store.pet.OwnerPlayerID, store.pet.Version = ownerID, version+1
	return store.pet, true, nil
}

// DeleteAdmin applies one optimistic soft deletion marker.
func (store *adminStore) DeleteAdmin(_ context.Context, petID int64, version int64) (bool, error) {
	if store.pet.ID != petID || store.pet.Version != version {
		return false, nil
	}
	store.pet.Version++
	return true, nil
}

// AppendAudit records or rejects one audit row.
func (store *adminStore) AppendAudit(_ context.Context, _ int64, _ int64, action string, _ string) error {
	if store.auditErr != nil {
		return store.auditErr
	}
	store.auditAction = action
	return nil
}

// TestAuditRequiresActorAndReason verifies protected attribution.
func TestAuditRequiresActorAndReason(t *testing.T) {
	for _, audit := range []Audit{{}, {ActorPlayerID: 1}, {Reason: "reason"}} {
		if !errors.Is(audit.Validate(), petrecord.ErrInvalidState) {
			t.Fatalf("expected invalid audit %+v", audit)
		}
	}
	if err := (Audit{ActorPlayerID: 1, Reason: "  QA  "}).Validate(); err != nil {
		t.Fatal(err)
	}
}

// TestUpdateStatsCommitsAuditAtomically verifies mutation and attribution.
func TestUpdateStatsCommitsAuditAtomically(t *testing.T) {
	service, store := adminFixture()
	saved, err := service.UpdateStats(context.Background(), 50, 5, 4, 3, 1, Audit{ActorPlayerID: 1, Reason: "QA"})
	if err != nil {
		t.Fatal(err)
	}
	if saved.Energy != 15 || saved.Happiness != 14 || saved.Experience != 3 || saved.Version != 2 || store.auditAction != "stats_updated" {
		t.Fatalf("saved=%+v action=%q", saved, store.auditAction)
	}
}

// TestUpdateStatsRollsBackWhenAuditFails verifies no unaudited mutation survives.
func TestUpdateStatsRollsBackWhenAuditFails(t *testing.T) {
	service, store := adminFixture()
	sentinel := errors.New("audit unavailable")
	store.auditErr = sentinel
	_, err := service.UpdateStats(context.Background(), 50, 5, 4, 3, 1, Audit{ActorPlayerID: 1, Reason: "QA"})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected audit failure, got %v", err)
	}
	if store.pet.Version != 1 || store.pet.Energy != 10 {
		t.Fatalf("expected rollback, got %+v", store.pet)
	}
}

// TestTransferAndDeleteUseOptimisticAudit verifies inventory lifecycle administration.
func TestTransferAndDeleteUseOptimisticAudit(t *testing.T) {
	service, store := adminFixture()
	saved, err := service.TransferOwner(context.Background(), 50, 9, 1, Audit{ActorPlayerID: 1, Reason: "transfer"})
	if err != nil || saved.OwnerPlayerID != 9 || store.auditAction != "owner_transferred" {
		t.Fatalf("saved=%+v action=%q err=%v", saved, store.auditAction, err)
	}
	if err = service.Delete(context.Background(), 50, 2, Audit{ActorPlayerID: 1, Reason: "cleanup"}); err != nil {
		t.Fatal(err)
	}
	if store.pet.Version != 3 || store.auditAction != "deleted" {
		t.Fatalf("pet=%+v action=%q", store.pet, store.auditAction)
	}
}

// adminFixture creates one inventory pet and projection-safe service.
func adminFixture() (*Service, *adminStore) {
	store := &adminStore{pet: petrecord.Pet{ID: 50, OwnerPlayerID: 7, Name: "Pixel", State: petrecord.StateInventory, Energy: 10, Happiness: 10, Version: 1}}
	rooms := roomlive.NewRegistry(nil)
	runtimeService := petruntime.New(petpolicy.Config{Enabled: true}, store, nil, rooms, nil, nil, nil, nil, nil, nil, nil)
	return New(store, nil, nil, nil, runtimeService, rooms), store
}
