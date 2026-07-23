// Package plugin wires native dynamic plugins into controlled host capabilities.
package plugin

import (
	"context"

	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	plugincommand "github.com/niflaot/pixels/internal/plugin/command"
	pluginconfig "github.com/niflaot/pixels/internal/plugin/config"
	pluginevent "github.com/niflaot/pixels/internal/plugin/event"
	pluginhost "github.com/niflaot/pixels/internal/plugin/host"
	"github.com/niflaot/pixels/internal/plugin/loader"
	chatsend "github.com/niflaot/pixels/internal/realm/chat/send"
	realmconn "github.com/niflaot/pixels/internal/realm/connection"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/http/pluginroutes"
	"github.com/niflaot/pixels/pkg/i18n"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides dynamic plugin discovery and controlled SDK capabilities.
var Module = fx.Module(
	"plugins",
	fx.Provide(
		newEventHub,
		newCommandTree,
		newBackend,
		loader.NewNative,
		loader.New,
	),
	fx.Invoke(connectRuntime),
	fx.Invoke(loadPlugins),
)

// newEventHub creates the shared plugin event dispatcher.
func newEventHub(config pluginconfig.Config, log *zap.Logger) *pluginevent.Hub {
	return pluginevent.NewHub(config.CallbackTimeout, log)
}

// newCommandTree creates the shared plugin command dispatcher.
func newCommandTree(config pluginconfig.Config, translations i18n.Translator, log *zap.Logger) *plugincommand.Tree {
	return plugincommand.NewTree(config.CommandPrefix, config.CallbackTimeout, translations, log)
}

// newBackend composes internal registries behind the controlled plugin facade.
func newBackend(config pluginconfig.Config, players *playerlive.Registry, bindings *binding.Registry, connections *netconn.Registry, handlers *realmconn.Handlers, permissions permissionservice.Checker, routes *pluginroutes.Registry, events *pluginevent.Hub, commands *plugincommand.Tree, log *zap.Logger) *pluginhost.Backend {
	return pluginhost.NewBackend(players, bindings, connections, handlers.Inbound, permissions, routes, events, commands, config.CallbackTimeout, log)
}

// connectRuntime installs command, chat-event, and lifecycle bridges before loading plugins.
func connectRuntime(backend *pluginhost.Backend, subscriber bus.Subscriber, chat *chatsend.Service) error {
	return backend.Connect(subscriber, chat)
}

// loadPlugins completes registration before the HTTP application is assembled.
func loadPlugins(plugins *loader.Loader) error {
	return plugins.Load(context.Background())
}
