package openapi

// RoomListRequest contains room list filters.
type RoomListRequest struct {
	APIKeyRequest
	// Query stores an optional search query.
	Query string `query:"query,omitempty"`
	// Limit stores an optional result limit.
	Limit int `query:"limit,omitempty" minimum:"1" maximum:"100"`
}

// RoomIDRequest contains a room id path parameter.
type RoomIDRequest struct {
	APIKeyRequest
	// ID identifies the room.
	ID int64 `path:"id" required:"true" minimum:"1"`
}

// RoomForwardRequest contains room forward input.
type RoomForwardRequest struct {
	RoomIDRequest
	// TargetRoomID identifies the room clients should enter.
	TargetRoomID int64 `json:"targetRoomId" required:"true" minimum:"1"`
}

// RoomTeleportRequest contains a single-player room forwarding request.
type RoomTeleportRequest struct {
	APIKeyRequest
	// PlayerID identifies the live player.
	PlayerID int64 `path:"playerId" required:"true" minimum:"1"`
	// TargetRoomID identifies the destination room.
	TargetRoomID int64 `json:"targetRoomId" required:"true" minimum:"1" maximum:"2147483647"`
	// Bypass skips password, doorbell, and invisible gating once.
	Bypass bool `json:"bypass"`
}

// RoomListResponse contains room list results.
type RoomListResponse struct {
	// Total stores the returned room count.
	Total int `json:"total" required:"true"`
	// Items stores safe room rows.
	Items []RoomResponse `json:"items" required:"true"`
}

// RoomResponse contains safe room metadata.
type RoomResponse struct {
	// ID identifies the room.
	ID int64 `json:"id" required:"true"`
	// Name stores the room name.
	Name string `json:"name" required:"true"`
	// OwnerPlayerID identifies the owner.
	OwnerPlayerID int64 `json:"ownerPlayerId" required:"true"`
	// OwnerName stores the owner snapshot.
	OwnerName string `json:"ownerName" required:"true"`
	// ModelName stores the layout model.
	ModelName string `json:"modelName" required:"true"`
	// MaxUsers stores room capacity.
	MaxUsers int `json:"maxUsers" required:"true"`
	// CategoryID stores the optional category id.
	CategoryID *int64 `json:"categoryId,omitempty"`
	// Score stores navigator score.
	Score int `json:"score" required:"true"`
}

// RoomOccupancyResponse contains active room occupancy.
type RoomOccupancyResponse struct {
	// RoomID identifies the room.
	RoomID int64 `json:"roomId" required:"true"`
	// Count stores active occupant count.
	Count int `json:"count" required:"true"`
	// MaxUsers stores maximum occupancy.
	MaxUsers int `json:"maxUsers" required:"true"`
	// PlayerIDs stores active player ids.
	PlayerIDs []int64 `json:"playerIds" required:"true"`
}

// RoomActionResponse contains runtime action counts.
type RoomActionResponse struct {
	// Matched stores matched runtime occupants.
	Matched int `json:"matched" required:"true"`
	// Sent stores successful packet sends.
	Sent int `json:"sent" required:"true"`
	// Errors stores failed packet sends.
	Errors int `json:"errors" required:"true"`
}

// RoomAuditRequest contains room audit filters.
type RoomAuditRequest struct {
	RoomIDRequest
	// Before stores an optional descending id cursor.
	Before int64 `query:"before,omitempty" minimum:"1"`
	// Limit caps returned history rows.
	Limit int `query:"limit,omitempty" minimum:"1" maximum:"200"`
	// Type optionally filters comma-separated moderation actions.
	Type string `query:"type,omitempty"`
}

// PlayerAuditRequest contains player audit filters.
type PlayerAuditRequest struct {
	APIKeyRequest
	// PlayerID identifies the affected or acting player.
	PlayerID int64 `path:"playerId" required:"true" minimum:"1"`
	// RoomID optionally restricts one room.
	RoomID int64 `query:"roomId,omitempty" minimum:"1"`
	// Before stores an optional descending id cursor.
	Before int64 `query:"before,omitempty" minimum:"1"`
	// Limit caps returned history rows.
	Limit int `query:"limit,omitempty" minimum:"1" maximum:"200"`
	// Type optionally filters comma-separated moderation actions.
	Type string `query:"type,omitempty"`
}

// RoomRightsAuditEntry describes one rights history row.
type RoomRightsAuditEntry struct {
	// ID identifies the audit row.
	ID int64 `json:"id" required:"true"`
	// RoomID identifies the room.
	RoomID int64 `json:"roomId" required:"true"`
	// PlayerID identifies the affected player.
	PlayerID int64 `json:"playerId" required:"true"`
	// ActorKind identifies the source family.
	ActorKind string `json:"actorKind" required:"true"`
	// ActorID optionally identifies the acting player.
	ActorID *int64 `json:"actorId,omitempty"`
	// Action identifies the rights mutation.
	Action string `json:"action" required:"true"`
	// CreatedAt stores the mutation time.
	CreatedAt string `json:"createdAt" required:"true" format:"date-time"`
}

// RoomModerationAuditEntry describes one moderation history row.
type RoomModerationAuditEntry struct {
	// ID identifies the audit row.
	ID int64 `json:"id" required:"true"`
	// RoomID identifies the room.
	RoomID int64 `json:"roomId" required:"true"`
	// TargetPlayerID identifies the affected player.
	TargetPlayerID int64 `json:"targetPlayerId" required:"true"`
	// ActorKind identifies the source family.
	ActorKind string `json:"actorKind" required:"true"`
	// ActorID optionally identifies the acting player.
	ActorID *int64 `json:"actorId,omitempty"`
	// Action identifies the moderation action.
	Action string `json:"action" required:"true"`
	// DurationSeconds optionally stores sanction duration.
	DurationSeconds *int64 `json:"durationSeconds,omitempty"`
	// ExpiresAt optionally stores sanction expiry.
	ExpiresAt *string `json:"expiresAt,omitempty" format:"date-time"`
	// CreatedAt stores the action time.
	CreatedAt string `json:"createdAt" required:"true" format:"date-time"`
}

// RoomSanction describes one active room ban or mute.
type RoomSanction struct {
	// RoomID identifies the room.
	RoomID int64 `json:"roomId" required:"true"`
	// PlayerID identifies the affected player.
	PlayerID int64 `json:"playerId" required:"true"`
	// Username stores the current username.
	Username string `json:"username" required:"true"`
	// EndsAt stores sanction expiry.
	EndsAt string `json:"endsAt" required:"true" format:"date-time"`
	// CreatedAt stores sanction creation time.
	CreatedAt string `json:"createdAt" required:"true" format:"date-time"`
	// UpdatedAt stores sanction update time.
	UpdatedAt string `json:"updatedAt" required:"true" format:"date-time"`
}

// RoomRightsAuditResponse contains room rights history.
type RoomRightsAuditResponse struct {
	// Total stores returned rows.
	Total int `json:"total" required:"true"`
	// Items stores rights audit rows.
	Items []RoomRightsAuditEntry `json:"items" required:"true"`
}

// RoomModerationAuditResponse contains moderation history.
type RoomModerationAuditResponse struct {
	// Total stores returned rows.
	Total int `json:"total" required:"true"`
	// Items stores moderation rows.
	Items []RoomModerationAuditEntry `json:"items" required:"true"`
}

// RoomSanctionResponse contains current sanctions.
type RoomSanctionResponse struct {
	// Total stores returned rows.
	Total int `json:"total" required:"true"`
	// Items stores current sanctions.
	Items []RoomSanction `json:"items" required:"true"`
}
