package info

import (
	"testing"

	roommodel "github.com/niflaot/pixels/internal/realm/room/model"
)

// TestChatMapsRoomChatSettings verifies room chat projection.
func TestChatMapsRoomChatSettings(t *testing.T) {
	settings := chat(roommodel.Room{ChatMode: 1, ChatWeight: 2, ChatSpeed: 3, ChatDistance: 4, ChatProtection: 5})
	if settings.Mode != 1 || settings.Weight != 2 || settings.Speed != 3 || settings.Distance != 4 || settings.Protection != 5 {
		t.Fatalf("unexpected chat settings %#v", settings)
	}
}
