package http

import (
	"errors"
	"time"

	fiberws "github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/pixels/internal/auth/sso"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/build"
	"github.com/niflaot/pixels/pkg/config"
	"github.com/niflaot/pixels/pkg/http/openapi"
	ws "github.com/niflaot/pixels/pkg/http/websocket"
	wsroutes "github.com/niflaot/pixels/pkg/http/websocket/routes"
)

const development = "development"

// registerPublic registers routes that do not require authentication.
func registerPublic(app *fiber.App, config config.AppConfig, info build.Info, websocket *ws.Adapter) {
	app.Get("/status", statusHandler(config, info))
	app.Get("/ws", websocketGate, fiberws.New(websocket.Handle))
	app.Get("/docs", docsHandler(config))
}

// registerPrivate registers private authenticated fallback routes.
func registerPrivate(app *fiber.App, sso *sso.Service, registry *netconn.Registry) {
	app.Post("/api/sso/tickets", createSSOTicketHandler(sso))
	wsroutes.Register(app, registry)
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
func createSSOTicketHandler(service *sso.Service) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var request CreateSSOTicketRequest
		if err := ctx.BodyParser(&request); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid sso ticket request body")
		}

		ticket, err := service.Create(ctx.Context(), ssoRequest(request))
		if err != nil {
			if errors.Is(err, sso.ErrInvalidTicket) {
				return fiber.NewError(fiber.StatusBadRequest, "invalid sso ticket request")
			}

			return err
		}

		return ctx.Status(fiber.StatusCreated).JSON(CreateSSOTicketResponse{
			Ticket:    ticket.Value,
			ExpiresAt: ticket.ExpiresAt.UTC().Format(time.RFC3339),
		})
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
