// Package main starts the Pixels emulator.
package main

import (
	"github.com/niflaot/pixels/internal/auth/sso"
	permissionmodule "github.com/niflaot/pixels/internal/permission/module"
	pluginmodule "github.com/niflaot/pixels/internal/plugin"
	realmbot "github.com/niflaot/pixels/internal/realm/bot"
	realmcamera "github.com/niflaot/pixels/internal/realm/camera"
	realmcatalog "github.com/niflaot/pixels/internal/realm/catalog"
	realmchat "github.com/niflaot/pixels/internal/realm/chat"
	realmconn "github.com/niflaot/pixels/internal/realm/connection"
	realmcrafting "github.com/niflaot/pixels/internal/realm/crafting"
	realmfurniture "github.com/niflaot/pixels/internal/realm/furniture"
	realmgamecenter "github.com/niflaot/pixels/internal/realm/gamecenter"
	realmgroup "github.com/niflaot/pixels/internal/realm/group"
	realminventory "github.com/niflaot/pixels/internal/realm/inventory"
	realmmarketplace "github.com/niflaot/pixels/internal/realm/marketplace"
	realmmessenger "github.com/niflaot/pixels/internal/realm/messenger"
	realmmoderation "github.com/niflaot/pixels/internal/realm/moderation"
	realmnavigator "github.com/niflaot/pixels/internal/realm/navigator"
	realmpet "github.com/niflaot/pixels/internal/realm/pet"
	realmplayer "github.com/niflaot/pixels/internal/realm/player"
	realmprogression "github.com/niflaot/pixels/internal/realm/progression"
	realmroom "github.com/niflaot/pixels/internal/realm/room"
	realmsanction "github.com/niflaot/pixels/internal/realm/sanction"
	realmsession "github.com/niflaot/pixels/internal/realm/session"
	realmsubscription "github.com/niflaot/pixels/internal/realm/subscription"
	realmtrade "github.com/niflaot/pixels/internal/realm/trade"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/build"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/config"
	pixelhttp "github.com/niflaot/pixels/pkg/http"
	ws "github.com/niflaot/pixels/pkg/http/websocket"
	"github.com/niflaot/pixels/pkg/i18n"
	"github.com/niflaot/pixels/pkg/logger"
	"github.com/niflaot/pixels/pkg/postgres"
	"github.com/niflaot/pixels/pkg/redis"
	"github.com/niflaot/pixels/pkg/storage"
	"go.uber.org/fx"
)

// main starts the dependency graph.
func main() {
	newApp().Run()
}

// newApp builds the dependency graph.
func newApp() *fx.App {
	return fx.New(options()...)
}

// options returns the dependency graph options.
func options() []fx.Option {
	options := make([]fx.Option, 0, 20)
	options = append(options, build.Module)
	options = append(options, config.Module)
	options = append(options, netconn.Module)
	options = append(options, bus.Module)
	options = append(options, i18n.Module)
	options = append(options, permissionmodule.Module)
	options = append(options, realmconn.Module)
	options = append(options, realmcrafting.Module)
	options = append(options, realmcamera.Module)
	options = append(options, realmbot.Module)
	options = append(options, realmcatalog.Module)
	options = append(options, realmchat.Module)
	options = append(options, realmfurniture.Module)
	options = append(options, realmgamecenter.Module)
	options = append(options, realmgroup.Module)
	options = append(options, realminventory.Module)
	options = append(options, realmmessenger.Module)
	options = append(options, realmmoderation.Module)
	options = append(options, realmmarketplace.Module)
	options = append(options, realmnavigator.Module)
	options = append(options, realmpet.Module)
	options = append(options, realmplayer.Module)
	options = append(options, realmprogression.Module)
	options = append(options, realmroom.Module)
	options = append(options, realmsanction.Module)
	options = append(options, realmsession.Module)
	options = append(options, realmsubscription.Module)
	options = append(options, realmtrade.Module)
	options = append(options, sso.Module)
	options = append(options, ws.Module)
	options = append(options, pluginmodule.Module)
	options = append(options, pixelhttp.Module)
	options = append(options, logger.Module)
	options = append(options, postgres.Module)
	options = append(options, redis.Module)
	options = append(options, storage.Module)
	options = append(options, fx.WithLogger(logger.NewFx))

	return options
}
