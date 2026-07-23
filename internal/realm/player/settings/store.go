package settings

import "context"

// Store persists user settings and home-room selection.
type Store interface {
	// Find returns persisted settings or their defaults.
	Find(context.Context, int64) (Record, error)
	// SetVolume replaces all three volume fields.
	SetVolume(context.Context, int64, int32, int32, int32) (Record, error)
	// SetOldChat replaces legacy chat selection.
	SetOldChat(context.Context, int64, bool) (Record, error)
	// SetCameraFollowBlocked replaces camera-follow privacy.
	SetCameraFollowBlocked(context.Context, int64, bool) (Record, error)
	// SetHomeRoom replaces or clears the profile home room.
	SetHomeRoom(context.Context, int64, *int64) error
}

// AdminStore applies optimistic protected settings mutations.
type AdminStore interface {
	// UpdateAdmin applies one protected partial settings mutation.
	UpdateAdmin(context.Context, int64, int64, AdminPatch) (Record, error)
}
