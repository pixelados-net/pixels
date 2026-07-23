package access

import (
	"context"
	"errors"
	"testing"

	"github.com/niflaot/pixels/internal/permission"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
)

// fixedChecker stores one permission result for authorization tests.
type fixedChecker struct {
	// allowed stores the permission decision.
	allowed bool
	// err stores an optional resolution failure.
	err error
}

// HasPermission returns the configured permission decision.
func (checker fixedChecker) HasPermission(context.Context, int64, permission.Node) (bool, error) {
	return checker.allowed, checker.err
}

// TestCanManageAllowsLocalAndGlobalAuthority verifies both supported authorization paths.
func TestCanManageAllowsLocalAndGlobalAuthority(t *testing.T) {
	active, err := roomlive.NewRoom(roomlive.Snapshot{ID: 1, OwnerPlayerID: 7, MaxUsers: 10})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	for _, test := range []struct {
		name     string
		playerID int64
		checker  fixedChecker
		want     bool
	}{
		{name: "owner", playerID: 7, want: true},
		{name: "global", playerID: 8, checker: fixedChecker{allowed: true}, want: true},
		{name: "guest", playerID: 8, want: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			allowed, checkErr := CanManage(context.Background(), test.checker, active, test.playerID)
			if checkErr != nil || allowed != test.want {
				t.Fatalf("allowed=%v err=%v", allowed, checkErr)
			}
		})
	}
}

// TestCanManagePropagatesPermissionFailure verifies infrastructure errors remain explicit.
func TestCanManagePropagatesPermissionFailure(t *testing.T) {
	expected := errors.New("permission unavailable")
	_, err := CanManage(context.Background(), fixedChecker{err: expected}, nil, 8)
	if !errors.Is(err, expected) {
		t.Fatalf("expected permission error, got %v", err)
	}
}

// BenchmarkCanManageLocal measures the room-rights hot path.
func BenchmarkCanManageLocal(b *testing.B) {
	active, err := roomlive.NewRoom(roomlive.Snapshot{ID: 1, OwnerPlayerID: 7, MaxUsers: 10})
	if err != nil {
		b.Fatalf("create room: %v", err)
	}
	b.ReportAllocs()
	for b.Loop() {
		_, _ = CanManage(context.Background(), nil, active, 7)
	}
}
