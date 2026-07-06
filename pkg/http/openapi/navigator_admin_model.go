package openapi

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
