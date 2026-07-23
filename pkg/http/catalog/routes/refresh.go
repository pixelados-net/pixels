package routes

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	outupdated "github.com/niflaot/pixels/networking/outbound/catalog/updated"
	"go.uber.org/zap"
)

// refreshHandler reloads catalog cache and publishes the new generation.
func refreshHandler(dependencies Dependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		if err := dependencies.Catalog.Refresh(ctx.Context()); err != nil {
			return fmt.Errorf("refresh catalog admin cache: %w", err)
		}
		sent, failures, err := publishCatalog(ctx.Context(), dependencies)
		if err != nil {
			return err
		}

		return ctx.JSON(RefreshResponse{Connections: sent, Failures: failures})
	}
}

// publishCatalog tells connected clients to reload the current catalog generation.
func publishCatalog(ctx context.Context, dependencies Dependencies) (int, int, error) {
	packet, err := outupdated.Encode()
	if err != nil {
		return 0, 0, err
	}
	connections := dependencies.Connections.ListAll()
	sent := 0
	failures := 0
	for _, connection := range connections {
		if err := connection.Send(ctx, packet); err != nil {
			failures++
			dependencies.Log.Warn("catalog publication delivery failed", zap.String("connection_id", string(connection.ID())), zap.Error(err))
			continue
		}
		sent++
	}

	return sent, failures, nil
}

// sanitizeListHandler lists furniture definitions without active offers.
func sanitizeListHandler(dependencies Dependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		definitions, err := dependencies.Catalog.SanitizeList(ctx.Context())
		if err != nil {
			return fmt.Errorf("list catalog sanitize definitions: %w", err)
		}
		responses := make([]DefinitionResponse, 0, len(definitions))
		for _, definition := range definitions {
			responses = append(responses, definitionResponse(definition))
		}

		return ctx.JSON(responses)
	}
}
