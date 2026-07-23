package record

import (
	"context"
	"time"
)

// Store persists bots, serving mappings, and visit history.
type Store interface {
	// WithinTransaction runs related bot mutations atomically.
	WithinTransaction(context.Context, func(context.Context) error) error
	// Find returns one bot with ordered chat lines.
	Find(context.Context, int64) (Bot, bool, error)
	// Inventory lists bots currently held by one player.
	Inventory(context.Context, int64) ([]Bot, error)
	// Room lists bots currently placed in one room.
	Room(context.Context, int64) ([]Bot, error)
	// CountInventory counts bots held by one player.
	CountInventory(context.Context, int64) (int, error)
	// Place moves an owned inventory bot into a room.
	Place(context.Context, int64, int64, int64, int, int, float64, int16) (Bot, bool, error)
	// Pickup moves a placed bot to a receiving owner.
	Pickup(context.Context, int64, int64, int64) (Bot, bool, error)
	// ForcePickup moves a placed bot back to its current owner.
	ForcePickup(context.Context, int64) (Bot, bool, error)
	// Delete permanently removes one owned inventory bot.
	Delete(context.Context, int64, int64) (bool, error)
	// Save replaces mutable settings and ordered chat lines.
	Save(context.Context, Bot) (Bot, bool, error)
	// SavePosition persists a placed bot's latest world position.
	SavePosition(context.Context, int64, int64, int, int, float64, int16) error
	// CloneRoom copies every placed bot into another room and owner.
	CloneRoom(context.Context, int64, int64, int64) (int, error)
	// ListServeItems returns every keyword mapping.
	ListServeItems(context.Context) ([]ServeItem, error)
	// CreateServeItem inserts a keyword mapping.
	CreateServeItem(context.Context, string, int64) (ServeItem, error)
	// UpdateServeItem changes a keyword mapping.
	UpdateServeItem(context.Context, int64, string, int64) (ServeItem, bool, error)
	// DeleteServeItem removes a keyword mapping.
	DeleteServeItem(context.Context, int64) (bool, error)
	// RecordVisit appends one room entry.
	RecordVisit(context.Context, int64, int64) error
	// VisitsSince returns recent visits in chronological order.
	VisitsSince(context.Context, int64, int64, time.Time, int) ([]Visit, error)
}
