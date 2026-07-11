package room

import (
	settingscmd "github.com/niflaot/pixels/internal/realm/room/control/commands/settings"
	controlhandler "github.com/niflaot/pixels/internal/realm/room/control/handlers/control"
	settingshandler "github.com/niflaot/pixels/internal/realm/room/control/handlers/settings"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// registerSettingsHandlers registers room settings, filters, and mute-all packet adapters.
func registerSettingsHandlers(registry *netconn.HandlerRegistry, deps HandlerDeps) {
	settingshandler.RegisterRequest(registry, controlhandler.Wrap(settingshandler.NewRequest(settingscmd.RequestHandler{
		Players: deps.Players, Bindings: deps.Bindings, Rooms: deps.Rooms, Authorize: deps.Settings,
	}, deps.Log), deps.Translations, deps.Log))
	settingshandler.RegisterSave(registry, controlhandler.Wrap(settingshandler.NewSave(settingscmd.SaveHandler{
		Players: deps.Players, Bindings: deps.Bindings, Rooms: deps.ConfigRooms, Authorize: deps.Settings,
		Runtime: deps.Runtime, Connections: deps.Connections, Events: deps.Events,
	}, deps.Log), deps.Translations, deps.Log))
	settingshandler.RegisterFilterRequest(registry, controlhandler.Wrap(settingshandler.NewFilterRequest(settingscmd.FilterRequestHandler{
		Players: deps.Players, Bindings: deps.Bindings, Rooms: deps.Rooms, Authorize: deps.Settings, Filters: deps.WordFilters,
	}, deps.Log), deps.Translations, deps.Log))
	settingshandler.RegisterFilterModify(registry, controlhandler.Wrap(settingshandler.NewFilterModify(settingscmd.FilterModifyHandler{
		Players: deps.Players, Bindings: deps.Bindings, Filters: deps.WordFilters, Events: deps.Events,
	}, deps.Log), deps.Translations, deps.Log))
	settingshandler.RegisterMuteAll(registry, controlhandler.Wrap(settingshandler.NewMuteAll(settingscmd.MuteAllHandler{
		Players: deps.Players, Bindings: deps.Bindings, Rooms: deps.Rooms, Authorize: deps.Settings,
		Runtime: deps.Runtime, Connections: deps.Connections, Events: deps.Events,
	}, deps.Log), deps.Translations, deps.Log))
}
