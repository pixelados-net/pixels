package subscription

import (
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
	realmconn "github.com/niflaot/pixels/internal/realm/connection"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	calendarcmd "github.com/niflaot/pixels/internal/realm/subscription/commands/calendar"
	clubcmd "github.com/niflaot/pixels/internal/realm/subscription/commands/club"
	hotelviewcmd "github.com/niflaot/pixels/internal/realm/subscription/commands/hotelview"
	targetedcmd "github.com/niflaot/pixels/internal/realm/subscription/commands/targeted"
	"github.com/niflaot/pixels/internal/realm/subscription/core"
	calendarhandler "github.com/niflaot/pixels/internal/realm/subscription/handlers/calendar"
	clubhandler "github.com/niflaot/pixels/internal/realm/subscription/handlers/club"
	hotelviewhandler "github.com/niflaot/pixels/internal/realm/subscription/handlers/hotelview"
	targetedhandler "github.com/niflaot/pixels/internal/realm/subscription/handlers/targeted"
	"github.com/niflaot/pixels/pkg/i18n"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// HandlerDeps contains subscription packet handler dependencies.
type HandlerDeps struct {
	fx.In

	// Players stores live player compositions.
	Players *playerlive.Registry
	// Bindings maps connections to authenticated players.
	Bindings *binding.Registry
	// Subscriptions manages subscription behavior.
	Subscriptions *core.Service
	// Catalog reads catalog-backed offers.
	Catalog *catalogservice.Service
	// Furniture reads calendar reward definitions.
	Furniture furnitureservice.DefinitionGranter
	// Permissions resolves calendar staff bypasses.
	Permissions permissionservice.Checker
	// Translations localizes targeted offer copy.
	Translations i18n.Translator
	// Log records command dispatch failures.
	Log *zap.Logger
}

// RegisterConnectionHandlers registers subscription packet handlers.
func RegisterConnectionHandlers(handlers *realmconn.Handlers, dependencies HandlerDeps) {
	if handlers == nil || handlers.Inbound == nil {
		return
	}
	clubhandler.Register(handlers.Inbound, clubhandler.New(clubcmd.Handler{Players: dependencies.Players,
		Bindings: dependencies.Bindings, Subscriptions: dependencies.Subscriptions, Catalog: dependencies.Catalog}, dependencies.Log))
	targetedhandler.Register(handlers.Inbound, targetedhandler.New(targetedcmd.Handler{Players: dependencies.Players,
		Bindings: dependencies.Bindings, Subscriptions: dependencies.Subscriptions, Catalog: dependencies.Catalog,
		Translations: dependencies.Translations}, dependencies.Log))
	calendarhandler.Register(handlers.Inbound, calendarhandler.New(calendarcmd.Handler{Players: dependencies.Players,
		Bindings: dependencies.Bindings, Subscriptions: dependencies.Subscriptions, Catalog: dependencies.Catalog,
		Furniture: dependencies.Furniture, Permissions: dependencies.Permissions}, dependencies.Log))
	hotelviewhandler.Register(handlers.Inbound, hotelviewhandler.New(hotelviewcmd.Handler{Players: dependencies.Players,
		Bindings: dependencies.Bindings, Subscriptions: dependencies.Subscriptions}, dependencies.Log))
}
