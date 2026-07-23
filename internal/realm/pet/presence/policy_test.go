package presence

import (
	"context"
	"testing"

	"github.com/niflaot/pixels/internal/permission"
	petpolicy "github.com/niflaot/pixels/internal/realm/pet/policy"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	petruntime "github.com/niflaot/pixels/internal/realm/pet/runtime"
)

// policyStore returns one configured inventory count.
type policyStore struct {
	petrecord.Store
	// count stores the durable inventory count.
	count int
}

// CountInventory returns the configured inventory count.
func (store *policyStore) CountInventory(context.Context, int64) (int, error) {
	return store.count, nil
}

// policyChecker allows one configured permission node.
type policyChecker struct {
	// allowed stores the only allowed node.
	allowed permission.Node
}

// HasPermission reports whether the queried node is configured.
func (checker policyChecker) HasPermission(_ context.Context, _ int64, node permission.Node) (bool, error) {
	return node == checker.allowed, nil
}

// TestCheckInventoryLimitHonorsBypass verifies exact-boundary and permission policy.
func TestCheckInventoryLimitHonorsBypass(t *testing.T) {
	service := &Service{config: petpolicy.Config{MaxInventory: 2}, store: &policyStore{count: 2}}
	if err := service.checkInventoryLimit(context.Background(), 7); err != petrecord.ErrInventoryLimit {
		t.Fatalf("expected inventory limit, got %v", err)
	}
	service.permissions = policyChecker{allowed: petpolicy.InventoryLimitBypass}
	if err := service.checkInventoryLimit(context.Background(), 7); err != nil {
		t.Fatalf("expected inventory bypass, got %v", err)
	}
}

// TestCheckRoomLimitsHonorsBypass verifies loaded room counts and permission policy.
func TestCheckRoomLimitsHonorsBypass(t *testing.T) {
	state := &petruntime.Service{}
	service := &Service{config: petpolicy.Config{MaxPerRoom: 0, MaxPerOwnerRoom: 0}, runtime: state}
	if err := service.checkRoomLimits(context.Background(), 9, 7); err != petrecord.ErrRoomLimit {
		t.Fatalf("expected room limit, got %v", err)
	}
	service.permissions = policyChecker{allowed: petpolicy.RoomLimitBypass}
	if err := service.checkRoomLimits(context.Background(), 9, 7); err != nil {
		t.Fatalf("expected room bypass, got %v", err)
	}
}

// TestExpectedPresenceErrors verifies protocol-safe failures remain classified.
func TestExpectedPresenceErrors(t *testing.T) {
	for _, err := range []error{petrecord.ErrPetNotFound, petrecord.ErrNoRights, petrecord.ErrInventoryLimit, petrecord.ErrRoomLimit, petrecord.ErrTileNotFree, petrecord.ErrPetsDisabled, petrecord.ErrInvalidState, petrecord.ErrConflict} {
		if !IsExpected(err) {
			t.Fatalf("expected safe classification for %v", err)
		}
	}
}
