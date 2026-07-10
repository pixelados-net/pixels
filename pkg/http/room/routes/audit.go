package routes

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	roomaudit "github.com/niflaot/pixels/internal/realm/room/audit"
	roommoderation "github.com/niflaot/pixels/internal/realm/room/moderation"
	moderationmodel "github.com/niflaot/pixels/internal/realm/room/moderation/model"
)

const (
	// playersPath stores the player administration base path.
	playersPath = "/api/admin/players"
)

// AuditResponse contains a paginated audit or sanction page.
type AuditResponse struct {
	// Total stores the returned row count.
	Total int `json:"total"`
	// Items stores typed rows.
	Items any `json:"items"`
}

// registerAudit mounts room audit and sanction routes.
func registerAudit(app *fiber.App, dependencies Dependencies) {
	if dependencies.Audit == nil || dependencies.Moderation == nil {
		return
	}
	app.Get(roomPath+"/:id/rights/history", rightsHistoryHandler(dependencies.Audit))
	app.Get(roomPath+"/:id/moderation/history", moderationHistoryHandler(dependencies.Audit))
	app.Get(roomPath+"/:id/bans", bansHandler(dependencies.Moderation))
	app.Get(roomPath+"/:id/mutes", mutesHandler(dependencies.Moderation))
	app.Get(playersPath+"/:playerId/moderation/history", playerHistoryHandler(dependencies.Audit, false))
	app.Get(playersPath+"/:playerId/moderation/actions", playerHistoryHandler(dependencies.Audit, true))
}

// rightsHistoryHandler returns rights history for one room.
func rightsHistoryHandler(manager roomaudit.Manager) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		roomID, query, err := roomAuditQuery(ctx)
		if err != nil {
			return err
		}
		query.RoomID = &roomID
		items, err := manager.RightsHistory(ctx.Context(), query)
		if err != nil {
			return auditError(err)
		}

		return ctx.JSON(AuditResponse{Total: len(items), Items: items})
	}
}

// moderationHistoryHandler returns moderation history for one room.
func moderationHistoryHandler(manager roomaudit.Manager) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		roomID, query, err := roomAuditQuery(ctx)
		if err != nil {
			return err
		}
		query.RoomID = &roomID
		items, err := manager.ModerationHistory(ctx.Context(), query)
		if err != nil {
			return auditError(err)
		}

		return ctx.JSON(AuditResponse{Total: len(items), Items: items})
	}
}

// bansHandler returns active room bans.
func bansHandler(reader roommoderation.Reader) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		roomID, err := roomIDParam(ctx)
		if err != nil {
			return err
		}
		items, err := reader.ListBans(ctx.Context(), roomID)
		if err != nil {
			return err
		}

		return ctx.JSON(AuditResponse{Total: len(items), Items: items})
	}
}

// mutesHandler returns active room mutes.
func mutesHandler(reader roommoderation.Reader) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		roomID, err := roomIDParam(ctx)
		if err != nil {
			return err
		}
		items, err := reader.ListMutes(ctx.Context(), roomID)
		if err != nil {
			return err
		}

		return ctx.JSON(AuditResponse{Total: len(items), Items: items})
	}
}

// playerHistoryHandler returns target or actor history for one player.
func playerHistoryHandler(manager roomaudit.Manager, actor bool) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		playerID, err := positiveParam(ctx, "playerId", "invalid player id")
		if err != nil {
			return err
		}
		query, err := auditQuery(ctx)
		if err != nil {
			return err
		}
		if actor {
			query.ActorPlayerID = &playerID
		} else {
			query.TargetPlayerID = &playerID
		}
		items, err := manager.ModerationHistory(ctx.Context(), query)
		if err != nil {
			return auditError(err)
		}

		return ctx.JSON(AuditResponse{Total: len(items), Items: items})
	}
}

// auditError maps invalid filters while preserving infrastructure failures.
func auditError(err error) error {
	if errors.Is(err, roomaudit.ErrInvalidQuery) {
		return fiber.NewError(fiber.StatusBadRequest, "invalid room audit query")
	}

	return err
}

// roomAuditQuery parses a room id and audit filters.
func roomAuditQuery(ctx *fiber.Ctx) (int64, roomaudit.Query, error) {
	roomID, err := roomIDParam(ctx)
	if err != nil {
		return 0, roomaudit.Query{}, err
	}
	query, err := auditQuery(ctx)

	return roomID, query, err
}

// auditQuery parses indexed audit query filters.
func auditQuery(ctx *fiber.Ctx) (roomaudit.Query, error) {
	query := roomaudit.Query{Limit: ctx.QueryInt("limit", 50)}
	if value := ctx.Query("before"); value != "" {
		parsed, err := positiveInt(value, "invalid audit cursor")
		if err != nil {
			return roomaudit.Query{}, err
		}
		query.Before = &parsed
	}
	if value := ctx.Query("roomId"); value != "" {
		parsed, err := positiveInt(value, "invalid room id filter")
		if err != nil {
			return roomaudit.Query{}, err
		}
		query.RoomID = &parsed
	}
	for _, value := range strings.Split(ctx.Query("type"), ",") {
		if value != "" {
			query.ActionTypes = append(query.ActionTypes, moderationmodel.Action(value))
		}
	}

	return query, nil
}

// positiveParam parses one positive path id.
func positiveParam(ctx *fiber.Ctx, name string, message string) (int64, error) {
	return positiveInt(ctx.Params(name), message)
}

// positiveInt parses one positive decimal integer.
func positiveInt(value string, message string) (int64, error) {
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, message)
	}

	return parsed, nil
}
