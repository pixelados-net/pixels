// Package routes exposes protected bot administration.
package routes

import "time"

// ServeItemRequest creates or replaces one bartender keyword mapping.
type ServeItemRequest struct {
	// Keyword stores the whole-word trigger.
	Keyword string `json:"keyword"`
	// DefinitionID identifies the delivered hand item.
	DefinitionID int64 `json:"definitionId"`
}

// ServeItemResponse describes one bartender keyword mapping.
type ServeItemResponse struct {
	// ID identifies the mapping.
	ID int64 `json:"id"`
	// Keyword stores the whole-word trigger.
	Keyword string `json:"keyword"`
	// DefinitionID identifies the delivered hand item.
	DefinitionID int64 `json:"definitionId"`
}

// ServeItemListResponse contains all bartender keyword mappings.
type ServeItemListResponse struct {
	// Items stores ordered mappings.
	Items []ServeItemResponse `json:"items"`
	// Count stores the result size.
	Count int `json:"count"`
}

// BotResponse describes bot support state.
type BotResponse struct {
	// ID identifies the bot.
	ID int64 `json:"id"`
	// OwnerPlayerID identifies its owner.
	OwnerPlayerID int64 `json:"ownerPlayerId"`
	// OwnerName stores the visible owner name.
	OwnerName string `json:"ownerName"`
	// RoomID identifies placement and is nil in inventory.
	RoomID *int64 `json:"roomId"`
	// BehaviorType identifies the behavior plugin.
	BehaviorType string `json:"behaviorType"`
	// Name stores the visible name.
	Name string `json:"name"`
	// Motto stores the visible motto.
	Motto string `json:"motto"`
	// Figure stores the Nitro figure.
	Figure string `json:"figure"`
	// Gender stores the Nitro gender value.
	Gender string `json:"gender"`
	// X stores the placed x coordinate.
	X *int `json:"x"`
	// Y stores the placed y coordinate.
	Y *int `json:"y"`
	// Z stores the placed height.
	Z *float64 `json:"z"`
	// Rotation stores body rotation.
	Rotation *int16 `json:"rotation"`
	// CanWalk reports whether random movement is active.
	CanWalk bool `json:"canWalk"`
	// DanceType stores the selected dance.
	DanceType int16 `json:"danceType"`
	// ChatAuto reports whether automatic chat is active.
	ChatAuto bool `json:"chatAuto"`
	// ChatRandom reports whether chat order is random.
	ChatRandom bool `json:"chatRandom"`
	// ChatDelaySeconds stores automatic chat cadence.
	ChatDelaySeconds int `json:"chatDelaySeconds"`
	// BubbleStyle stores the selected bubble.
	BubbleStyle int32 `json:"bubbleStyle"`
	// EffectID stores an optional avatar effect.
	EffectID *int32 `json:"effectId"`
	// ChatLines stores ordered configured text.
	ChatLines []string `json:"chatLines"`
	// CreatedAt stores creation time.
	CreatedAt time.Time `json:"createdAt"`
	// UpdatedAt stores last mutation time.
	UpdatedAt time.Time `json:"updatedAt"`
	// Version stores optimistic state.
	Version int64 `json:"version"`
}
