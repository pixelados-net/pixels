package http

import (
	"errors"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/pixels/internal/auth/sso"
	"github.com/niflaot/pixels/pkg/build"
	"github.com/niflaot/pixels/pkg/config"
	"github.com/niflaot/pixels/pkg/http/openapi"
)

const development = "development"

// registerPublic registers routes that do not require authentication.
func registerPublic(app *fiber.App, config config.AppConfig, info build.Info) {
	app.Get("/status", statusHandler(config, info))
	app.Get("/ws", websocketGate, websocket.New(wsHandler))
	app.Get("/docs", docsHandler(config))
}

// registerPrivate registers private authenticated fallback routes.
func registerPrivate(app *fiber.App, sso *sso.Service) {
	app.Post("/api/sso/tickets", createSSOTicketHandler(sso))
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
	if websocket.IsWebSocketUpgrade(ctx) {
		return ctx.Next()
	}

	return fiber.ErrUpgradeRequired
}

// wsHandler handles websocket sessions.
func wsHandler(conn *websocket.Conn) {
	_ = conn.Close()
}

// docsHandler returns the Scalar documentation page.
func docsHandler(config config.AppConfig) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		if config.App.Environment != development {
			return notFoundHandler(ctx)
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
			return fiber.ErrBadRequest
		}

		ticket, err := service.Create(ctx.Context(), ssoRequest(request))
		if err != nil {
			if errors.Is(err, sso.ErrInvalidTicket) {
				return fiber.ErrBadRequest
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
		UserID: request.UserID,
		IP:     request.IP,
		TTL:    time.Duration(request.TTLSeconds) * time.Second,
	}
}

// notFoundHandler returns an authenticated not found response.
func notFoundHandler(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusNotFound).JSON(ErrorResponse{
		Error: "not found",
	})
}
