package openapi

// AdminPlayerSettingsRequest documents an optimistic attributed settings mutation.
type AdminPlayerSettingsRequest struct {
	AdminPlayerPathRequest
	// ExpectedVersion stores the current settings version.
	ExpectedVersion int64 `json:"expectedVersion" required:"true" minimum:"1"`
	// ActorPlayerID identifies the administrative actor.
	ActorPlayerID int64 `json:"actorPlayerId" required:"true" minimum:"1"`
	// Reason explains the administrative mutation.
	Reason string `json:"reason" required:"true" minLength:"1" maxLength:"500"`
	// VolumeSystem optionally replaces system volume.
	VolumeSystem *int32 `json:"volumeSystem,omitempty" minimum:"0" maximum:"100"`
	// VolumeFurniture optionally replaces furniture volume.
	VolumeFurniture *int32 `json:"volumeFurniture,omitempty" minimum:"0" maximum:"100"`
	// VolumeTrax optionally replaces music volume.
	VolumeTrax *int32 `json:"volumeTrax,omitempty" minimum:"0" maximum:"100"`
	// OldChat optionally replaces legacy chat rendering.
	OldChat *bool `json:"oldChat,omitempty"`
	// CameraFollowBlocked optionally replaces camera-follow privacy.
	CameraFollowBlocked *bool `json:"cameraFollowBlocked,omitempty"`
	// SafetyLocked optionally replaces server-controlled safety state.
	SafetyLocked *bool `json:"safetyLocked,omitempty"`
}

// AdminPlayerTagsRequest documents an attributed complete tag replacement.
type AdminPlayerTagsRequest struct {
	AdminPlayerPathRequest
	// Tags stores at most five normalized public tags.
	Tags []string `json:"tags" required:"true" maxItems:"5"`
	// ActorPlayerID identifies the administrative actor.
	ActorPlayerID int64 `json:"actorPlayerId" required:"true" minimum:"1"`
	// Reason explains the administrative mutation.
	Reason string `json:"reason" required:"true" minLength:"1" maxLength:"500"`
}

// AdminPlayerAttributedRequest documents an attributed player mutation.
type AdminPlayerAttributedRequest struct {
	AdminPlayerPathRequest
	// ActorPlayerID identifies the administrative actor.
	ActorPlayerID int64 `json:"actorPlayerId" required:"true" minimum:"1"`
	// Reason explains the administrative mutation.
	Reason string `json:"reason" required:"true" minLength:"1" maxLength:"500"`
}

// AdminPlayerSettingsResponse documents durable client settings.
type AdminPlayerSettingsResponse struct {
	// PlayerID identifies the settings owner.
	PlayerID int64 `json:"playerId" required:"true"`
	// VolumeSystem stores system volume.
	VolumeSystem int32 `json:"volumeSystem" required:"true"`
	// VolumeFurniture stores furniture volume.
	VolumeFurniture int32 `json:"volumeFurniture" required:"true"`
	// VolumeTrax stores music volume.
	VolumeTrax int32 `json:"volumeTrax" required:"true"`
	// OldChat reports legacy chat selection.
	OldChat bool `json:"oldChat" required:"true"`
	// CameraFollowBlocked reports camera-follow privacy.
	CameraFollowBlocked bool `json:"cameraFollowBlocked" required:"true"`
	// SafetyLocked reports server-controlled safety state.
	SafetyLocked bool `json:"safetyLocked" required:"true"`
	// Version stores optimistic settings state.
	Version int64 `json:"version" required:"true"`
}

// AdminPlayerTagsResponse documents one complete tag replacement.
type AdminPlayerTagsResponse struct {
	// PlayerID identifies the profile owner.
	PlayerID int64 `json:"playerId" required:"true"`
	// Tags stores normalized public tags.
	Tags []string `json:"tags" required:"true"`
}

// AdminPlayerOutfit documents one saved wardrobe slot.
type AdminPlayerOutfit struct {
	// SlotID identifies the wardrobe slot.
	SlotID int32 `json:"SlotID" required:"true" minimum:"1" maximum:"10"`
	// Figure stores the validated Nitro figure.
	Figure string `json:"Figure" required:"true" maxLength:"512"`
	// Gender stores the avatar gender.
	Gender string `json:"Gender" required:"true" enum:"M,F"`
}

// AdminPlayerClothing documents unlocked clothing sets.
type AdminPlayerClothing struct {
	// FigureSetIDs stores unlocked figure set identifiers.
	FigureSetIDs []int32 `json:"FigureSetIDs" required:"true"`
	// ProductCodes stores unlocked product codes.
	ProductCodes []string `json:"ProductCodes" required:"true"`
}

// AdminPlayerWardrobeResponse documents saved outfits and clothing unlocks.
type AdminPlayerWardrobeResponse struct {
	// Outfits stores saved wardrobe slots.
	Outfits []AdminPlayerOutfit `json:"outfits" required:"true"`
	// Clothing stores unlocked figure sets and product codes.
	Clothing AdminPlayerClothing `json:"clothing" required:"true"`
}
