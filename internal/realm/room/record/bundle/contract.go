// Package bundle clones and administers complete catalog room templates.
package bundle

import (
	"context"

	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
)

// Manager administers templates and clones purchased rooms.
type Manager interface {
	// Clone clones a marked template for one buyer.
	Clone(ctx context.Context, params CloneParams) (CloneResult, error)
	// Preview groups furniture in a marked template.
	Preview(ctx context.Context, templateRoomID int64) ([]Product, error)
	// Mark marks a room as a bundle source.
	Mark(ctx context.Context, roomID int64) (roommodel.Room, error)
	// Unmark removes bundle-source status when unreferenced.
	Unmark(ctx context.Context, roomID int64) (roommodel.Room, error)
	// Templates lists marked active templates.
	Templates(ctx context.Context) ([]roommodel.Room, error)
	// FindTemplate validates and returns one marked template.
	FindTemplate(ctx context.Context, roomID int64) (roommodel.Room, bool, error)
}

// Store persists atomic room bundle operations.
type Store interface {
	// WithinTransaction runs bundle work atomically.
	WithinTransaction(ctx context.Context, work func(context.Context) error) error
	// LockRoomOwner serializes room-limit checks for one owner.
	LockRoomOwner(ctx context.Context, ownerPlayerID int64) error
	// CountRoomsByOwner counts non-template rooms owned by a player.
	CountRoomsByOwner(ctx context.Context, ownerPlayerID int64) (int, error)
	// CloneBundleRoom copies a template room for a buyer.
	CloneBundleRoom(ctx context.Context, templateRoomID int64, buyerPlayerID int64, buyerName string) (roommodel.Room, error)
	// RecordBundlePurchase records bundle provenance.
	RecordBundlePurchase(ctx context.Context, params PurchaseRecord) error
	// SetBundleTemplate changes template status.
	SetBundleTemplate(ctx context.Context, roomID int64, enabled bool) (roommodel.Room, bool, error)
	// CountActiveBundleReferences counts enabled offers referencing a room.
	CountActiveBundleReferences(ctx context.Context, roomID int64) (int, error)
	// ListBundleTemplateRooms lists active template rooms.
	ListBundleTemplateRooms(ctx context.Context) ([]roommodel.Room, error)
}

// BotCloner copies placed bots inside the bundle transaction.
type BotCloner interface {
	// CloneRoom copies every placed bot and its chat into a new owner room.
	CloneRoom(context.Context, int64, int64, int64) (int, error)
}

// CloneParams contains a purchased room clone request.
type CloneParams struct {
	// TemplateRoomID identifies the marked source room.
	TemplateRoomID int64
	// BuyerPlayerID identifies the new room owner.
	BuyerPlayerID int64
	// BuyerName stores the owner-name snapshot.
	BuyerName string
	// CatalogItemID identifies the purchased offer.
	CatalogItemID int64
}

// CloneResult contains a completed room clone.
type CloneResult struct {
	// Room stores the created room.
	Room roommodel.Room
	// FurnitureCount stores the number of cloned furniture rows.
	FurnitureCount int
	// BotCount stores the number of cloned bot rows.
	BotCount int
}

// Product contains one grouped furniture definition.
type Product struct {
	// DefinitionID identifies the furniture definition.
	DefinitionID int64
	// Quantity stores matching template items.
	Quantity int32
}

// PurchaseRecord contains durable bundle provenance.
type PurchaseRecord struct {
	// CatalogItemID identifies the purchased offer.
	CatalogItemID int64
	// TemplateRoomID identifies the source room.
	TemplateRoomID int64
	// CreatedRoomID identifies the cloned room.
	CreatedRoomID int64
	// BuyerPlayerID identifies the buyer.
	BuyerPlayerID int64
	// FurnitureCount stores cloned furniture count.
	FurnitureCount int
	// BotCount stores cloned bot count.
	BotCount int
}
