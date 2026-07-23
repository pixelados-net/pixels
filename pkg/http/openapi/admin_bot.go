package openapi

import (
	"net/http"
	"time"
)

// BotIDRequest documents one bot path identifier.
type BotIDRequest struct {
	APIKeyRequest
	// ID identifies the bot or serve mapping.
	ID int64 `path:"id" required:"true" minimum:"1"`
}

// BotServeItemRequest documents one bartender keyword mapping.
type BotServeItemRequest struct {
	APIKeyRequest
	// Keyword stores the whole-word trigger.
	Keyword string `json:"keyword" required:"true" minLength:"1" maxLength:"32" example:"coffee"`
	// DefinitionID identifies the delivered hand item.
	DefinitionID int64 `json:"definitionId" required:"true" minimum:"1"`
}

// BotServeItemPatchRequest documents one bartender mapping replacement.
type BotServeItemPatchRequest struct {
	BotIDRequest
	// Keyword stores the whole-word trigger.
	Keyword string `json:"keyword" required:"true" minLength:"1" maxLength:"32"`
	// DefinitionID identifies the delivered hand item.
	DefinitionID int64 `json:"definitionId" required:"true" minimum:"1"`
}

// BotServeItemResponse documents one bartender keyword mapping.
type BotServeItemResponse struct {
	// ID identifies the mapping.
	ID int64 `json:"id" required:"true"`
	// Keyword stores the normalized trigger.
	Keyword string `json:"keyword" required:"true"`
	// DefinitionID identifies the delivered item.
	DefinitionID int64 `json:"definitionId" required:"true"`
}

// BotServeItemListResponse documents all bartender mappings.
type BotServeItemListResponse struct {
	// Items stores ordered mappings.
	Items []BotServeItemResponse `json:"items" required:"true"`
	// Count stores the result size.
	Count int `json:"count" required:"true"`
}

// BotAdminResponse documents complete bot support state.
type BotAdminResponse struct {
	// ID identifies the bot.
	ID int64 `json:"id" required:"true"`
	// OwnerPlayerID identifies the owner.
	OwnerPlayerID int64 `json:"ownerPlayerId" required:"true"`
	// OwnerName stores the owner name.
	OwnerName string `json:"ownerName" required:"true"`
	// RoomID identifies current placement.
	RoomID *int64 `json:"roomId,omitempty"`
	// BehaviorType identifies the behavior plugin.
	BehaviorType string `json:"behaviorType" required:"true"`
	// Name stores the visible name.
	Name string `json:"name" required:"true"`
	// Motto stores the visible motto.
	Motto string `json:"motto" required:"true"`
	// Figure stores the Nitro figure.
	Figure string `json:"figure" required:"true"`
	// Gender stores the Nitro gender code.
	Gender string `json:"gender" required:"true"`
	// X stores the optional placed x coordinate.
	X *int `json:"x,omitempty"`
	// Y stores the optional placed y coordinate.
	Y *int `json:"y,omitempty"`
	// Z stores the optional placed height.
	Z *float64 `json:"z,omitempty"`
	// Rotation stores the optional placed body rotation.
	Rotation *int16 `json:"rotation,omitempty"`
	// ChatLines stores ordered automatic speech.
	ChatLines []string `json:"chatLines" required:"true"`
	// CanWalk reports random movement policy.
	CanWalk bool `json:"canWalk" required:"true"`
	// DanceType stores the persistent dance selection.
	DanceType int16 `json:"danceType" required:"true"`
	// ChatAuto reports whether automatic speech is enabled.
	ChatAuto bool `json:"chatAuto" required:"true"`
	// ChatRandom reports whether automatic lines are selected randomly.
	ChatRandom bool `json:"chatRandom" required:"true"`
	// ChatDelaySeconds stores automatic speech cadence.
	ChatDelaySeconds int `json:"chatDelaySeconds" required:"true"`
	// BubbleStyle stores the bot chat bubble style.
	BubbleStyle int32 `json:"bubbleStyle" required:"true"`
	// EffectID stores the optional persistent avatar effect.
	EffectID *int32 `json:"effectId,omitempty"`
	// CreatedAt stores creation time.
	CreatedAt time.Time `json:"createdAt" required:"true"`
	// UpdatedAt stores mutation time.
	UpdatedAt time.Time `json:"updatedAt" required:"true"`
	// Version stores optimistic state.
	Version int64 `json:"version" required:"true"`
}

// adminBotOperations returns protected bot administration operations.
func adminBotOperations() []operation {
	return []operation{
		adminBot(http.MethodGet, "/api/admin/bots/serve-items", "List bot serve items", &APIKeyRequest{}, &BotServeItemListResponse{}, http.StatusOK),
		adminBot(http.MethodPost, "/api/admin/bots/serve-items", "Create bot serve item", &BotServeItemRequest{}, &BotServeItemResponse{}, http.StatusCreated),
		adminBot(http.MethodPatch, "/api/admin/bots/serve-items/{id}", "Update bot serve item", &BotServeItemPatchRequest{}, &BotServeItemResponse{}, http.StatusOK),
		adminBot(http.MethodDelete, "/api/admin/bots/serve-items/{id}", "Delete bot serve item", &BotIDRequest{}, nil, http.StatusNoContent),
		adminBot(http.MethodGet, "/api/admin/bots/{id}", "Read bot support state", &BotIDRequest{}, &BotAdminResponse{}, http.StatusOK),
		adminBot(http.MethodPost, "/api/admin/bots/{id}/force-pickup", "Force pickup bot", &BotIDRequest{}, &BotAdminResponse{}, http.StatusOK),
	}
}

// adminBot creates one protected bot operation.
func adminBot(method string, path string, summary string, request any, body any, status int) operation {
	responses := errorResponses(http.StatusBadRequest, http.StatusUnauthorized, http.StatusNotFound, http.StatusConflict, http.StatusInternalServerError)
	if body == nil {
		responses = append([]response{emptyResponse(status, summary+".")}, responses...)
	} else {
		responses = append([]response{jsonResponse(status, body, summary+".")}, responses...)
	}
	return operation{method: method, path: path, tag: "Bots", summary: summary, description: summary + ".", request: request, responses: responses, secured: true}
}
