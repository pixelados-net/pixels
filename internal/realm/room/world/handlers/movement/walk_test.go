package movement

import (
	"testing"

	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	walkcmd "github.com/niflaot/pixels/internal/realm/room/world/commands/walk"
)

// TestStalePresenceErrorsAreSoft verifies post-leave walk packets are harmless.
func TestWalkStalePresenceErrorsAreSoft(t *testing.T) {
	if !isStalePresenceError(walkcmd.ErrPlayerNotInRoom) || !isStalePresenceError(roomlive.ErrRoomNotFound) {
		t.Fatal("expected stale room presence errors to be soft")
	}
}
