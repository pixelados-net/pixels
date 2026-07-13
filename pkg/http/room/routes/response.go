// Package routes contains protected room administration routes.
package routes

// ListResponse contains room list results.
type ListResponse struct {
	// Total stores the returned room count.
	Total int `json:"total"`
	// Items stores safe room rows.
	Items []RoomResponse `json:"items"`
}

// RoomResponse contains safe room metadata.
type RoomResponse struct {
	// ID identifies the room.
	ID int64 `json:"id"`
	// Name stores the room name.
	Name string `json:"name"`
	// OwnerPlayerID identifies the owner.
	OwnerPlayerID int64 `json:"ownerPlayerId"`
	// OwnerName stores the owner snapshot.
	OwnerName string `json:"ownerName"`
	// ModelName stores the room layout model.
	ModelName string `json:"modelName"`
	// MaxUsers stores the room capacity.
	MaxUsers int `json:"maxUsers"`
	// CategoryID stores the optional category id.
	CategoryID *int64 `json:"categoryId,omitempty"`
	// Score stores the navigator score.
	Score int `json:"score"`
	// IsBundleTemplate reports whether the room is a bundle source.
	IsBundleTemplate bool `json:"isBundleTemplate"`
}

// OccupancyResponse contains active room occupancy.
type OccupancyResponse struct {
	// RoomID identifies the room.
	RoomID int64 `json:"roomId"`
	// Count stores active occupant count.
	Count int `json:"count"`
	// MaxUsers stores maximum occupancy.
	MaxUsers int `json:"maxUsers"`
	// PlayerIDs stores active player ids.
	PlayerIDs []int64 `json:"playerIds"`
}

// ForwardRequest contains room forwarding input.
type ForwardRequest struct {
	// TargetRoomID identifies the room clients should enter.
	TargetRoomID int64 `json:"targetRoomId"`
}

// TeleportRequest contains one player forwarding request.
type TeleportRequest struct {
	// TargetRoomID identifies the destination room.
	TargetRoomID int64 `json:"targetRoomId"`
	// Bypass reports whether closed-room gating should be skipped once.
	Bypass bool `json:"bypass"`
}

// ActionResponse contains admin runtime action counts.
type ActionResponse struct {
	// Matched stores matched runtime occupants.
	Matched int `json:"matched"`
	// Sent stores successful packet sends.
	Sent int `json:"sent"`
	// Errors stores failed packet sends.
	Errors int `json:"errors"`
}

// CategoryListResponse contains navigator room categories.
type CategoryListResponse struct {
	// Total stores the returned category count.
	Total int `json:"total"`
	// Items stores category rows.
	Items []CategoryResponse `json:"items"`
}

// CategoryResponse contains safe room category data.
type CategoryResponse struct {
	// ID identifies the category.
	ID int64 `json:"id"`
	// Caption stores the visible caption.
	Caption string `json:"caption"`
	// Visible reports whether the category is visible.
	Visible bool `json:"visible"`
	// Order stores navigator order.
	Order int `json:"order"`
}

// LiftedListResponse contains navigator lifted rooms.
type LiftedListResponse struct {
	// Total stores the returned lifted room count.
	Total int `json:"total"`
	// Items stores lifted room rows.
	Items []LiftedResponse `json:"items"`
}

// LiftedResponse contains safe lifted room data.
type LiftedResponse struct {
	// ID identifies the lifted row.
	ID int64 `json:"id"`
	// RoomID identifies the promoted room.
	RoomID int64 `json:"roomId"`
	// AreaID stores the visual area id.
	AreaID int `json:"areaId"`
	// Image stores the image key.
	Image string `json:"image"`
	// Caption stores the caption.
	Caption string `json:"caption"`
}
