// Package host composes capability-specific dynamic-plugin facades.
package host

import (
	"time"

	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	plugincommand "github.com/niflaot/pixels/internal/plugin/command"
	pluginevent "github.com/niflaot/pixels/internal/plugin/event"
	pluginpermission "github.com/niflaot/pixels/internal/plugin/permission"
	pluginplayer "github.com/niflaot/pixels/internal/plugin/player"
	pluginroute "github.com/niflaot/pixels/internal/plugin/route"
	pluginruntime "github.com/niflaot/pixels/internal/plugin/runtime"
	chatsend "github.com/niflaot/pixels/internal/realm/chat/send"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/http/pluginroutes"
	sdkplugin "github.com/niflaot/pixels/sdk/plugin"
	"go.uber.org/zap"
)

// Backend owns dependencies used to construct scoped plugin hosts.
type Backend struct {
	// players stores connected player state.
	players *playerlive.Registry
	// bindings resolves authenticated connections.
	bindings *binding.Registry
	// connections sends packets and disconnects sessions.
	connections *netconn.Registry
	// inbound stores the shared inbound packet pipeline.
	inbound *netconn.HandlerRegistry
	// permissions resolves real permission nodes.
	permissions permissionservice.Checker
	// routes stores isolated route declarations.
	routes *pluginroutes.Registry
	// events dispatches plugin-facing typed events.
	events *pluginevent.Hub
	// commands dispatches plugin chat commands.
	commands *plugincommand.Tree
	// timeout bounds plugin callbacks.
	timeout time.Duration
	// log records isolated plugin failures.
	log *zap.Logger
}

// NewBackend creates the shared dynamic plugin host backend.
func NewBackend(players *playerlive.Registry, bindings *binding.Registry, connections *netconn.Registry, inbound *netconn.HandlerRegistry, permissions permissionservice.Checker, routes *pluginroutes.Registry, events *pluginevent.Hub, commands *plugincommand.Tree, timeout time.Duration, log *zap.Logger) *Backend {
	if log == nil {
		log = zap.NewNop()
	}
	return &Backend{players: players, bindings: bindings, connections: connections, inbound: inbound, permissions: permissions, routes: routes, events: events, commands: commands, timeout: timeout, log: log}
}

// HostFor creates a namespace-enforcing facade for one plugin scope.
func (backend *Backend) HostFor(scope *pluginruntime.Scope) sdkplugin.Host {
	players := pluginplayer.NewAccess(backend.players, backend.bindings, backend.connections, backend.inbound, backend.permissions, backend.timeout, backend.log, scope)
	return &scopedHost{
		players:     players,
		routes:      pluginroute.NewAccess(backend.routes, scope),
		events:      pluginevent.NewAccess(backend.events, scope),
		commands:    plugincommand.NewAccess(backend.commands, scope),
		permissions: pluginpermission.NewAccess(scope),
	}
}

// Connect installs plugin commands, events, and player lifecycle forwarding.
func (backend *Backend) Connect(subscriber bus.Subscriber, chat *chatsend.Service) error {
	access := pluginplayer.NewAccess(backend.players, backend.bindings, backend.connections, backend.inbound, backend.permissions, backend.timeout, backend.log, pluginruntime.NewScope("pixels"))
	backend.commands.SetPlayers(access)
	chat.SetPluginRuntime(backend.commands, backend.events)
	return backend.events.RegisterPlayerConnected(subscriber, access)
}

// scopedHost implements the single plugin-facing capability facade.
type scopedHost struct {
	// players exposes scoped player operations.
	players sdkplugin.PlayerAccess
	// routes exposes scoped route registration.
	routes sdkplugin.RouteRegistrar
	// events exposes scoped listener registration.
	events sdkplugin.EventHub
	// commands exposes scoped command registration.
	commands sdkplugin.CommandTree
	// permissions exposes scoped node registration.
	permissions sdkplugin.PermissionRegistrar
}

// Players returns bounded live-player access.
func (host *scopedHost) Players() sdkplugin.PlayerAccess { return host.players }

// Routes returns isolated HTTP route registration.
func (host *scopedHost) Routes() sdkplugin.RouteRegistrar { return host.routes }

// Events returns plugin-facing event registration.
func (host *scopedHost) Events() sdkplugin.EventHub { return host.events }

// Commands returns the shared Brigadier command tree.
func (host *scopedHost) Commands() sdkplugin.CommandTree { return host.commands }

// Permissions returns namespaced permission registration.
func (host *scopedHost) Permissions() sdkplugin.PermissionRegistrar { return host.permissions }
