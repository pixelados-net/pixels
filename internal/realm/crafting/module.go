package crafting

import (
	realmconn "github.com/niflaot/pixels/internal/realm/connection"
	craftingconfig "github.com/niflaot/pixels/internal/realm/crafting/config"
	craftingdb "github.com/niflaot/pixels/internal/realm/crafting/database"
	craftingexchange "github.com/niflaot/pixels/internal/realm/crafting/exchange"
	exchangehandlers "github.com/niflaot/pixels/internal/realm/crafting/exchange/handlers"
	craftingrecipe "github.com/niflaot/pixels/internal/realm/crafting/recipe"
	recipehandlers "github.com/niflaot/pixels/internal/realm/crafting/recipe/handlers"
	craftingrecord "github.com/niflaot/pixels/internal/realm/crafting/record"
	craftingrecycler "github.com/niflaot/pixels/internal/realm/crafting/recycler"
	recyclerhandlers "github.com/niflaot/pixels/internal/realm/crafting/recycler/handlers"
	"go.uber.org/fx"
)

// Module provides the complete crafting, recycler, and exchange realm.
var Module = fx.Module("realm-crafting", fx.Provide(craftingconfig.Load, craftingdb.New, NewStore, craftingrecipe.New, craftingrecycler.New, craftingexchange.New, recipehandlers.New, recyclerhandlers.New, exchangehandlers.New), fx.Invoke(RegisterConnectionHandlers))

// NewStore exposes PostgreSQL persistence through the crafting contract.
func NewStore(repository *craftingdb.Repository) craftingrecord.Store { return repository }

// RegisterConnectionHandlers registers every crafting realm packet adapter.
func RegisterConnectionHandlers(handlers *realmconn.Handlers, recipes *recipehandlers.Handler, recycler *recyclerhandlers.Handler, exchange *exchangehandlers.Handler) {
	if handlers == nil || handlers.Inbound == nil {
		return
	}
	recipehandlers.Register(handlers.Inbound, recipes)
	recyclerhandlers.Register(handlers.Inbound, recycler)
	exchangehandlers.Register(handlers.Inbound, exchange)
}
