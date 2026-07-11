package info

import (
	"testing"

	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
)

// TestChatMapsRoomChatSettings verifies room chat projection.
func TestChatMapsRoomChatSettings(t *testing.T) {
	settings := chat(roommodel.Room{ChatMode: 1, ChatWeight: 2, ChatSpeed: 3, ChatDistance: 4, ChatProtection: 5})
	if settings.Mode != 1 || settings.Weight != 2 || settings.Speed != 3 || settings.Distance != 4 || settings.Protection != 5 {
		t.Fatalf("unexpected chat settings %#v", settings)
	}
}

// TestModerationMapsDedicatedPolicies verifies room info does not reuse chat protection.
func TestModerationMapsDedicatedPolicies(t *testing.T) {
	settings := moderation(roommodel.Room{ModerationMute: 1, ModerationKick: 2, ModerationBan: 0, ChatProtection: 2})
	if settings.AllowMute != 1 || settings.AllowKick != 2 || settings.AllowBan != 0 {
		t.Fatalf("unexpected moderation settings %#v", settings)
	}
}
