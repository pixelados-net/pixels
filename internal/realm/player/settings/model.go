// Package settings owns durable player client settings.
package settings

// Record contains persisted settings used by Nitro and live snapshots.
type Record struct {
	// PlayerID identifies the owning player.
	PlayerID int64
	// VolumeSystem stores system volume from zero through one hundred.
	VolumeSystem int32
	// VolumeFurniture stores furniture volume from zero through one hundred.
	VolumeFurniture int32
	// VolumeTrax stores music volume from zero through one hundred.
	VolumeTrax int32
	// OldChat reports whether legacy chat rendering is selected.
	OldChat bool
	// CameraFollowBlocked reports whether automatic camera following is disabled.
	CameraFollowBlocked bool
	// SafetyLocked reports immutable hotel-side safety state.
	SafetyLocked bool
	// Version stores optimistic persistence state.
	Version int64
}

// AdminPatch contains optional protected settings mutations.
type AdminPatch struct {
	// VolumeSystem replaces system volume.
	VolumeSystem *int32
	// VolumeFurniture replaces furniture volume.
	VolumeFurniture *int32
	// VolumeTrax replaces music volume.
	VolumeTrax *int32
	// OldChat replaces legacy chat selection.
	OldChat *bool
	// CameraFollowBlocked replaces camera-follow privacy.
	CameraFollowBlocked *bool
	// SafetyLocked replaces server-controlled safety state.
	SafetyLocked *bool
}

// Default returns neutral settings for a player without a persisted row.
func Default(playerID int64) Record {
	return Record{PlayerID: playerID, VolumeSystem: 100, VolumeFurniture: 100, VolumeTrax: 100, Version: 1}
}
