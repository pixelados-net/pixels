package http

import (
	"errors"
	"strings"
	"time"

	fiberws "github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/pixels/internal/auth/sso"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	navservice "github.com/niflaot/pixels/internal/realm/navigator/core"
	playereffect "github.com/niflaot/pixels/internal/realm/player/effect"
	playerprofile "github.com/niflaot/pixels/internal/realm/player/profile"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	playersettings "github.com/niflaot/pixels/internal/realm/player/settings"
	playerwardrobe "github.com/niflaot/pixels/internal/realm/player/wardrobe"
	roomentry "github.com/niflaot/pixels/internal/realm/room/access/entry"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	roomlayout "github.com/niflaot/pixels/internal/realm/room/world/layout"
	"github.com/niflaot/pixels/pkg/build"
	"github.com/niflaot/pixels/pkg/config"
	botroutes "github.com/niflaot/pixels/pkg/http/bot/routes"
	cameraroutes "github.com/niflaot/pixels/pkg/http/camera/routes"
	catalogroutes "github.com/niflaot/pixels/pkg/http/catalog/routes"
	chatroutes "github.com/niflaot/pixels/pkg/http/chat/routes"
	"github.com/niflaot/pixels/pkg/http/clientconfig"
	craftingroutes "github.com/niflaot/pixels/pkg/http/crafting/routes"
	currencyroutes "github.com/niflaot/pixels/pkg/http/currency/routes"
	gameroutes "github.com/niflaot/pixels/pkg/http/game/routes"
	grouproutes "github.com/niflaot/pixels/pkg/http/group/routes"
	messengerroutes "github.com/niflaot/pixels/pkg/http/messenger/routes"
	moderationroutes "github.com/niflaot/pixels/pkg/http/moderation/routes"
	notificationroutes "github.com/niflaot/pixels/pkg/http/notification/routes"
	"github.com/niflaot/pixels/pkg/http/openapi"
	permissionroutes "github.com/niflaot/pixels/pkg/http/permission/routes"
	petroutes "github.com/niflaot/pixels/pkg/http/pet/routes"
	playerroutes "github.com/niflaot/pixels/pkg/http/player/routes"
	"github.com/niflaot/pixels/pkg/http/pluginroutes"
	progressionroutes "github.com/niflaot/pixels/pkg/http/progression/routes"
	roomroutes "github.com/niflaot/pixels/pkg/http/room/routes"
	subscriptionroutes "github.com/niflaot/pixels/pkg/http/subscription/routes"
	tradingroutes "github.com/niflaot/pixels/pkg/http/trading/routes"
	ws "github.com/niflaot/pixels/pkg/http/websocket"
	wsroutes "github.com/niflaot/pixels/pkg/http/websocket/routes"
	"github.com/niflaot/pixels/pkg/i18n"
	redispkg "github.com/niflaot/pixels/pkg/redis"
	"go.uber.org/zap"
)

const development = "development"

// registerPublic registers routes that do not require authentication.
func registerPublic(app *fiber.App, config config.AppConfig, info build.Info, websocket *ws.Adapter, currencies currencyservice.Reader, layouts roomlayout.Manager, translations i18n.Translator) {
	app.Get("/status", statusHandler(config, info))
	app.Get("/ws", websocketGate, fiberws.New(websocket.Handle))
	app.Get("/docs", docsHandler(config))
	clientconfig.Register(app, currencies, layouts, translations)
}

// registerPrivate registers private authenticated fallback routes.
func registerPrivate(app *fiber.App, sso *sso.Service, redisClient *redispkg.Client, players playerservice.AdminManager, effects playereffect.Manager, settings *playersettings.Service, profile *playerprofile.Service, wardrobe *playerwardrobe.Service, rooms roomservice.Manager, runtime *roomlive.Registry, roomEntry *roomentry.Service, navigator navservice.Manager, currencyAdmin currencyroutes.Dependencies, catalogAdmin catalogroutes.Dependencies, botAdmin botroutes.Dependencies, petAdmin petroutes.Dependencies, groupAdmin grouproutes.Dependencies, craftingAdmin craftingroutes.Dependencies, cameraAdmin cameraroutes.Dependencies, progressionAdmin progressionroutes.Dependencies, gameAdmin gameroutes.Dependencies, permissionAdmin permissionroutes.Dependencies, roomAdmin roomroutes.Dependencies, chatAdmin chatroutes.Dependencies, messengerAdmin messengerroutes.Dependencies, moderationAdmin moderationroutes.Dependencies, subscriptionAdmin subscriptionroutes.Dependencies, tradingAdmin tradingroutes.Dependencies, pluginRoutes *pluginroutes.Registry, log *zap.Logger) {
	app.Post("/api/sso/tickets", createSSOTicketHandler(sso, redisClient))
	playerroutes.Register(app, players, redisClient, currencyAdmin.Players, currencyAdmin.Connections, effects)
	if settings != nil && profile != nil && wardrobe != nil {
		playerroutes.RegisterRemaining(app, playerroutes.RemainingDependencies{Players: players, Settings: settings, Profile: profile, Wardrobe: wardrobe, Live: currencyAdmin.Players})
	}
	wsroutes.Register(app, currencyAdmin.Connections)
	roomroutes.Register(app, rooms, runtime, currencyAdmin.Connections, navigator, currencyAdmin.Players, roomEntry, roomAdmin)
	notificationroutes.Register(app, currencyAdmin.Players, currencyAdmin.Connections, currencyAdmin.Translations)
	currencyroutes.Register(app, currencyAdmin)
	catalogroutes.Register(app, catalogAdmin)
	if botAdmin.Bots != nil {
		botroutes.Register(app, botAdmin)
	}
	if petAdmin.Pets != nil {
		petroutes.Register(app, petAdmin)
	}
	if groupAdmin.Identity != nil {
		grouproutes.Register(app, groupAdmin)
	}
	if craftingAdmin.Store != nil {
		craftingroutes.Register(app, craftingAdmin)
	}
	if cameraAdmin.Camera != nil {
		cameraroutes.Register(app, cameraAdmin)
	}
	if progressionAdmin.Admin != nil {
		progressionroutes.Register(app, progressionAdmin)
	}
	if gameAdmin.Center != nil {
		gameroutes.Register(app, gameAdmin)
	}
	permissionroutes.Register(app, permissionAdmin)
	chatroutes.Register(app, chatAdmin)
	messengerroutes.Register(app, messengerAdmin)
	moderationroutes.Register(app, moderationAdmin)
	subscriptionroutes.Register(app, subscriptionAdmin)
	tradingroutes.Register(app, tradingAdmin)
	if pluginRoutes != nil {
		pluginRoutes.Register(app, log)
	}
	app.Use(notFoundHandler)
}

// statusHandler returns public server status.
func statusHandler(config config.AppConfig, info build.Info) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		return ctx.JSON(StatusResponse{
			Status:      "ok",
			Environment: config.App.Environment,
			Version:     info.Version,
		})
	}
}

// websocketGate requires websocket upgrade requests.
func websocketGate(ctx *fiber.Ctx) error {
	if fiberws.IsWebSocketUpgrade(ctx) {
		return ctx.Next()
	}

	return fiber.NewError(fiber.StatusUpgradeRequired, "websocket upgrade required")
}

// docsHandler returns the Scalar documentation page.
func docsHandler(config config.AppConfig) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		if config.App.Environment != development {
			return fiber.NewError(fiber.StatusNotFound, "documentation is available only in development")
		}

		ctx.Type("html")

		return ctx.SendString(`<!doctype html>
<html>
<head>
  <title>Pixels API</title>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
</head>
<body>
  <script id="api-reference" type="application/json">` + openapi.Spec + `</script>
  <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
</body>
</html>`)
	}
}

// createSSOTicketHandler creates one-time SSO tickets.
func createSSOTicketHandler(service *sso.Service, redisClient *redispkg.Client) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var request CreateSSOTicketRequest
		if err := ctx.BodyParser(&request); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid sso ticket request body")
		}

		idempotencyKey := strings.TrimSpace(ctx.Get(ssoIdempotencyHeader))
		store := ssoIdempotencyStore{client: redisClient}
		if idempotencyKey != "" {
			record, claimed, err := store.claim(ctx.Context(), idempotencyKey, request)
			if err != nil {
				if errors.Is(err, errSSOIdempotencyConflict) || errors.Is(err, errSSOIdempotencyPending) {
					return fiber.NewError(fiber.StatusConflict, err.Error())
				}
				return err
			}
			if !claimed && record.State == "complete" && record.Response != nil {
				ctx.Set(ssoReplayHeader, "true")
				return ctx.Status(fiber.StatusOK).JSON(record.Response)
			}
			if !claimed {
				return fiber.NewError(fiber.StatusConflict, errSSOIdempotencyPending.Error())
			}
		}

		ticket, err := service.Create(ctx.Context(), ssoRequest(request))
		if err != nil {
			if idempotencyKey != "" {
				_ = store.release(ctx.Context(), idempotencyKey)
			}
			if errors.Is(err, sso.ErrInvalidTicket) {
				return fiber.NewError(fiber.StatusBadRequest, "invalid sso ticket request")
			}

			return err
		}

		response := CreateSSOTicketResponse{
			Ticket:    ticket.Value,
			ExpiresAt: ticket.ExpiresAt.UTC().Format(time.RFC3339),
		}
		if idempotencyKey != "" {
			if err := store.complete(ctx.Context(), idempotencyKey, response, time.Until(ticket.ExpiresAt)); err != nil {
				return err
			}
		}
		return ctx.Status(fiber.StatusCreated).JSON(response)
	}
}

// ssoRequest converts an HTTP request into an SSO request.
func ssoRequest(request CreateSSOTicketRequest) sso.CreateRequest {
	return sso.CreateRequest{
		PlayerID: request.PlayerID,
		IP:       request.IP,
		TTL:      time.Duration(request.TTLSeconds) * time.Second,
	}
}

// notFoundHandler returns an authenticated not found response.
func notFoundHandler(ctx *fiber.Ctx) error {
	return fiber.NewError(fiber.StatusNotFound, "route not found")
}
