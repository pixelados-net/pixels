// Package record defines durable bot records and persistence contracts.
package record

import "time"

const (
	// BehaviorGeneric identifies the standard decorative bot.
	BehaviorGeneric = "generic"
	// BehaviorBartender identifies the keyword-driven serving bot.
	BehaviorBartender = "bartender"
	// BehaviorVisitorLog identifies the room visitor logger.
	BehaviorVisitorLog = "visitor_log"
)

// Bot stores one durable owner-configurable room bot.
type Bot struct {
	// ID identifies the bot.
	ID int64
	// OwnerPlayerID identifies the current inventory owner.
	OwnerPlayerID int64
	// OwnerName stores the visible owner snapshot loaded with the bot.
	OwnerName string
	// RoomID identifies placement and is nil while in inventory.
	RoomID *int64
	// BehaviorType selects a registered behavior.
	BehaviorType string
	// Name stores the visible bot name.
	Name string
	// Motto stores the visible bot motto.
	Motto string
	// Figure stores the Nitro avatar figure.
	Figure string
	// Gender stores the Nitro gender code.
	Gender string
	// X stores the optional placed tile coordinate.
	X *int
	// Y stores the optional placed tile coordinate.
	Y *int
	// Z stores the optional placed height.
	Z *float64
	// Rotation stores the optional placed body rotation.
	Rotation *int16
	// CanWalk controls generic random walking.
	CanWalk bool
	// DanceType stores the persistent dance selection.
	DanceType int16
	// ChatAuto controls scheduled chat.
	ChatAuto bool
	// ChatRandom controls random rather than sequential line selection.
	ChatRandom bool
	// ChatDelaySeconds stores the scheduled chat delay.
	ChatDelaySeconds int
	// BubbleStyle stores the chat bubble style.
	BubbleStyle int32
	// EffectID stores an optional avatar effect.
	EffectID *int32
	// ChatLines stores ordered configured chat lines.
	ChatLines []string
	// CreatedAt stores record creation time.
	CreatedAt time.Time
	// UpdatedAt stores last durable mutation time.
	UpdatedAt time.Time
	// Version stores optimistic mutation state.
	Version int64
}

// Inventory reports whether the bot is currently owned outside a room.
func (bot Bot) Inventory() bool {
	return bot.RoomID == nil
}

// ServeItem maps a whole-word keyword to a protocol hand item definition.
type ServeItem struct {
	// ID identifies the mapping.
	ID int64
	// Keyword stores the normalized whole-word trigger.
	Keyword string
	// DefinitionID stores the delivered protocol hand item id.
	DefinitionID int64
}

// Visit stores one player room-entry observation.
type Visit struct {
	// RoomID identifies the visited room.
	RoomID int64
	// PlayerID identifies the visitor.
	PlayerID int64
	// PlayerName stores the visitor name resolved by persistence.
	PlayerName string
	// EnteredAt stores entry time.
	EnteredAt time.Time
}
