package model

import (
	"testing"
	"time"
)

// TestGenderValid verifies supported profile gender values.
func TestGenderValid(t *testing.T) {
	if !GenderMale.Valid() {
		t.Fatal("expected male gender to be valid")
	}

	if !GenderFemale.Valid() {
		t.Fatal("expected female gender to be valid")
	}

	if Gender("X").Valid() {
		t.Fatal("expected unknown gender to be invalid")
	}
}

// TestProfileCreated verifies profile ownership detection.
func TestProfileCreated(t *testing.T) {
	if (Profile{}).Created() {
		t.Fatal("expected empty profile")
	}

	if !(Profile{PlayerID: 1}).Created() {
		t.Fatal("expected created profile")
	}
}

// TestProfileUpdatedAfter verifies profile update time comparison.
func TestProfileUpdatedAfter(t *testing.T) {
	now := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	profile := Profile{}
	profile.UpdatedAt = now

	if !profile.UpdatedAfter(now.Add(-time.Second)) {
		t.Fatal("expected profile to be updated after earlier time")
	}
}
