package http

import (
	"github.com/gofiber/contrib/fiberzap"
	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/pixels/internal/auth/sso"
	navservice "github.com/niflaot/pixels/internal/realm/navigator/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/live"
	roomservice "github.com/niflaot/pixels/internal/realm/room/service"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/build"
	"github.com/niflaot/pixels/pkg/config"
	ws "github.com/niflaot/pixels/pkg/http/websocket"
	"go.uber.org/zap"
)

// New creates the Fiber application.
func New(log *zap.Logger, config config.AppConfig, info build.Info, sso *sso.Service, websocket *ws.Adapter, registry *netconn.Registry, rooms roomservice.Manager, runtime *roomlive.Registry, navigator navservice.Manager) *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler:          errorHandler,
	})

	app.Use(fiberzap.New(fiberzap.Config{
		Logger:   log,
		Fields:   []string{"latency", "status", "method", "url", "error"},
		Messages: []string{"http server request failed", "http client request failed", "http request completed"},
	}))

	registerPublic(app, config, info, websocket)
	app.Use(auth(config.App.AccessKey))
	registerPrivate(app, sso, registry, rooms, runtime, navigator)

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
