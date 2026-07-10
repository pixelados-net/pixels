package catalog

import (
	buycmd "github.com/niflaot/pixels/internal/realm/catalog/commands/buy"
	pagecmd "github.com/niflaot/pixels/internal/realm/catalog/commands/page"
	pagescmd "github.com/niflaot/pixels/internal/realm/catalog/commands/pages"
	buyhandler "github.com/niflaot/pixels/internal/realm/catalog/handlers/buy"
	pagehandler "github.com/niflaot/pixels/internal/realm/catalog/handlers/page"
	pageshandler "github.com/niflaot/pixels/internal/realm/catalog/handlers/pages"
	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
	realmconn "github.com/niflaot/pixels/internal/realm/connection"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
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
	// Catalog manages catalog reads and purchases.
	Catalog catalogservice.Manager
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
		Players: dependencies.Players, Bindings: dependencies.Bindings, Catalog: dependencies.Catalog, Log: dependencies.Log,
	}, dependencies.Log))
}
