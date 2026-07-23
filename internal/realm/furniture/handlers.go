package furniture

import (
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	chatfilter "github.com/niflaot/pixels/internal/realm/chat/filter"
	realmconn "github.com/niflaot/pixels/internal/realm/connection"
	decorcmd "github.com/niflaot/pixels/internal/realm/furniture/commands/decor"
	interactcmd "github.com/niflaot/pixels/internal/realm/furniture/commands/interact"
	inventorycmd "github.com/niflaot/pixels/internal/realm/furniture/commands/inventory"
	movecmd "github.com/niflaot/pixels/internal/realm/furniture/commands/move"
	pickupcmd "github.com/niflaot/pixels/internal/realm/furniture/commands/pickup"
	placecmd "github.com/niflaot/pixels/internal/realm/furniture/commands/place"
	presentcmd "github.com/niflaot/pixels/internal/realm/furniture/commands/present"
	decorhandler "github.com/niflaot/pixels/internal/realm/furniture/handlers/decor"
	inventoryhandler "github.com/niflaot/pixels/internal/realm/furniture/handlers/inventory"
	movehandler "github.com/niflaot/pixels/internal/realm/furniture/handlers/move"
	pickuphandler "github.com/niflaot/pixels/internal/realm/furniture/handlers/pickup"
	placehandler "github.com/niflaot/pixels/internal/realm/furniture/handlers/place"
	presenthandler "github.com/niflaot/pixels/internal/realm/furniture/handlers/present"
	usehandler "github.com/niflaot/pixels/internal/realm/furniture/handlers/use"
	"github.com/niflaot/pixels/internal/realm/furniture/interactions"
	essential "github.com/niflaot/pixels/internal/realm/furniture/interactions/essential"
	lovelock "github.com/niflaot/pixels/internal/realm/furniture/interactions/lovelock"
	mysterybox "github.com/niflaot/pixels/internal/realm/furniture/interactions/mysterybox"
	rentable "github.com/niflaot/pixels/internal/realm/furniture/interactions/rentable"
	roller "github.com/niflaot/pixels/internal/realm/furniture/interactions/roller"
	stackheight "github.com/niflaot/pixels/internal/realm/furniture/interactions/stackheight"
	teleport "github.com/niflaot/pixels/internal/realm/furniture/interactions/teleport"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	"github.com/niflaot/pixels/internal/realm/furniture/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	roomwordfilter "github.com/niflaot/pixels/internal/realm/room/control/wordfilter"
	roomdecor "github.com/niflaot/pixels/internal/realm/room/decoration"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/i18n"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// HandlerDeps contains furniture handler dependencies.
type HandlerDeps struct {
	fx.In

	// Players stores live player state.
	Players *playerlive.Registry
	// Bindings stores player connection bindings.
	Bindings *binding.Registry
	// Furniture manages placed and inventory furniture records.
	Furniture service.Manager
	// Groups resolves warmed linked social-group furniture identity.
	Groups furnituremodel.GroupPolicy
	// PlayerDirectory resolves player records for gift tags.
	PlayerDirectory playerservice.Finder
	// PlayerAdmin persists mannequin outfit application.
	PlayerAdmin playerservice.AdminManager
	// FurnitureStates changes durable furniture interaction state.
	FurnitureStates service.StateUpdater
	// StackHeights changes exact custom stack-helper surfaces.
	StackHeights service.StackHeightUpdater
	// Runtime stores active rooms.
	Runtime *roomlive.Registry
	// Permissions resolves global furniture management authority.
	Permissions permissionservice.Checker
	// Connections stores active network connections.
	Connections *netconn.Registry
	// Events publishes furniture lifecycle events.
	Events *bus.Bus
	// Translations resolves end-user messages.
	Translations i18n.Translator
	// Log records command dispatch.
	Log *zap.Logger
	// Teleports coordinates paired furniture travel.
	Teleports *teleport.Service
	// Interactions resolves generic furniture state behavior.
	Interactions *interactions.Registry
	// EssentialInteractions coordinates specialized furniture behavior.
	EssentialInteractions *essential.Service
	// Rollers coordinates autonomous roller furniture.
	Rollers *roller.Service
	// Rentables coordinates rental, extension, cancellation, and buyout.
	Rentables *rentable.Service
	// Lovelocks coordinates two-player friend furniture confirmation.
	Lovelocks *lovelock.Service
	// MysteryBoxes handles reveals, trophies, and account key state.
	MysteryBoxes mysterybox.Handler
	// Decoration manages room surfaces and mood-light presets.
	Decoration *roomdecor.Service
	// WordFilters applies room-local filtering to decorator text.
	WordFilters roomwordfilter.Manager
	// GlobalFilter applies hotel-wide filtering to decorator text.
	GlobalFilter *chatfilter.Service
}

// RegisterConnectionHandlers registers furniture packet handlers.
func RegisterConnectionHandlers(handlers *realmconn.Handlers, deps HandlerDeps) {
	if handlers == nil || handlers.Inbound == nil {
		return
	}

	inventoryhandler.Register(handlers.Inbound, inventoryhandler.New(inventorycmd.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Furniture: deps.Furniture, Groups: deps.Groups,
	}, deps.Log))
	inventoryhandler.RegisterUnseen(handlers.Inbound)
	placehandler.Register(handlers.Inbound, placehandler.New(placecmd.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Furniture: deps.Furniture, PlayerDirectory: deps.PlayerDirectory,
		Runtime: deps.Runtime, Permissions: deps.Permissions, Connections: deps.Connections, Events: deps.Events, Groups: deps.Groups, Translations: deps.Translations, Log: deps.Log,
		RollerNoRules: deps.Rollers.NoRules(),
	}, deps.Log))
	moveCommands := movecmd.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Furniture: deps.Furniture,
		Runtime: deps.Runtime, Permissions: deps.Permissions, Connections: deps.Connections, Events: deps.Events, Translations: deps.Translations, Log: deps.Log,
		RollerNoRules: deps.Rollers.NoRules(),
	}
	movehandler.Register(handlers.Inbound, movehandler.New(moveCommands, deps.Log))
	movehandler.RegisterWall(handlers.Inbound, movehandler.NewWall(moveCommands, deps.Log))
	pickupCommands := pickupcmd.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Furniture: deps.Furniture,
		Runtime: deps.Runtime, Permissions: deps.Permissions, Connections: deps.Connections, Events: deps.Events, Translations: deps.Translations, Log: deps.Log,
	}
	pickuphandler.Register(handlers.Inbound, pickuphandler.New(pickupCommands, deps.Log))
	pickuphandler.RegisterWall(handlers.Inbound, pickuphandler.NewWall(pickupCommands, deps.Log))
	presenthandler.Register(handlers.Inbound, presenthandler.New(presentcmd.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Furniture: deps.Furniture,
		Runtime: deps.Runtime, Connections: deps.Connections, Log: deps.Log,
	}, deps.Log))
	decorCommands := decorcmd.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Furniture: deps.Furniture, States: deps.FurnitureStates,
		Decoration: deps.Decoration, Runtime: deps.Runtime, Permissions: deps.Permissions, Connections: deps.Connections,
		WordFilters: deps.WordFilters, GlobalFilter: deps.GlobalFilter, PlayerAdmin: deps.PlayerAdmin,
		Events: deps.Events, Translations: deps.Translations, Log: deps.Log,
	}
	decorhandler.Register(handlers.Inbound, decorCommands, deps.Log)
	interactionHandler := interactcmd.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Furniture: deps.Furniture, States: deps.FurnitureStates,
		Runtime: deps.Runtime, Permissions: deps.Permissions, Connections: deps.Connections,
		Events: deps.Events, Translations: deps.Translations, Behaviors: deps.Interactions, Teleports: deps.Teleports, Essentials: deps.EssentialInteractions, Decorator: &decorCommands, Log: deps.Log,
	}
	usehandler.Register(handlers.Inbound, usehandler.New(interactionHandler, deps.Log))
	usehandler.RegisterDedicated(handlers.Inbound,
		usehandler.NewDiceActivate(interactionHandler, deps.Log), usehandler.NewDiceClose(interactionHandler, deps.Log),
		usehandler.NewColorWheel(interactionHandler, deps.Log), usehandler.NewWall(interactionHandler, deps.Log),
	)
	usehandler.RegisterModernBridges(handlers.Inbound,
		usehandler.NewOneWayDoor(interactionHandler, deps.Log), usehandler.NewRandomState(interactionHandler, deps.Log))
	stackheight.Register(handlers.Inbound, stackheight.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Furniture: deps.Furniture, States: deps.StackHeights,
		Runtime: deps.Runtime, Permissions: deps.Permissions, Connections: deps.Connections,
	})
	rentable.Register(handlers.Inbound, rentable.Handler{Players: deps.Players, Bindings: deps.Bindings, Service: deps.Rentables})
	lovelock.Register(handlers.Inbound, deps.EssentialInteractions, lovelock.Handler{Players: deps.Players, Bindings: deps.Bindings, Runtime: deps.Runtime, Service: deps.Lovelocks})
	mysterybox.Register(handlers.Inbound, deps.EssentialInteractions, deps.MysteryBoxes)
}
