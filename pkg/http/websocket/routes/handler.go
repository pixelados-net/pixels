package routes

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// listHandler lists active connections.
func listHandler(registry *netconn.Registry) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		connections := listConnections(registry, ctx.Query("kind"))
		items := make([]ConnectionResponse, 0, len(connections))
		for _, connection := range connections {
			items = append(items, connectionResponse(connection))
		}

		return ctx.JSON(ListResponse{Total: len(items), Items: items})
	}
}

// countHandler counts active connections.
func countHandler(registry *netconn.Registry) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		kind := ctx.Query("kind")
		response := CountResponse{Total: registry.CountAll()}
		if kind != "" {
			count := registry.Count(netconn.Kind(kind))
			response.Kind = kind
			response.Count = &count
		}

		return ctx.JSON(response)
	}
}

// reasonsHandler lists supported disconnect reasons.
func reasonsHandler() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		return ctx.JSON(ReasonsResponse{Items: reasonItems()})
	}
}

// disconnectHandler disconnects one connection.
func disconnectHandler(registry *netconn.Registry) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		reason, err := parseDisconnectBody(ctx)
		if err != nil {
			return err
		}

		err = registry.Disconnect(ctx.Context(), netconn.Kind(ctx.Params("kind")), netconn.ID(ctx.Params("id")), reason)
		if errors.Is(err, netconn.ErrConnectionNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "connection not found")
		}
		if err != nil {
			return err
		}

		return ctx.JSON(DisconnectResponse{Matched: 1, Disconnected: 1})
	}
}

// disconnectKindHandler disconnects every connection of one kind.
func disconnectKindHandler(registry *netconn.Registry) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		reason, err := parseDisconnectBody(ctx)
		if err != nil {
			return err
		}

		kind := netconn.Kind(ctx.Params("kind"))
		matched := registry.Count(kind)
		errors := registry.DisconnectKind(ctx.Context(), kind, reason)

		return ctx.JSON(disconnectResponse(matched, errors))
	}
}

// disconnectAllHandler disconnects every registered connection.
func disconnectAllHandler(registry *netconn.Registry) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		reason, err := parseDisconnectBody(ctx)
		if err != nil {
			return err
		}

		matched := registry.CountAll()
		errors := registry.DisconnectAll(ctx.Context(), reason)

		return ctx.JSON(disconnectResponse(matched, errors))
	}
}

// listConnections returns all or kind-filtered connections.
func listConnections(registry *netconn.Registry, kind string) []netconn.Connection {
	if kind != "" {
		return registry.List(netconn.Kind(kind))
	}

	return registry.ListAll()
}

// parseDisconnectBody parses a required disconnect request body.
func parseDisconnectBody(ctx *fiber.Ctx) (netconn.Reason, error) {
	var request DisconnectRequest
	if err := ctx.BodyParser(&request); err != nil {
		return netconn.Reason{}, fiber.NewError(fiber.StatusBadRequest, "invalid disconnect request body")
	}

	return reasonFromRequest(request)
}

// disconnectResponse converts disconnect errors into response counts.
func disconnectResponse(matched int, errors []error) DisconnectResponse {
	failed := len(errors)

	return DisconnectResponse{Matched: matched, Disconnected: matched - failed, Errors: failed}
}
