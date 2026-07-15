package http

import (
	"github.com/gofiber/contrib/fiberzap"
	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/pixels/internal/auth/sso"
	navservice "github.com/niflaot/pixels/internal/realm/navigator/core"
	playereffect "github.com/niflaot/pixels/internal/realm/player/effect"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	roomentry "github.com/niflaot/pixels/internal/realm/room/access/entry"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	roomlayout "github.com/niflaot/pixels/internal/realm/room/world/layout"
	"github.com/niflaot/pixels/pkg/build"
	"github.com/niflaot/pixels/pkg/config"
	botroutes "github.com/niflaot/pixels/pkg/http/bot/routes"
	catalogroutes "github.com/niflaot/pixels/pkg/http/catalog/routes"
	chatroutes "github.com/niflaot/pixels/pkg/http/chat/routes"
	currencyroutes "github.com/niflaot/pixels/pkg/http/currency/routes"
	messengerroutes "github.com/niflaot/pixels/pkg/http/messenger/routes"
	moderationroutes "github.com/niflaot/pixels/pkg/http/moderation/routes"
	permissionroutes "github.com/niflaot/pixels/pkg/http/permission/routes"
	roomroutes "github.com/niflaot/pixels/pkg/http/room/routes"
	subscriptionroutes "github.com/niflaot/pixels/pkg/http/subscription/routes"
	tradingroutes "github.com/niflaot/pixels/pkg/http/trading/routes"
	ws "github.com/niflaot/pixels/pkg/http/websocket"
	redispkg "github.com/niflaot/pixels/pkg/redis"
	"go.uber.org/zap"
)

// New creates the Fiber application without permission administration.
func New(log *zap.Logger, config config.AppConfig, info build.Info, sso *sso.Service, redisClient *redispkg.Client, players playerservice.AdminManager, websocket *ws.Adapter, rooms roomservice.Manager, layouts roomlayout.Manager, runtime *roomlive.Registry, roomEntry *roomentry.Service, navigator navservice.Manager, currencyAdmin currencyroutes.Dependencies, catalogAdmin catalogroutes.Dependencies) *fiber.App {
	return newApplication(log, config, info, sso, redisClient, players, nil, websocket, rooms, layouts, runtime, roomEntry, navigator, currencyAdmin, catalogAdmin, botroutes.Dependencies{}, permissionroutes.Dependencies{}, roomroutes.Dependencies{}, chatroutes.Dependencies{}, messengerroutes.Dependencies{}, moderationroutes.Dependencies{}, subscriptionroutes.Dependencies{}, tradingroutes.Dependencies{})
}

// NewWithPermissions creates the complete Fiber application.
func NewWithPermissions(log *zap.Logger, config config.AppConfig, info build.Info, sso *sso.Service, redisClient *redispkg.Client, players playerservice.AdminManager, effects playereffect.Manager, websocket *ws.Adapter, rooms roomservice.Manager, layouts roomlayout.Manager, runtime *roomlive.Registry, roomEntry *roomentry.Service, navigator navservice.Manager, currencyAdmin currencyroutes.Dependencies, catalogAdmin catalogroutes.Dependencies, botAdmin botroutes.Dependencies, permissionAdmin permissionroutes.Dependencies, roomAdmin roomroutes.Dependencies, chatAdmin chatroutes.Dependencies, messengerAdmin messengerroutes.Dependencies, moderationAdmin moderationroutes.Dependencies, subscriptionAdmin subscriptionroutes.Dependencies, tradingAdmin tradingroutes.Dependencies) *fiber.App {
	return newApplication(log, config, info, sso, redisClient, players, effects, websocket, rooms, layouts, runtime, roomEntry, navigator, currencyAdmin, catalogAdmin, botAdmin, permissionAdmin, roomAdmin, chatAdmin, messengerAdmin, moderationAdmin, subscriptionAdmin, tradingAdmin)
}

// newApplication creates the Fiber application with optional permission administration.
func newApplication(log *zap.Logger, config config.AppConfig, info build.Info, sso *sso.Service, redisClient *redispkg.Client, players playerservice.AdminManager, effects playereffect.Manager, websocket *ws.Adapter, rooms roomservice.Manager, layouts roomlayout.Manager, runtime *roomlive.Registry, roomEntry *roomentry.Service, navigator navservice.Manager, currencyAdmin currencyroutes.Dependencies, catalogAdmin catalogroutes.Dependencies, botAdmin botroutes.Dependencies, permissionAdmin permissionroutes.Dependencies, roomAdmin roomroutes.Dependencies, chatAdmin chatroutes.Dependencies, messengerAdmin messengerroutes.Dependencies, moderationAdmin moderationroutes.Dependencies, subscriptionAdmin subscriptionroutes.Dependencies, tradingAdmin tradingroutes.Dependencies) *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler:          errorHandler,
	})

	app.Use(fiberzap.New(fiberzap.Config{
		Logger:   log,
		Fields:   []string{"latency", "status", "method", "url", "error"},
		Messages: []string{"http server request failed", "http client request failed", "http request completed"},
	}))

	registerPublic(app, config, info, websocket, currencyAdmin.Currencies, layouts, currencyAdmin.Translations)
	app.Use(auth(config.App.AccessKey))
	registerPrivate(app, sso, redisClient, players, effects, rooms, runtime, roomEntry, navigator, currencyAdmin, catalogAdmin, botAdmin, permissionAdmin, roomAdmin, chatAdmin, messengerAdmin, moderationAdmin, subscriptionAdmin, tradingAdmin)

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
