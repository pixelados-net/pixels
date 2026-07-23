package routes

// FilterRequest contains one global filter mutation.
type FilterRequest struct {
	// Word stores the whole word to add.
	Word string `json:"word"`
}

// BubbleRequest contains one unlock threshold mutation.
type BubbleRequest struct {
	// MinWeight stores the minimum primary-group weight.
	MinWeight int32 `json:"minWeight"`
}
