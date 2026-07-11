package room

import (
	moderationcmd "github.com/niflaot/pixels/internal/realm/room/control/commands/moderation"
	rightscmd "github.com/niflaot/pixels/internal/realm/room/control/commands/rights"
	controlhandler "github.com/niflaot/pixels/internal/realm/room/control/handlers/control"
	moderationhandler "github.com/niflaot/pixels/internal/realm/room/control/handlers/moderation"
	rightshandler "github.com/niflaot/pixels/internal/realm/room/control/handlers/rights"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// registerRightsHandlers registers room rights packet adapters.
func registerRightsHandlers(registry *netconn.HandlerRegistry, deps HandlerDeps) {
	rightshandler.RegisterGrant(registry, controlhandler.Wrap(rightshandler.NewGrant(rightscmd.GrantHandler{Players: deps.Players, Bindings: deps.Bindings, Rights: deps.Rights}, deps.Log), deps.Translations, deps.Log))
	rightshandler.RegisterRevoke(registry, controlhandler.Wrap(rightshandler.NewRevoke(rightscmd.RevokeHandler{Players: deps.Players, Bindings: deps.Bindings, Rights: deps.Rights}, deps.Log), deps.Translations, deps.Log))
	rightshandler.RegisterRevokeAll(registry, controlhandler.Wrap(rightshandler.NewRevokeAll(rightscmd.RevokeAllHandler{Players: deps.Players, Bindings: deps.Bindings, Rights: deps.Rights}, deps.Log), deps.Translations, deps.Log))
	rightshandler.RegisterRelinquish(registry, controlhandler.Wrap(rightshandler.NewRelinquish(rightscmd.RelinquishHandler{Players: deps.Players, Bindings: deps.Bindings, Rights: deps.Rights}, deps.Log), deps.Translations, deps.Log))
	rightshandler.RegisterList(registry, controlhandler.Wrap(rightshandler.NewList(rightscmd.ListHandler{Players: deps.Players, Bindings: deps.Bindings, Rights: deps.Rights}, deps.Log), deps.Translations, deps.Log))
}

// registerModerationHandlers registers room moderation packet adapters.
func registerModerationHandlers(registry *netconn.HandlerRegistry, deps HandlerDeps) {
	moderationhandler.RegisterKick(registry, controlhandler.Wrap(moderationhandler.NewKick(moderationcmd.KickHandler{Players: deps.Players, Bindings: deps.Bindings, Moderation: deps.Moderation, Runtime: deps.Runtime}, deps.Log), deps.Translations, deps.Log))
	moderationhandler.RegisterMute(registry, controlhandler.Wrap(moderationhandler.NewMute(
		moderationcmd.MuteHandler{Players: deps.Players, Bindings: deps.Bindings, Moderation: deps.Moderation, Runtime: deps.Runtime},
		moderationcmd.UnmuteHandler{Players: deps.Players, Bindings: deps.Bindings, Moderation: deps.Moderation}, deps.Log,
	), deps.Translations, deps.Log))
	moderationhandler.RegisterBan(registry, controlhandler.Wrap(moderationhandler.NewBan(moderationcmd.BanHandler{Players: deps.Players, Bindings: deps.Bindings, Moderation: deps.Moderation, Runtime: deps.Runtime}, deps.Log), deps.Translations, deps.Log))
	moderationhandler.RegisterUnban(registry, controlhandler.Wrap(moderationhandler.NewUnban(moderationcmd.UnbanHandler{Players: deps.Players, Bindings: deps.Bindings, Moderation: deps.Moderation}, deps.Log), deps.Translations, deps.Log))
	moderationhandler.RegisterBanList(registry, controlhandler.Wrap(moderationhandler.NewBanList(moderationcmd.BanListHandler{
		Players: deps.Players, Bindings: deps.Bindings, Moderation: deps.Moderation,
		Rooms: deps.Rooms, Authorize: deps.Settings,
	}, deps.Log), deps.Translations, deps.Log))
}
