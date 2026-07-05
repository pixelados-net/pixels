package http

import (
	"github.com/gofiber/contrib/fiberzap"
	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/pixels/internal/auth/sso"
	"github.com/niflaot/pixels/pkg/build"
	"github.com/niflaot/pixels/pkg/config"
	"go.uber.org/zap"
)

// New creates the Fiber application.
func New(log *zap.Logger, config config.AppConfig, info build.Info, sso *sso.Service) *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler:          errorHandler,
	})

	app.Use(fiberzap.New(fiberzap.Config{
		Logger: log,
		Fields: []string{"latency", "status", "method", "url"},
	}))

	registerPublic(app, config, info)
	app.Use(auth(config.App.AccessKey))
	registerPrivate(app, sso)

	return app
}

// errorHandler writes JSON error responses.
func errorHandler(ctx *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	if fiberError, ok := err.(*fiber.Error); ok {
		code = fiberError.Code
	}

	return ctx.Status(code).JSON(ErrorResponse{
		Error: err.Error(),
	})
}
