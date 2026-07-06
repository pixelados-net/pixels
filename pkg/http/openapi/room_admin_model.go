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
