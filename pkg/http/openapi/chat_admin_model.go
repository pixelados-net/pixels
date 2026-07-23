package openapi

// ChatFilterRequest contains one global filter mutation.
type ChatFilterRequest struct {
	APIKeyRequest
	// Word stores the normalized whole word.
	Word string `json:"word" required:"true" minLength:"1" maxLength:"32"`
}

// ChatFilterDeleteRequest contains one filter path word.
type ChatFilterDeleteRequest struct {
	APIKeyRequest
	// Word identifies the filter word.
	Word string `path:"word" required:"true" minLength:"1" maxLength:"32"`
}

// ChatBubbleRequest contains one bubble threshold mutation.
type ChatBubbleRequest struct {
	APIKeyRequest
	// BubbleID identifies the Nitro style.
	BubbleID int32 `path:"bubbleId" required:"true" minimum:"0"`
	// MinWeight stores the minimum primary-group weight.
	MinWeight int32 `json:"minWeight" required:"true" minimum:"0"`
}

// ChatBubbleDeleteRequest contains one bubble path identity.
type ChatBubbleDeleteRequest struct {
	APIKeyRequest
	// BubbleID identifies the Nitro style.
	BubbleID int32 `path:"bubbleId" required:"true" minimum:"0"`
}

// ChatHistoryRoomRequest contains room history filters.
type ChatHistoryRoomRequest struct {
	RoomIDRequest
	// Before stores an optional descending id cursor.
	Before int64 `query:"before,omitempty" minimum:"1"`
	// Limit caps returned rows.
	Limit int `query:"limit,omitempty" minimum:"1" maximum:"200"`
}

// ChatHistoryPlayerRequest contains player history filters.
type ChatHistoryPlayerRequest struct {
	APIKeyRequest
	// PlayerID identifies the speaker.
	PlayerID int64 `path:"playerId" required:"true" minimum:"1"`
	// RoomID optionally restricts one room.
	RoomID int64 `query:"roomId,omitempty" minimum:"1"`
	// Before stores an optional descending id cursor.
	Before int64 `query:"before,omitempty" minimum:"1"`
	// Limit caps returned rows.
	Limit int `query:"limit,omitempty" minimum:"1" maximum:"200"`
}

// ChatFilterListResponse contains global filter words.
type ChatFilterListResponse struct {
	// Total stores dictionary size.
	Total int `json:"total" required:"true"`
	// Items stores normalized words.
	Items []string `json:"items" required:"true"`
}

// ChatBubbleUnlock describes one configured threshold.
type ChatBubbleUnlock struct {
	// BubbleID identifies the Nitro style.
	BubbleID int32 `json:"bubbleId" required:"true"`
	// MinWeight stores the minimum group weight.
	MinWeight int32 `json:"minWeight" required:"true"`
}

// ChatBubbleListResponse contains configured thresholds.
type ChatBubbleListResponse struct {
	// Total stores threshold count.
	Total int `json:"total" required:"true"`
	// Items stores thresholds.
	Items []ChatBubbleUnlock `json:"items" required:"true"`
}

// ChatHistoryEntry describes one delivered message.
type ChatHistoryEntry struct {
	// ID identifies the history row.
	ID int64 `json:"id" required:"true"`
	// RoomID identifies the room.
	RoomID int64 `json:"roomId" required:"true"`
	// PlayerID identifies the speaker.
	PlayerID int64 `json:"playerId" required:"true"`
	// TargetPlayerID optionally identifies a whisper recipient.
	TargetPlayerID *int64 `json:"targetPlayerId,omitempty"`
	// Kind stores talk, shout, or whisper.
	Kind string `json:"kind" required:"true"`
	// Message stores delivered text.
	Message string `json:"message" required:"true"`
	// Censored reports whether filtering changed the text.
	Censored bool `json:"censored" required:"true"`
	// CreatedAt stores delivery time.
	CreatedAt string `json:"createdAt" required:"true" format:"date-time"`
}

// ChatHistoryResponse contains one keyset page.
type ChatHistoryResponse struct {
	// Total stores page size.
	Total int `json:"total" required:"true"`
	// Items stores history entries.
	Items []ChatHistoryEntry `json:"items" required:"true"`
}

// ChatMutationResponse confirms a mutation.
type ChatMutationResponse struct {
	// Updated reports successful completion.
	Updated bool `json:"updated" required:"true"`
}
