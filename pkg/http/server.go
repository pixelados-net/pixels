package http

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/pixels/pkg/build"
	"github.com/niflaot/pixels/pkg/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Start registers Fiber server lifecycle hooks.
func Start(lifecycle fx.Lifecycle, app *fiber.App, log *zap.Logger, config config.AppConfig, info build.Info) {
	lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			log.Info("starting pixels server", serverFields(config, info)...)

			go func() {
				if err := app.Listen(config.App.Address()); err != nil {
					log.Error("pixels server stopped", zap.Error(err))
				}
			}()

			return nil
		},
		OnStop: func(context.Context) error {
			return app.Shutdown()
		},
	})
}

// serverFields builds structured server startup log fields.
func serverFields(config config.AppConfig, info build.Info) []zap.Field {
	return []zap.Field{
		zap.String("environment", config.App.Environment),
		zap.String("host", config.App.Host),
		zap.Int("port", config.App.Port),
		zap.String("version", info.Version),
		zap.String("commit", info.Commit),
	}
}
