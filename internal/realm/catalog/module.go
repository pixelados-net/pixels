// Package catalog contains catalog realm dependency wiring.
package catalog

import (
	"context"

	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	catalogadmin "github.com/niflaot/pixels/internal/realm/catalog/admin"
	"github.com/niflaot/pixels/internal/realm/catalog/gift"
	catalogrepo "github.com/niflaot/pixels/internal/realm/catalog/repository"
	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
	catalogtrophy "github.com/niflaot/pixels/internal/realm/catalog/trophy"
	chatfilter "github.com/niflaot/pixels/internal/realm/chat/filter"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	playereffect "github.com/niflaot/pixels/internal/realm/player/effect"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	roombundle "github.com/niflaot/pixels/internal/realm/room/record/bundle"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/postgres"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides catalog persistence, cache, and purchase behavior.
var Module = fx.Module(
	"realm-catalog",
	fx.Provide(
		NewStore,
		NewService,
		NewManager,
		NewReader,
		gift.NewOptions,
		NewAdminService,
		NewAdminManager,
		NewVoucherManager,
	),
	fx.Invoke(RegisterLifecycle, RegisterConnectionHandlers),
)

// NewService creates permission-aware catalog behavior.
func NewService(store catalogrepo.Store, currencies currencyservice.Granter, furniture furnitureservice.DefinitionGranter, teleportPairs furnitureservice.TeleportPairer, events bus.Publisher, log *zap.Logger, permissions permissionservice.Checker, players playerservice.Finder, rooms roombundle.Manager, effects playereffect.Manager, filter *chatfilter.Service) *catalogservice.Service {
	return catalogservice.New(store, currencies, furniture, events, log, permissions).WithTeleportPairer(teleportPairs).WithPlayers(players).WithRoomBundles(rooms).WithEffects(effects).WithTrophies(catalogtrophy.New(filter))
}

// NewVoucherManager exposes voucher administration behavior.
func NewVoucherManager(service *catalogadmin.Service) catalogadmin.VoucherManager {
	return service
}

// NewAdminService creates catalog administration with room template validation.
func NewAdminService(store catalogrepo.Store, catalog catalogservice.Manager, definitions furnitureservice.DefinitionGranter, rooms roombundle.Manager) *catalogadmin.Service {
	return catalogadmin.New(store, catalog, definitions).WithRoomBundles(rooms)
}

// NewStore creates catalog persistence behavior.
func NewStore(pool *postgres.Pool) catalogrepo.Store {
	return catalogrepo.New(pool)
}

// NewAdminManager exposes catalog administration behavior.
func NewAdminManager(service *catalogadmin.Service) catalogadmin.Manager {
	return service
}

// NewManager exposes catalog management behavior.
func NewManager(service *catalogservice.Service) catalogservice.Manager {
	return service
}

// NewReader exposes catalog read behavior.
func NewReader(service *catalogservice.Service) catalogservice.Reader {
	return service
}

// RegisterLifecycle loads catalog data when the application starts.
func RegisterLifecycle(lifecycle fx.Lifecycle, service *catalogservice.Service) {
	lifecycle.Append(fx.Hook{OnStart: func(ctx context.Context) error {
		return service.Refresh(ctx)
	}})
}
