package social

import (
	"testing"

	messengerservice "github.com/niflaot/pixels/internal/realm/messenger/core"
)

// TestFollowErrorCodeMapsDomainFailures verifies native follow error mapping.
func TestFollowErrorCodeMapsDomainFailures(t *testing.T) {
	cases := map[error]int32{
		messengerservice.ErrNotFriend:        0,
		messengerservice.ErrFriendOffline:    1,
		messengerservice.ErrFriendNotInRoom:  2,
		messengerservice.ErrFollowingBlocked: 3,
	}
	for failure, expected := range cases {
		if actual := followErrorCode(failure); actual != expected {
			t.Fatalf("expected %d for %v, got %d", expected, failure, actual)
		}
	}
}
