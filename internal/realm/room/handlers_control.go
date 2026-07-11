package room

import (
	bancmd "github.com/niflaot/pixels/internal/realm/room/commands/moderation/ban"
	kickcmd "github.com/niflaot/pixels/internal/realm/room/commands/moderation/kick"
	listbanscmd "github.com/niflaot/pixels/internal/realm/room/commands/moderation/listbans"
	mutecmd "github.com/niflaot/pixels/internal/realm/room/commands/moderation/mute"
	unbancmd "github.com/niflaot/pixels/internal/realm/room/commands/moderation/unban"
	unmutecmd "github.com/niflaot/pixels/internal/realm/room/commands/moderation/unmute"
	grantcmd "github.com/niflaot/pixels/internal/realm/room/commands/rights/grant"
	listcmd "github.com/niflaot/pixels/internal/realm/room/commands/rights/list"
	relinquishcmd "github.com/niflaot/pixels/internal/realm/room/commands/rights/relinquish"
	revokecmd "github.com/niflaot/pixels/internal/realm/room/commands/rights/revoke"
	revokeallcmd "github.com/niflaot/pixels/internal/realm/room/commands/rights/revokeall"
	controlhandler "github.com/niflaot/pixels/internal/realm/room/handlers/control"
	banhandler "github.com/niflaot/pixels/internal/realm/room/handlers/moderation/ban"
	kickhandler "github.com/niflaot/pixels/internal/realm/room/handlers/moderation/kick"
	listbanshandler "github.com/niflaot/pixels/internal/realm/room/handlers/moderation/listbans"
	mutehandler "github.com/niflaot/pixels/internal/realm/room/handlers/moderation/mute"
	unbanhandler "github.com/niflaot/pixels/internal/realm/room/handlers/moderation/unban"
	granthandler "github.com/niflaot/pixels/internal/realm/room/handlers/rights/grant"
	listhandler "github.com/niflaot/pixels/internal/realm/room/handlers/rights/list"
	relinquishhandler "github.com/niflaot/pixels/internal/realm/room/handlers/rights/relinquish"
	revokehandler "github.com/niflaot/pixels/internal/realm/room/handlers/rights/revoke"
	revokeallhandler "github.com/niflaot/pixels/internal/realm/room/handlers/rights/revokeall"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// registerRightsHandlers registers room rights packet adapters.
func registerRightsHandlers(registry *netconn.HandlerRegistry, deps HandlerDeps) {
	granthandler.Register(registry, controlhandler.Wrap(granthandler.New(grantcmd.Handler{Players: deps.Players, Bindings: deps.Bindings, Rights: deps.Rights}, deps.Log), deps.Translations, deps.Log))
	revokehandler.Register(registry, controlhandler.Wrap(revokehandler.New(revokecmd.Handler{Players: deps.Players, Bindings: deps.Bindings, Rights: deps.Rights}, deps.Log), deps.Translations, deps.Log))
	revokeallhandler.Register(registry, controlhandler.Wrap(revokeallhandler.New(revokeallcmd.Handler{Players: deps.Players, Bindings: deps.Bindings, Rights: deps.Rights}, deps.Log), deps.Translations, deps.Log))
	relinquishhandler.Register(registry, controlhandler.Wrap(relinquishhandler.New(relinquishcmd.Handler{Players: deps.Players, Bindings: deps.Bindings, Rights: deps.Rights}, deps.Log), deps.Translations, deps.Log))
	listhandler.Register(registry, controlhandler.Wrap(listhandler.New(listcmd.Handler{Players: deps.Players, Bindings: deps.Bindings, Rights: deps.Rights}, deps.Log), deps.Translations, deps.Log))
}

// registerModerationHandlers registers room moderation packet adapters.
func registerModerationHandlers(registry *netconn.HandlerRegistry, deps HandlerDeps) {
	kickhandler.Register(registry, controlhandler.Wrap(kickhandler.New(kickcmd.Handler{Players: deps.Players, Bindings: deps.Bindings, Moderation: deps.Moderation, Runtime: deps.Runtime}, deps.Log), deps.Translations, deps.Log))
	mutehandler.Register(registry, controlhandler.Wrap(mutehandler.New(
		mutecmd.Handler{Players: deps.Players, Bindings: deps.Bindings, Moderation: deps.Moderation, Runtime: deps.Runtime},
		unmutecmd.Handler{Players: deps.Players, Bindings: deps.Bindings, Moderation: deps.Moderation}, deps.Log,
	), deps.Translations, deps.Log))
	banhandler.Register(registry, controlhandler.Wrap(banhandler.New(bancmd.Handler{Players: deps.Players, Bindings: deps.Bindings, Moderation: deps.Moderation, Runtime: deps.Runtime}, deps.Log), deps.Translations, deps.Log))
	unbanhandler.Register(registry, controlhandler.Wrap(unbanhandler.New(unbancmd.Handler{Players: deps.Players, Bindings: deps.Bindings, Moderation: deps.Moderation}, deps.Log), deps.Translations, deps.Log))
	listbanshandler.Register(registry, controlhandler.Wrap(listbanshandler.New(listbanscmd.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Moderation: deps.Moderation,
		Rooms: deps.Rooms, Authorize: deps.Settings,
	}, deps.Log), deps.Translations, deps.Log))
}
