package catalog

import (
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	builderscmd "github.com/niflaot/pixels/internal/realm/catalog/commands/builders"
	buycmd "github.com/niflaot/pixels/internal/realm/catalog/commands/buy"
	commercecmd "github.com/niflaot/pixels/internal/realm/catalog/commands/commerce"
	pagecmd "github.com/niflaot/pixels/internal/realm/catalog/commands/page"
	pagescmd "github.com/niflaot/pixels/internal/realm/catalog/commands/pages"
	"github.com/niflaot/pixels/internal/realm/catalog/gift"
	buyhandler "github.com/niflaot/pixels/internal/realm/catalog/handlers/buy"
	commercehandler "github.com/niflaot/pixels/internal/realm/catalog/handlers/commerce"
	pagehandler "github.com/niflaot/pixels/internal/realm/catalog/handlers/page"
	pageshandler "github.com/niflaot/pixels/internal/realm/catalog/handlers/pages"
	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
	realmconn "github.com/niflaot/pixels/internal/realm/connection"
	placecmd "github.com/niflaot/pixels/internal/realm/furniture/commands/place"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/i18n"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// HandlerDeps contains catalog packet handler dependencies.
type HandlerDeps struct {
	fx.In

	// Players stores live player compositions.
	Players *playerlive.Registry
	// Bindings maps authenticated connections to players.
	Bindings *binding.Registry
	// Connections stores active transport-agnostic sessions.
	Connections *netconn.Registry
	// Catalog manages catalog reads and purchases.
	Catalog catalogservice.Manager
	// CatalogService exposes extended catalog commerce behavior.
	CatalogService *catalogservice.Service
	// Builders stores discontinued-tier placement policy.
	Builders builderscmd.Config
	// Furniture manages Builders Club placement persistence.
	Furniture furnitureservice.Manager
	// PlayerDirectory resolves durable furniture owner metadata.
	PlayerDirectory playerservice.Finder
	// Runtime stores active destination rooms.
	Runtime *roomlive.Registry
	// Permissions resolves furniture placement authority.
	Permissions permissionservice.Checker
	// Events publishes successful furniture placements.
	Events bus.Publisher
	// Groups resolves linked social-group furniture state.
	Groups furnituremodel.GroupPolicy
	// Club purchases subscription offers selected from catalog pages.
	Club buycmd.ClubPurchaser
	// GiftOptions stores immutable wrapping choices.
	GiftOptions gift.Options
	// Translations localizes catalog page text.
	Translations i18n.Translator
	// Log records command dispatch and purchase failures.
	Log *zap.Logger
}

// RegisterConnectionHandlers registers catalog packet handlers.
func RegisterConnectionHandlers(handlers *realmconn.Handlers, dependencies HandlerDeps) {
	if handlers == nil || handlers.Inbound == nil {
		return
	}
	pageshandler.Register(handlers.Inbound, pageshandler.New(pagescmd.Handler{
		Players: dependencies.Players, Bindings: dependencies.Bindings, Catalog: dependencies.Catalog, Translations: dependencies.Translations,
	}, dependencies.Log))
	pagehandler.Register(handlers.Inbound, pagehandler.New(pagecmd.Handler{
		Players: dependencies.Players, Bindings: dependencies.Bindings, Catalog: dependencies.Catalog, Translations: dependencies.Translations,
	}, dependencies.Log))
	buyhandler.Register(handlers.Inbound, buyhandler.New(buycmd.Handler{
		Players: dependencies.Players, Bindings: dependencies.Bindings, Catalog: dependencies.Catalog,
		Club: dependencies.Club, Log: dependencies.Log, Translations: dependencies.Translations,
	}, dependencies.Log))
	commercehandler.Register(handlers.Inbound, commercehandler.New(commercecmd.Handler{
		Players: dependencies.Players, Bindings: dependencies.Bindings, Catalog: dependencies.CatalogService,
		Connections: dependencies.Connections, GiftOptions: dependencies.GiftOptions, Log: dependencies.Log,
	}, dependencies.Log))
	builderscmd.Register(handlers.Inbound, builderscmd.Handler{
		Config: dependencies.Builders, Players: dependencies.Players, Bindings: dependencies.Bindings,
		Runtime: dependencies.Runtime, Catalog: dependencies.CatalogService,
		Place: placecmd.Handler{Players: dependencies.Players, Bindings: dependencies.Bindings,
			Furniture: dependencies.Furniture, PlayerDirectory: dependencies.PlayerDirectory,
			Runtime: dependencies.Runtime, Permissions: dependencies.Permissions,
			Connections: dependencies.Connections, Events: dependencies.Events,
			Groups: dependencies.Groups, Translations: dependencies.Translations, Log: dependencies.Log},
	})
}
