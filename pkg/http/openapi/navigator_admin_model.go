package openapi

// NavigatorPlayerRequest documents a player-scoped navigator read.
type NavigatorPlayerRequest struct {
	APIKeyRequest
	// PlayerID identifies the target player.
	PlayerID int64 `query:"playerId" required:"true" minimum:"1"`
}

// NavigatorRoomIDsResponse documents an ordered player room list.
type NavigatorRoomIDsResponse struct {
	// PlayerID identifies the target player.
	PlayerID int64 `json:"playerId" required:"true"`
	// RoomIDs stores ordered room identifiers.
	RoomIDs []int64 `json:"roomIds" required:"true"`
}

// NavigatorDeleteResponse documents deleted history rows.
type NavigatorDeleteResponse struct {
	// PlayerID identifies the target player.
	PlayerID int64 `json:"playerId" required:"true"`
	// Deleted stores removed visit rows.
	Deleted int64 `json:"deleted" required:"true"`
}

// NavigatorHistoryDeleteRequest documents an attributed history deletion.
type NavigatorHistoryDeleteRequest struct {
	NavigatorPlayerRequest
	// ActorPlayerID identifies the administrative actor.
	ActorPlayerID int64 `json:"actorPlayerId" required:"true" minimum:"1"`
	// Reason explains the administrative mutation.
	Reason string `json:"reason" required:"true" minLength:"1" maxLength:"500"`
}

// NavigatorOfficialRequest documents an attributed optimistic official-room mutation.
type NavigatorOfficialRequest struct {
	APIKeyRequest
	// RoomID identifies the target room.
	RoomID int64 `path:"roomId" required:"true" minimum:"1"`
	// ExpectedVersion stores the current room version.
	ExpectedVersion int64 `json:"expectedVersion" required:"true" minimum:"1"`
	// ActorPlayerID identifies the administrative actor.
	ActorPlayerID int64 `json:"actorPlayerId" required:"true" minimum:"1"`
	// Reason explains the administrative mutation.
	Reason string `json:"reason" required:"true" minLength:"1" maxLength:"500"`
}

// CategoryListResponse contains navigator room categories.
type CategoryListResponse struct {
	// Total stores the returned category count.
	Total int `json:"total" required:"true"`
	// Items stores category rows.
	Items []CategoryResponse `json:"items" required:"true"`
}

// CategoryResponse contains safe category data.
type CategoryResponse struct {
	// ID identifies the category.
	ID int64 `json:"id" required:"true"`
	// Caption stores the visible caption.
	Caption string `json:"caption" required:"true"`
	// Visible reports whether the category is visible.
	Visible bool `json:"visible" required:"true"`
	// Order stores navigator ordering.
	Order int `json:"order" required:"true"`
}

// LiftedListResponse contains navigator lifted rooms.
type LiftedListResponse struct {
	// Total stores returned lifted room count.
	Total int `json:"total" required:"true"`
	// Items stores lifted room rows.
	Items []LiftedResponse `json:"items" required:"true"`
}

// LiftedResponse contains safe lifted room data.
type LiftedResponse struct {
	// ID identifies the lifted row.
	ID int64 `json:"id" required:"true"`
	// RoomID identifies the promoted room.
	RoomID int64 `json:"roomId" required:"true"`
	// AreaID stores the visual area id.
	AreaID int `json:"areaId" required:"true"`
	// Image stores the image key.
	Image string `json:"image" required:"true"`
	// Caption stores the caption.
	Caption string `json:"caption" required:"true"`
}
