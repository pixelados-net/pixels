// Package service contains catalog browsing and purchase behavior.
package service

import (
	"context"
	"time"

	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
)

// Reader reads player-visible catalog data.
type Reader interface {
	// Pages returns pages visible to one player capability set.
	Pages(ctx context.Context, playerID int64, hasClub bool) ([]catalogmodel.Page, error)

	// Page returns one visible page and its enabled offers.
	Page(ctx context.Context, pageID int64, playerID int64, hasClub bool) (catalogmodel.Page, []catalogmodel.Item, error)

	// Definition returns cached furniture metadata for one catalog offer.
	Definition(ctx context.Context, definitionID int64) (furnituremodel.Definition, bool, error)

	// SanitizeList returns definitions without an enabled active offer.
	SanitizeList(ctx context.Context) ([]furnituremodel.Definition, error)
}

// Manager reads catalog data and processes purchases.
type Manager interface {
	Reader

	// Purchase buys one catalog offer.
	Purchase(ctx context.Context, params PurchaseParams) (PurchaseResult, error)

	// Refresh reloads the complete catalog cache.
	Refresh(ctx context.Context) error
}

// TransactionalPurchaser buys inside an existing transaction and defers projections until commit.
type TransactionalPurchaser interface {
	// PurchaseWithin mutates the active transaction and returns post-commit work.
	PurchaseWithin(ctx context.Context, params PurchaseParams) (PurchaseResult, func(context.Context), error)
}

// AtomicPurchaser combines one catalog purchase with a caller-owned mutation.
type AtomicPurchaser interface {
	// PurchaseAndMutate commits the purchase only when mutate succeeds.
	PurchaseAndMutate(ctx context.Context, params PurchaseParams, mutate func(context.Context, PurchaseResult) error) (PurchaseResult, error)
}

// GiftManager purchases wrapped catalog offers for another player.
type GiftManager interface {
	// PurchaseGift buys one wrapped offer for another player.
	PurchaseGift(ctx context.Context, params GiftPurchaseParams) (PurchaseResult, error)
}

// VoucherManager redeems one-time catalog voucher rewards.
type VoucherManager interface {
	// RedeemVoucher redeems one voucher for a player.
	RedeemVoucher(ctx context.Context, playerID int64, code string) (RedeemResult, error)
}

// BundleReader reads the products that compose catalog offers.
type BundleReader interface {
	// Products returns cached bundle products for one offer.
	Products(ctx context.Context, catalogItemID int64) []catalogmodel.Product
}

// TrophyFormatter creates immutable protocol-compatible trophy inscriptions.
type TrophyFormatter interface {
	// Format composes one buyer name and requested inscription.
	Format(username string, text string) string
}

// NoveltyManager manages per-player catalog freshness state.
type NoveltyManager interface {
	// MarkNewAdditionsSeen records catalog novelty acknowledgement.
	MarkNewAdditionsSeen(ctx context.Context, playerID int64) error
	// NewAdditionsAvailable reports whether one player has unseen novelty offers.
	NewAdditionsAvailable(ctx context.Context, playerID int64) (bool, error)
}

// SpendingReader reads committed catalog spending for rewards.
type SpendingReader interface {
	// CreditsSpentSince sums kickback-eligible catalog spending.
	CreditsSpentSince(ctx context.Context, playerID int64, since time.Time) (int64, error)
	// CreditsSpentBetween sums eligible spending inside one payday period.
	CreditsSpentBetween(ctx context.Context, playerID int64, after time.Time, through time.Time) (int64, error)
}

// PetGrantParams contains one trusted transactional pet catalog grant.
type PetGrantParams struct {
	// OwnerPlayerID identifies the inventory owner.
	OwnerPlayerID int64
	// TypeID identifies the server-selected species.
	TypeID int32
	// ProductCode identifies Nitro's pet localization product.
	ProductCode string
	// ExtraData stores the client-selected name, palette, and color.
	ExtraData string
	// CatalogItemID identifies the originating catalog offer.
	CatalogItemID int64
	// OperationKey makes one requested grant idempotent.
	OperationKey string
}

// PetReward identifies one pet created by a committed catalog purchase.
type PetReward struct {
	// ID identifies the granted pet.
	ID int64
	// OwnerPlayerID identifies the receiving inventory owner.
	OwnerPlayerID int64
}

// PetGranter creates and projects typed pet rewards.
type PetGranter interface {
	// GrantCatalog creates one pet inside the caller transaction.
	GrantCatalog(ctx context.Context, params PetGrantParams) (PetReward, error)
	// ProjectCatalog sends post-commit inventory projections.
	ProjectCatalog(ctx context.Context, reward PetReward)
}

// GroupCommerce validates and commits social-group catalog rewards.
type GroupCommerce interface {
	// ValidateCatalog validates membership, role, and entitlement state.
	ValidateCatalog(ctx context.Context, playerID int64, groupID int64, forum bool) error
	// CommitCatalog links granted items and activates an optional forum.
	CommitCatalog(ctx context.Context, playerID int64, groupID int64, forum bool, itemIDs []int64) error
	// ProjectCatalog refreshes committed group generations.
	ProjectCatalog(ctx context.Context, playerID int64, groupID int64, itemIDs []int64)
}

// PurchaseParams contains one catalog purchase request.
type PurchaseParams struct {
	// PlayerID identifies the buyer.
	PlayerID int64

	// CatalogItemID identifies the requested offer.
	CatalogItemID int64

	// HasClub reports whether the buyer has active club membership.
	HasClub bool

	// Amount stores the requested offer quantity.
	Amount int32

	// ExtraData stores client-supplied product data used only by supported layouts.
	ExtraData string

	// OperationKey optionally makes a system-originated purchase grant idempotent.
	OperationKey string

	// RecipientPlayerID optionally overrides the furniture recipient.
	RecipientPlayerID int64

	// Free bypasses the configured catalog price for system rewards.
	Free bool

	// Gift optionally stores wrapping metadata.
	Gift *GiftMetadata

	// OverrideCredits optionally replaces the offer's credits price.
	OverrideCredits *int64

	// OverridePoints optionally replaces the offer's points price.
	OverridePoints *int64

	// OverridePointsType optionally replaces the offer's points currency.
	OverridePointsType *int32
}

// GiftMetadata contains wrapped purchase metadata.
type GiftMetadata struct {
	// SpriteID identifies the selected wrapping furniture sprite.
	SpriteID int32
	// BoxID identifies the wrapping box.
	BoxID int32
	// RibbonID identifies the wrapping ribbon.
	RibbonID int32
	// SenderPlayerID identifies the visible sender.
	SenderPlayerID *int64
	// Message stores the gift message.
	Message string
}

// GiftPurchaseParams contains a catalog gift request.
type GiftPurchaseParams struct {
	// BuyerID identifies the paying player.
	BuyerID int64
	// ReceiverName identifies the recipient.
	ReceiverName string
	// CatalogItemID identifies the offer.
	CatalogItemID int64
	// HasClub reports the buyer's club entitlement.
	HasClub bool
	// SpriteID identifies the selected wrapping furniture sprite.
	SpriteID int32
	// BoxID identifies the wrapping box.
	BoxID int32
	// RibbonID identifies the ribbon.
	RibbonID int32
	// Message stores the gift message.
	Message string
	// ExtraData stores client-supplied product data used only by supported layouts.
	ExtraData string
	// ShowMyFace reports whether sender identity is visible.
	ShowMyFace bool
}

// RedeemResult contains one completed voucher redemption.
type RedeemResult struct {
	// ProductCode stores the client-facing reward code.
	ProductCode string
	// GrantedItems stores furniture granted by the voucher.
	GrantedItems []furnituremodel.Item
}

// PurchaseResult contains one completed purchase.
type PurchaseResult struct {
	// Item stores the purchased offer snapshot.
	Item catalogmodel.Item
	// RecipientPlayerID identifies the inventory owner receiving the purchase.
	RecipientPlayerID int64

	// GrantedItems stores created furniture instances.
	GrantedItems []furnituremodel.Item

	// GrantedEffectID identifies the effect reward when present.
	GrantedEffectID *int32

	// GrantedPet identifies a typed pet reward when present.
	GrantedPet *PetReward

	// Products stores the offer products resolved before the purchase commits.
	Products []catalogmodel.Product

	// CreatedRoomID identifies a room created by a room bundle offer.
	CreatedRoomID *int64

	// CreatedRoomName stores the visible name of a created room.
	CreatedRoomName string

	// ClonedFurnitureCount stores furniture copied into a created room.
	ClonedFurnitureCount int

	// ClonedBotCount stores bots copied into a created room.
	ClonedBotCount int

	// LimitedUnitNumber stores the optional LTD edition number.
	LimitedUnitNumber *int32

	// NewCreditsBalance stores the resulting credits balance.
	NewCreditsBalance int64

	// NewPointsBalance stores the resulting activity-points balance.
	NewPointsBalance int64

	// ChargedCredits stores the committed credits charge.
	ChargedCredits int64

	// ChargedPoints stores the committed activity-points charge.
	ChargedPoints int64
}

// managerAssertion verifies Service implements Manager.
var managerAssertion Manager = (*Service)(nil)
