package http

import (
	"github.com/gofiber/contrib/fiberzap"
	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/pixels/internal/auth/sso"
	navservice "github.com/niflaot/pixels/internal/realm/navigator/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/live"
	roomservice "github.com/niflaot/pixels/internal/realm/room/service"
	"github.com/niflaot/pixels/pkg/build"
	"github.com/niflaot/pixels/pkg/config"
	catalogroutes "github.com/niflaot/pixels/pkg/http/catalog/routes"
	currencyroutes "github.com/niflaot/pixels/pkg/http/currency/routes"
	permissionroutes "github.com/niflaot/pixels/pkg/http/permission/routes"
	ws "github.com/niflaot/pixels/pkg/http/websocket"
	"go.uber.org/zap"
)

// New creates the Fiber application without permission administration.
func New(log *zap.Logger, config config.AppConfig, info build.Info, sso *sso.Service, websocket *ws.Adapter, rooms roomservice.Manager, runtime *roomlive.Registry, navigator navservice.Manager, currencyAdmin currencyroutes.Dependencies, catalogAdmin catalogroutes.Dependencies) *fiber.App {
	return newApplication(log, config, info, sso, websocket, rooms, runtime, navigator, currencyAdmin, catalogAdmin, permissionroutes.Dependencies{})
}

// NewWithPermissions creates the complete Fiber application.
func NewWithPermissions(log *zap.Logger, config config.AppConfig, info build.Info, sso *sso.Service, websocket *ws.Adapter, rooms roomservice.Manager, runtime *roomlive.Registry, navigator navservice.Manager, currencyAdmin currencyroutes.Dependencies, catalogAdmin catalogroutes.Dependencies, permissionAdmin permissionroutes.Dependencies) *fiber.App {
	return newApplication(log, config, info, sso, websocket, rooms, runtime, navigator, currencyAdmin, catalogAdmin, permissionAdmin)
}

// newApplication creates the Fiber application with optional permission administration.
func newApplication(log *zap.Logger, config config.AppConfig, info build.Info, sso *sso.Service, websocket *ws.Adapter, rooms roomservice.Manager, runtime *roomlive.Registry, navigator navservice.Manager, currencyAdmin currencyroutes.Dependencies, catalogAdmin catalogroutes.Dependencies, permissionAdmin permissionroutes.Dependencies) *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler:          errorHandler,
	})

	app.Use(fiberzap.New(fiberzap.Config{
		Logger:   log,
		Fields:   []string{"latency", "status", "method", "url", "error"},
		Messages: []string{"http server request failed", "http client request failed", "http request completed"},
	}))

	registerPublic(app, config, info, websocket, currencyAdmin.Currencies, currencyAdmin.Translations)
	app.Use(auth(config.App.AccessKey))
	registerPrivate(app, sso, rooms, runtime, navigator, currencyAdmin, catalogAdmin, permissionAdmin)

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
