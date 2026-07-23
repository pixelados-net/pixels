package routes

import (
	bubblerepo "github.com/niflaot/pixels/internal/realm/chat/bubble/repository"
	historymodel "github.com/niflaot/pixels/internal/realm/chat/history/model"
)

// FilterListResponse contains the global filter dictionary.
type FilterListResponse struct {
	// Total stores dictionary size.
	Total int `json:"total"`
	// Items stores normalized words.
	Items []string `json:"items"`
}

// BubbleListResponse contains configured bubble thresholds.
type BubbleListResponse struct {
	// Total stores threshold count.
	Total int `json:"total"`
	// Items stores configured thresholds.
	Items []bubblerepo.Unlock `json:"items"`
}

// HistoryResponse contains one bounded chat history page.
type HistoryResponse struct {
	// Total stores page size.
	Total int `json:"total"`
	// Items stores history entries.
	Items []historymodel.Entry `json:"items"`
}

// MutationResponse confirms one chat administration mutation.
type MutationResponse struct {
	// Updated reports whether the operation completed.
	Updated bool `json:"updated"`
}
