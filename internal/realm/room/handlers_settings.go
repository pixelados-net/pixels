package room

import (
	mutetogglecmd "github.com/niflaot/pixels/internal/realm/room/commands/mute/toggle"
	settingsrequestcmd "github.com/niflaot/pixels/internal/realm/room/commands/settings/request"
	settingssavecmd "github.com/niflaot/pixels/internal/realm/room/commands/settings/save"
	filtermodifycmd "github.com/niflaot/pixels/internal/realm/room/commands/wordfilter/modify"
	filterrequestcmd "github.com/niflaot/pixels/internal/realm/room/commands/wordfilter/request"
	controlhandler "github.com/niflaot/pixels/internal/realm/room/handlers/control"
	mutetogglehandler "github.com/niflaot/pixels/internal/realm/room/handlers/mute/toggle"
	settingsrequesthandler "github.com/niflaot/pixels/internal/realm/room/handlers/settings/request"
	settingssavehandler "github.com/niflaot/pixels/internal/realm/room/handlers/settings/save"
	filtermodifyhandler "github.com/niflaot/pixels/internal/realm/room/handlers/wordfilter/modify"
	filterrequesthandler "github.com/niflaot/pixels/internal/realm/room/handlers/wordfilter/request"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// registerSettingsHandlers registers room settings, filters, and mute-all packet adapters.
func registerSettingsHandlers(registry *netconn.HandlerRegistry, deps HandlerDeps) {
	settingsrequesthandler.Register(registry, controlhandler.Wrap(settingsrequesthandler.New(settingsrequestcmd.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Rooms: deps.Rooms, Authorize: deps.Settings,
	}, deps.Log), deps.Translations, deps.Log))
	settingssavehandler.Register(registry, controlhandler.Wrap(settingssavehandler.New(settingssavecmd.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Rooms: deps.ConfigRooms, Authorize: deps.Settings,
		Runtime: deps.Runtime, Connections: deps.Connections, Events: deps.Events,
	}, deps.Log), deps.Translations, deps.Log))
	filterrequesthandler.Register(registry, controlhandler.Wrap(filterrequesthandler.New(filterrequestcmd.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Rooms: deps.Rooms, Authorize: deps.Settings, Filters: deps.WordFilters,
	}, deps.Log), deps.Translations, deps.Log))
	filtermodifyhandler.Register(registry, controlhandler.Wrap(filtermodifyhandler.New(filtermodifycmd.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Filters: deps.WordFilters, Events: deps.Events,
	}, deps.Log), deps.Translations, deps.Log))
	mutetogglehandler.Register(registry, controlhandler.Wrap(mutetogglehandler.New(mutetogglecmd.Handler{
		Players: deps.Players, Bindings: deps.Bindings, Rooms: deps.Rooms, Authorize: deps.Settings,
		Runtime: deps.Runtime, Connections: deps.Connections, Events: deps.Events,
	}, deps.Log), deps.Translations, deps.Log))
}
