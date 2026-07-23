package friend

import (
	"testing"

	messengerservice "github.com/niflaot/pixels/internal/realm/messenger/core"
)

// TestRequestErrorCodeMapsDomainFailures verifies native friend-request error mapping.
func TestRequestErrorCodeMapsDomainFailures(t *testing.T) {
	cases := map[error]int32{
		messengerservice.ErrOwnListFull:     1,
		messengerservice.ErrTargetListFull:  2,
		messengerservice.ErrRequestsBlocked: 3,
		messengerservice.ErrPlayerNotFound:  4,
	}
	for failure, expected := range cases {
		if actual := requestErrorCode(failure); actual != expected {
			t.Fatalf("expected %d for %v, got %d", expected, failure, actual)
		}
	}
}
