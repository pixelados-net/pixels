// Package routes exposes protected progression administration.
package routes

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/niflaot/pixels/internal/permission"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	playerachievement "github.com/niflaot/pixels/internal/realm/player/achievement"
	progressionengine "github.com/niflaot/pixels/internal/realm/progression/engine"
	progressionobservability "github.com/niflaot/pixels/internal/realm/progression/observability"
	progressionpolicy "github.com/niflaot/pixels/internal/realm/progression/policy"
	progressionpoll "github.com/niflaot/pixels/internal/realm/progression/poll"
	progressionpromo "github.com/niflaot/pixels/internal/realm/progression/promo"
	progressionquest "github.com/niflaot/pixels/internal/realm/progression/quest"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
	progressiontalent "github.com/niflaot/pixels/internal/realm/progression/talent"
	progressionrequest "github.com/niflaot/pixels/pkg/http/progression/routes/request"
	"go.uber.org/fx"
)

const basePath = "/api/admin/progression"

// Dependencies contains protected progression administration behavior.
type Dependencies struct {
	fx.In
	// Admin persists catalog mutations and audit records.
	Admin progressionrecord.AdminStore
	// Store reads player progression.
	Store progressionrecord.Store
	// Catalog owns the live immutable generation.
	Catalog *progressionengine.Catalog
	// Engine applies real achievement transitions.
	Engine *progressionengine.Service
	// Quests applies real quest transitions.
	Quests *progressionquest.Service
	// Talents applies real talent transitions.
	Talents *progressiontalent.Service
	// Promos applies real promotional claims.
	Promos *progressionpromo.Service
	// Polls owns live room word quizzes.
	Polls *progressionpoll.Service
	// Badges owns direct player badge changes.
	Badges *playerachievement.Service
	// Permissions authorizes administrative actors.
	Permissions permissionservice.Checker
	// Metrics stores lock-free operational telemetry.
	Metrics *progressionobservability.Metrics
}

// AuditRequest aliases the shared administrative attribution payload.
type AuditRequest = progressionrequest.Audit

// Register mounts every protected progression administration route.
func Register(app *fiber.App, dependencies Dependencies) {
	registerAchievements(app, dependencies)
	registerPlayers(app, dependencies)
	registerQuests(app, dependencies)
	registerQuizzes(app, dependencies)
	registerPromos(app, dependencies)
	app.Post(basePath+"/reload", dependencies.reload)
	app.Get(basePath+"/metrics", dependencies.metrics)
}

// metrics returns one lock-free operational snapshot.
func (dependencies Dependencies) metrics(ctx *fiber.Ctx) error {
	if err := dependencies.readActor(ctx, progressionpolicy.ManageDefinitions); err != nil {
		return err
	}
	return ctx.JSON(dependencies.Metrics.Snapshot())
}

// reload atomically swaps the immutable catalog generation.
func (dependencies Dependencies) reload(ctx *fiber.Ctx) error {
	var request AuditRequest
	if err := parseBody(ctx, &request); err != nil {
		return err
	}
	if err := dependencies.authorize(ctx, request.ActorPlayerID, progressionpolicy.ManageDefinitions); err != nil {
		return err
	}
	if err := dependencies.Catalog.Reload(ctx.Context()); err != nil {
		return err
	}
	if err := dependencies.Admin.WithinTransaction(ctx.Context(), func(txCtx context.Context) error {
		return dependencies.Admin.InsertAudit(txCtx, request.ActorPlayerID, "catalog.reload", "progression", request.Reason)
	}); err != nil {
		return err
	}
	return ctx.JSON(fiber.Map{"reloaded": true})
}

// mutate authorizes and atomically executes one audited mutation.
func (dependencies Dependencies) mutate(ctx *fiber.Ctx, node permission.Node, audit AuditRequest, action string, entity string, work func(context.Context) error) error {
	if err := validateAudit(audit); err != nil {
		return err
	}
	if err := dependencies.authorize(ctx, audit.ActorPlayerID, node); err != nil {
		return err
	}
	err := dependencies.Admin.WithinTransaction(ctx.Context(), func(txCtx context.Context) error {
		if err := work(txCtx); err != nil {
			return err
		}
		return dependencies.Admin.InsertAudit(txCtx, audit.ActorPlayerID, action, entity, strings.TrimSpace(audit.Reason))
	})
	return progressionError(err)
}

// authorize checks one registered progression permission.
func (dependencies Dependencies) authorize(ctx *fiber.Ctx, actorID int64, node permission.Node) error {
	if actorID <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid administrative actor")
	}
	allowed, err := dependencies.Permissions.HasPermission(ctx.Context(), actorID, node)
	if err != nil {
		return err
	}
	if !allowed {
		return fiber.NewError(fiber.StatusForbidden, "administrative actor lacks progression permission")
	}
	return nil
}

// readActor parses and authorizes the administrative read header.
func (dependencies Dependencies) readActor(ctx *fiber.Ctx, node permission.Node) error {
	value, err := strconv.ParseInt(ctx.Get("X-Actor-Player-ID"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "X-Actor-Player-ID is required")
	}
	return dependencies.authorize(ctx, value, node)
}

// parseBody parses one JSON request body.
func parseBody(ctx *fiber.Ctx, target any) error {
	if err := ctx.BodyParser(target); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid progression request")
	}
	return nil
}

// validateAudit validates administrative attribution.
func validateAudit(value AuditRequest) error {
	if value.ActorPlayerID <= 0 || strings.TrimSpace(value.Reason) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "actorPlayerId and reason are required")
	}
	return nil
}

// parsePositiveID parses one positive path identifier.
func parsePositiveID(ctx *fiber.Ctx, name string) (int64, error) {
	value, err := strconv.ParseInt(ctx.Params(name), 10, 64)
	if err != nil || value <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid "+name)
	}
	return value, nil
}

// parsePositiveLevel parses one positive path level.
func parsePositiveLevel(ctx *fiber.Ctx) (int32, error) {
	value, err := strconv.ParseInt(ctx.Params("level"), 10, 32)
	if err != nil || value <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid level")
	}
	return int32(value), nil
}

// progressionError maps domain and PostgreSQL failures to stable HTTP errors.
func progressionError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, progressionrecord.ErrNotFound):
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	case errors.Is(err, progressionrecord.ErrConflict):
		return fiber.NewError(fiber.StatusConflict, err.Error())
	case errors.Is(err, progressionrecord.ErrInvalid), errors.Is(err, progressionrecord.ErrUnavailable):
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		switch postgresError.Code {
		case "23505":
			return fiber.NewError(fiber.StatusConflict, postgresError.Message)
		case "23503", "23514", "22001", "22P02":
			return fiber.NewError(fiber.StatusUnprocessableEntity, postgresError.Message)
		}
	}
	return err
}

// entityID formats a stable audited entity identifier.
func entityID(kind string, value any) string { return fmt.Sprintf("%s:%v", kind, value) }
