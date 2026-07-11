package settings

import (
	"context"
	"errors"
	"testing"

	"github.com/niflaot/pixels/internal/permission"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// permissionsForTest stores actor permission decisions.
type permissionsForTest map[int64]map[permission.Node]bool

// HasPermission reports one configured permission decision.
func (permissions permissionsForTest) HasPermission(_ context.Context, playerID int64, node permission.Node) (bool, error) {
	return permissions[playerID][node], nil
}

// TestAuthorizerAppliesStaffAndOwnerPolicy verifies settings authorization paths.
func TestAuthorizerAppliesStaffAndOwnerPolicy(t *testing.T) {
	nodes := Nodes{OwnManage: "own", AnyManage: "any", OwnPolicyManage: "policy.own", AnyPolicyManage: "policy.any"}
	room := roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 7}}, OwnerPlayerID: 1}
	authorizer := New(permissionsForTest{1: {"own": true, "policy.own": true}, 2: {"own": true}, 3: {"any": true, "policy.any": true}}, nodes)
	for _, actorID := range []int64{1, 3} {
		allowed, err := authorizer.CanManage(context.Background(), room, actorID)
		if err != nil || !allowed {
			t.Fatalf("actor %d allowed=%v err=%v", actorID, allowed, err)
		}
	}
	if allowed, err := authorizer.CanManage(context.Background(), room, 2); err != nil || allowed {
		t.Fatalf("rights holder must not manage settings allowed=%v err=%v", allowed, err)
	}
	if err := authorizer.AuthorizePolicy(context.Background(), room, 1); err != nil {
		t.Fatalf("authorize owner policy: %v", err)
	}
	if err := authorizer.AuthorizePolicy(context.Background(), room, 3); err != nil {
		t.Fatalf("authorize global policy: %v", err)
	}
	if err := authorizer.AuthorizePolicy(context.Background(), room, 2); !errors.Is(err, ErrAccessDenied) {
		t.Fatalf("expected rights-holder policy denial, got %v", err)
	}
	if err := authorizer.Authorize(context.Background(), room, 4); !errors.Is(err, ErrAccessDenied) {
		t.Fatalf("expected access denied, got %v", err)
	}
}

// TestAuthorizerRejectsInvalidAndMissingCapabilities verifies defensive authorization paths.
func TestAuthorizerRejectsInvalidAndMissingCapabilities(t *testing.T) {
	room := roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 7}}, OwnerPlayerID: 1}
	authorizer := New(nil, Nodes{OwnManage: "own", AnyManage: "any"})
	if allowed, err := authorizer.CanManage(context.Background(), room, 0); !errors.Is(err, ErrAccessDenied) || allowed {
		t.Fatalf("invalid actor allowed=%v err=%v", allowed, err)
	}
	if allowed, err := authorizer.CanManage(context.Background(), room, 2); err != nil || allowed {
		t.Fatalf("guest allowed=%v err=%v", allowed, err)
	}
	if allowed, err := authorizer.CanManageAny(context.Background(), 1); err != nil || allowed {
		t.Fatalf("unexpected global capability allowed=%v err=%v", allowed, err)
	}
}

// BenchmarkCanManageOwner measures owner authorization overhead above persistence.
func BenchmarkCanManageOwner(b *testing.B) {
	authorizer := New(permissionsForTest{1: {"own": true}}, Nodes{OwnManage: "own", AnyManage: "any"})
	room := roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 7}}, OwnerPlayerID: 1}
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		allowed, err := authorizer.CanManage(ctx, room, 1)
		if err != nil || !allowed {
			b.Fatal("owner authorization failed")
		}
	}
}
