package routes

import (
	"testing"

	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
)

// TestApplyPrivacyPreservesOmittedFields verifies partial privacy updates.
func TestApplyPrivacyPreservesOmittedFields(t *testing.T) {
	blocked := true
	params := playerservice.PrivacyParams{BlockFriendRequests: false, BlockRoomInvites: false, BlockFollowing: true}
	applyPrivacy(&params, PrivacyRequest{BlockRoomInvites: &blocked})
	if params.BlockFriendRequests || !params.BlockRoomInvites || !params.BlockFollowing {
		t.Fatalf("unexpected privacy params %#v", params)
	}
}
