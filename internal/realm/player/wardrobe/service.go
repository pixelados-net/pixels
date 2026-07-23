package wardrobe

import (
	"context"
	"strings"
	"time"

	playerfigure "github.com/niflaot/pixels/internal/realm/player/figure"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
)

// Service validates and persists player wardrobe outfits.
type Service struct {
	// store persists wardrobe state.
	store Store
	// config contains bounded slot policy.
	config Config
	// figures stores immutable figure-data rules.
	figures *playerfigure.Catalog
	// players reads current club entitlement without a database query.
	players *playerlive.Registry
}

// New creates wardrobe behavior.
func New(store Store) *Service { return NewConfigured(store, nil, nil, DefaultConfig()) }

// NewConfigured creates wardrobe behavior with explicit slot policy.
func NewConfigured(store Store, figures *playerfigure.Catalog, players *playerlive.Registry, config Config) *Service {
	if config.MinimumSlot < MinSlot || config.MaximumSlot < config.MinimumSlot {
		config = DefaultConfig()
	}
	return &Service{store: store, config: config, figures: figures, players: players}
}

// Outfits returns all saved outfits for one player.
func (service *Service) Outfits(ctx context.Context, playerID int64) ([]Outfit, error) {
	return service.store.Outfits(ctx, playerID)
}

// Save validates and persists one outfit.
func (service *Service) Save(ctx context.Context, playerID int64, outfit Outfit) error {
	outfit.Figure = strings.TrimSpace(outfit.Figure)
	outfit.Gender = strings.ToUpper(strings.TrimSpace(outfit.Gender))
	gender := playermodel.Gender(outfit.Gender)
	if playerID <= 0 || outfit.SlotID < service.config.MinimumSlot || outfit.SlotID > service.config.MaximumSlot || !playerfigure.Valid(outfit.Figure) || !gender.Valid() {
		return ErrInvalidOutfit
	}
	if service.figures != nil {
		clothing, err := service.store.Clothing(ctx, playerID)
		if err != nil {
			return err
		}
		club := playermodel.ClubLevelNone
		if service.players != nil {
			if player, found := service.players.Find(playerID); found {
				club = player.Snapshot().ClubLevelAt(time.Now())
			}
		}
		if !service.figures.Allowed(outfit.Figure, gender, club, clothing.FigureSetIDs) {
			return ErrInvalidOutfit
		}
	}
	return service.store.SaveOutfit(ctx, playerID, outfit)
}

// Clothing returns one player's complete unlock snapshot.
func (service *Service) Clothing(ctx context.Context, playerID int64) (ClothingSnapshot, error) {
	if playerID <= 0 {
		return ClothingSnapshot{}, ErrInvalidClothingItem
	}
	return service.store.Clothing(ctx, playerID)
}

// Redeem atomically consumes one clothing furniture and returns current unlocks.
func (service *Service) Redeem(ctx context.Context, playerID int64, itemID int64) (RedeemResult, error) {
	if playerID <= 0 || itemID <= 0 {
		return RedeemResult{}, ErrInvalidClothingItem
	}
	return service.store.RedeemClothing(ctx, playerID, itemID)
}
