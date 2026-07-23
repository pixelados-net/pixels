package subscription

import (
	"context"
	"errors"

	buycmd "github.com/niflaot/pixels/internal/realm/catalog/commands/buy"
	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	subadmin "github.com/niflaot/pixels/internal/realm/subscription/admin"
	"github.com/niflaot/pixels/internal/realm/subscription/core"
	"github.com/niflaot/pixels/internal/realm/subscription/database"
	"github.com/niflaot/pixels/internal/realm/subscription/record"
	subruntime "github.com/niflaot/pixels/internal/realm/subscription/runtime"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/postgres"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides subscription persistence, behavior, and scheduling.
var Module = fx.Module(
	"realm-subscription",
	fx.Provide(NewStore, NewAdminStore, NewService, NewAdminService, NewScheduler, NewCatalogPurchaser),
	fx.Invoke(subruntime.RegisterScheduler, subruntime.RegisterPaydayClaims, subruntime.RegisterBootstrap, RegisterConnectionHandlers),
)

// catalogPurchaser adapts subscription behavior to catalog purchases.
type catalogPurchaser struct {
	// service manages subscription purchases.
	service *core.Service
}

// NewCatalogPurchaser creates the catalog subscription boundary.
func NewCatalogPurchaser(service *core.Service) buycmd.ClubPurchaser {
	return catalogPurchaser{service: service}
}

// PurchaseClub buys one subscription offer for one player.
func (purchaser catalogPurchaser) PurchaseClub(ctx context.Context, playerID int64, offerID int64, amount int32) (buycmd.ClubPurchase, error) {
	membership, err := purchaser.service.PurchaseOfferAmount(ctx, playerID, playerID, offerID, amount)
	if errors.Is(err, core.ErrOfferNotFound) {
		return buycmd.ClubPurchase{}, catalogservice.ErrOfferNotFound
	}
	if errors.Is(err, core.ErrInvalidAmount) {
		return buycmd.ClubPurchase{}, catalogservice.ErrInvalidAmount
	}
	if err != nil {
		return buycmd.ClubPurchase{}, err
	}
	if membership.ExpiresAt == nil {
		return buycmd.ClubPurchase{}, catalogservice.ErrOfferNotFound
	}

	return buycmd.ClubPurchase{ExpiresAt: *membership.ExpiresAt,
		LifetimeActiveSeconds: membership.LifetimeActiveSeconds,
		LifetimeVIPSeconds:    membership.LifetimeVIPSeconds, VIP: membership.Level == record.LevelVIP}, nil
}

// NewStore creates subscription persistence behavior.
func NewStore(pool *postgres.Pool) record.Store {
	return database.New(pool)
}

// NewAdminStore exposes subscription administration persistence.
func NewAdminStore(store record.Store) subadmin.Store {
	return store.(subadmin.Store)
}

// NewAdminService creates subscription administration behavior.
func NewAdminService(store subadmin.Store, service *core.Service) *subadmin.Service {
	return subadmin.New(store, service)
}

// NewService creates configured subscription behavior.
func NewService(config Config, store record.Store, players playerservice.ClubWriter, livePlayers *playerlive.Registry, currencies currencyservice.Manager, furniture furnitureservice.DefinitionGranter, catalog *catalogservice.Service, events bus.Publisher) *core.Service {
	config = config.Normalize()
	return core.New(core.Options{PaydayInterval: config.PaydayInterval, KickbackPercentage: config.KickbackPercentage,
		PaydayCurrencyType: config.PaydayCurrencyType, BonusRareCurrencyType: config.BonusRareCurrencyType,
		BonusRareThreshold: config.BonusRareThreshold, BonusRareProductID: config.BonusRareProductID},
		store, players, currencies, furniture, catalog, events, livePlayers)
}

// NewScheduler creates the global subscription lifecycle scheduler.
func NewScheduler(config Config, service *core.Service, log *zap.Logger) *subruntime.Scheduler {
	return subruntime.NewScheduler(config.Normalize().TickInterval, service, log)
}
