package model

import (
	"testing"
	"time"
)

// TestActionAndBanDurationValidation verifies closed moderation enums.
func TestActionAndBanDurationValidation(t *testing.T) {
	for _, action := range []Action{ActionKick, ActionMute, ActionUnmute, ActionBan, ActionUnban} {
		if !action.Valid() {
			t.Fatalf("expected valid action %q", action)
		}
	}
	if Action("invalid").Valid() {
		t.Fatal("expected invalid action")
	}
	tests := map[BanDuration]time.Duration{
		BanDurationHour: time.Hour, BanDurationDay: 24 * time.Hour,
		BanDurationPermanent: 10 * 365 * 24 * time.Hour,
	}
	for duration, expected := range tests {
		actual, valid := duration.Duration()
		if !valid || actual != expected {
			t.Fatalf("duration %q=%v valid=%v", duration, actual, valid)
		}
	}
	if _, valid := BanDuration("invalid").Duration(); valid {
		t.Fatal("expected invalid ban duration")
	}
}
