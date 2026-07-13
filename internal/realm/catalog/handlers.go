package catalog

import (
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
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
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
}
