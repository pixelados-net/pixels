// Package inventory contains player inventory realm wiring.
package inventory

import (
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	"github.com/niflaot/pixels/internal/realm/inventory/currency"
	currencybroadcast "github.com/niflaot/pixels/internal/realm/inventory/currency/broadcast"
	requestcmd "github.com/niflaot/pixels/internal/realm/inventory/currency/commands/request"
	currencyrepo "github.com/niflaot/pixels/internal/realm/inventory/currency/repository"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/postgres"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides inventory currency persistence and packet behavior.
var Module = fx.Module(
	"realm-inventory",
	fx.Provide(
		currency.LoadCatalog,
		NewCurrencyStore,
		NewCurrencyService,
		currencybroadcast.New,
		NewCurrencyManager,
		NewCurrencyGranter,
		NewCurrencyReader,
		NewCurrencyRequest,
	),
	fx.Invoke(
		RegisterConnectionHandlers,
		RegisterCurrencyBroadcaster,
	),
)

// NewCurrencyService creates currency behavior with permission-aware deductions.
func NewCurrencyService(store currencyrepo.Store, catalog *currency.Catalog, events bus.Publisher, log *zap.Logger, permissions permissionservice.Checker) *currencyservice.Service {
	return currencyservice.New(store, catalog, events, log, permissions)
}

// NewCurrencyStore creates the currency persistence store.
func NewCurrencyStore(pool *postgres.Pool) currencyrepo.Store {
	return currencyrepo.New(pool)
}

// NewCurrencyManager exposes currency management behavior.
func NewCurrencyManager(service *currencyservice.Service) currencyservice.Manager {
	return service
}

// NewCurrencyGranter exposes signed currency mutation behavior.
func NewCurrencyGranter(service *currencyservice.Service) currencyservice.Granter {
	return service
}

// NewCurrencyReader exposes currency read behavior.
func NewCurrencyReader(service *currencyservice.Service) currencyservice.Reader {
	return service
}

// NewCurrencyRequest creates the currency wallet command handler.
func NewCurrencyRequest(players *playerlive.Registry, bindings *binding.Registry, currencies currencyservice.Manager) *requestcmd.Handler {
	return &requestcmd.Handler{Players: players, Bindings: bindings, Currencies: currencies}
}
