// Package main starts the Pixels emulator.
package main

import (
	"github.com/niflaot/pixels/internal/auth/sso"
	permissionmodule "github.com/niflaot/pixels/internal/permission/module"
	realmcatalog "github.com/niflaot/pixels/internal/realm/catalog"
	realmconn "github.com/niflaot/pixels/internal/realm/connection"
	realmfurniture "github.com/niflaot/pixels/internal/realm/furniture"
	realminventory "github.com/niflaot/pixels/internal/realm/inventory"
	realmnavigator "github.com/niflaot/pixels/internal/realm/navigator"
	realmplayer "github.com/niflaot/pixels/internal/realm/player"
	realmroom "github.com/niflaot/pixels/internal/realm/room"
	realmsession "github.com/niflaot/pixels/internal/realm/session"
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
	options = append(options, pixelhttp.Module)
	options = append(options, realmconn.Module)
	options = append(options, realmcatalog.Module)
	options = append(options, realmfurniture.Module)
	options = append(options, realminventory.Module)
	options = append(options, realmnavigator.Module)
	options = append(options, realmplayer.Module)
	options = append(options, realmroom.Module)
	options = append(options, realmsession.Module)
	options = append(options, sso.Module)
	options = append(options, ws.Module)
	options = append(options, logger.Module)
	options = append(options, postgres.Module)
	options = append(options, redis.Module)
	options = append(options, fx.WithLogger(logger.NewFx))

	return options
}
