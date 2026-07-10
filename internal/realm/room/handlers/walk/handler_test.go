package walk

import (
	"testing"

	walkcmd "github.com/niflaot/pixels/internal/realm/room/commands/walk"
	roomlive "github.com/niflaot/pixels/internal/realm/room/live"
)

// TestStalePresenceErrorsAreSoft verifies post-leave walk packets are harmless.
func TestStalePresenceErrorsAreSoft(t *testing.T) {
	if !isStalePresenceError(walkcmd.ErrPlayerNotInRoom) || !isStalePresenceError(roomlive.ErrRoomNotFound) {
		t.Fatal("expected stale room presence errors to be soft")
	}
}
