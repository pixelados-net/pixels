package http

import (
	"github.com/gofiber/contrib/fiberzap"
	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/pixels/internal/auth/sso"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	navservice "github.com/niflaot/pixels/internal/realm/navigator/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomlive "github.com/niflaot/pixels/internal/realm/room/live"
	roomservice "github.com/niflaot/pixels/internal/realm/room/service"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/build"
	"github.com/niflaot/pixels/pkg/config"
	ws "github.com/niflaot/pixels/pkg/http/websocket"
	"github.com/niflaot/pixels/pkg/i18n"
	"go.uber.org/zap"
)

// New creates the Fiber application.
func New(log *zap.Logger, config config.AppConfig, info build.Info, sso *sso.Service, websocket *ws.Adapter, registry *netconn.Registry, players *playerlive.Registry, rooms roomservice.Manager, runtime *roomlive.Registry, navigator navservice.Manager, currencies currencyservice.Reader, translations i18n.Translator) *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler:          errorHandler,
	})

	app.Use(fiberzap.New(fiberzap.Config{
		Logger:   log,
		Fields:   []string{"latency", "status", "method", "url", "error"},
		Messages: []string{"http server request failed", "http client request failed", "http request completed"},
	}))

	registerPublic(app, config, info, websocket, currencies, translations)
	app.Use(auth(config.App.AccessKey))
	registerPrivate(app, sso, registry, players, rooms, runtime, navigator, translations)

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
