package service

import (
	"context"
	"errors"
	"testing"
)

// fakeAssigner records default permission group assignments.
type fakeAssigner struct {
	// playerID stores the assigned player id.
	playerID int64
	// err stores the assignment failure.
	err error
}

// AssignDefaultGroup records one player assignment.
func (assigner *fakeAssigner) AssignDefaultGroup(_ context.Context, playerID int64) error {
	assigner.playerID = playerID
	return assigner.err
}

// TestCreateAssignsDefaultPermissionGroup verifies new players become members.
func TestCreateAssignsDefaultPermissionGroup(t *testing.T) {
	assigner := &fakeAssigner{}
	record, err := New(newFakeStore(), assigner).Create(context.Background(), CreateParams{Username: "ian"})
	if err != nil || assigner.playerID != record.Player.ID {
		t.Fatalf("unexpected player=%#v assigned=%d err=%v", record.Player, assigner.playerID, err)
	}
}

// TestCreateReportsDefaultPermissionAssignmentFailure verifies assignment errors are actionable.
func TestCreateReportsDefaultPermissionAssignmentFailure(t *testing.T) {
	failure := errors.New("member group unavailable")
	_, err := New(newFakeStore(), &fakeAssigner{err: failure}).Create(context.Background(), CreateParams{Username: "ian"})
	if !errors.Is(err, failure) || err.Error() == failure.Error() {
		t.Fatalf("expected wrapped assignment failure, got %v", err)
	}
}
