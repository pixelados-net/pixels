package model

import (
	"testing"
	"time"
)

// TestIdentityEmpty verifies empty identity detection.
func TestIdentityEmpty(t *testing.T) {
	if !(Identity{}).Empty() {
		t.Fatal("expected empty identity")
	}

	if (Identity{ID: 1}).Empty() {
		t.Fatal("expected assigned identity")
	}
}

// TestTimestampsTouch verifies update timestamp mutation.
func TestTimestampsTouch(t *testing.T) {
	now := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	timestamps := Timestamps{}

	timestamps.Touch(now)

	if !timestamps.UpdatedAt.Equal(now) {
		t.Fatalf("expected updated at %s, got %s", now, timestamps.UpdatedAt)
	}
}

// TestSoftDeleteState verifies active and deleted state helpers.
func TestSoftDeleteState(t *testing.T) {
	now := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)

	if !(SoftDelete{}).Active() {
		t.Fatal("expected active record")
	}

	if !(SoftDelete{DeletedAt: &now}).Deleted() {
		t.Fatal("expected deleted record")
	}
}

// TestVersionNext verifies optimistic locking version increments.
func TestVersionNext(t *testing.T) {
	if (Version{Version: 4}).Next() != 5 {
		t.Fatal("expected next version")
	}
}
