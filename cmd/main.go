// Package main starts the Pixels emulator.
package main

import (
	"github.com/niflaot/pixels/pkg/build"
	"github.com/niflaot/pixels/pkg/config"
	pixelhttp "github.com/niflaot/pixels/pkg/http"
	"github.com/niflaot/pixels/pkg/logger"
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
	options := make([]fx.Option, 0, 6)
	options = append(options, build.Module)
	options = append(options, config.Module)
	options = append(options, pixelhttp.Module)
	options = append(options, logger.Module)
	options = append(options, redis.Module)
	options = append(options, fx.WithLogger(logger.NewFx))

	return options
}
