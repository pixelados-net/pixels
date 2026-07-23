// Package routes exposes protected game and poll administration.
package routes

import (
	"context"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	gamecenterdb "github.com/niflaot/pixels/internal/realm/gamecenter/database"
	gamecenterlobby "github.com/niflaot/pixels/internal/realm/gamecenter/lobby"
	gamecenterpolicy "github.com/niflaot/pixels/internal/realm/gamecenter/policy"
	progressiondb "github.com/niflaot/pixels/internal/realm/progression/database"
	progressionpoll "github.com/niflaot/pixels/internal/realm/progression/poll"
	roomgames "github.com/niflaot/pixels/internal/realm/room/world/games"
	"go.uber.org/fx"
)

const basePath = "/api/admin/games"

// Dependencies contains game administration behavior.
type Dependencies struct {
	fx.In
	// Center persists external Game Center registrations.
	Center *gamecenterdb.Repository
	// Lobby owns the enabled client cache.
	Lobby *gamecenterlobby.Service
	// Progression persists polls and shared audit records.
	Progression *progressiondb.Repository
	// Polls owns the enabled room poll cache.
	Polls *progressionpoll.Service
	// Scores reads durable room game history.
	Scores roomgames.ScoreStore
	// Metrics exposes lock-free room game telemetry.
	Metrics *roomgames.Metrics
	// Permissions authorizes administrative actors.
	Permissions permissionservice.Checker
}

// Register mounts every protected game administration route.
func Register(app *fiber.App, dependencies Dependencies) {
	app.Get(basePath+"/center", dependencies.listCenter)
	app.Post(basePath+"/center", dependencies.createCenter)
	app.Patch(basePath+"/center/:id", dependencies.updateCenter)
	app.Delete(basePath+"/center/:id", dependencies.deleteCenter)
	app.Put(basePath+"/center/:id/scores/:playerId", dependencies.putScore)
	app.Get(basePath+"/polls", dependencies.listPolls)
	app.Post(basePath+"/polls", dependencies.createPoll)
	app.Patch(basePath+"/polls/:id", dependencies.updatePoll)
	app.Delete(basePath+"/polls/:id", dependencies.deletePoll)
	app.Put(basePath+"/polls/:id/room", dependencies.assignPoll)
	app.Get(basePath+"/rooms/:roomId/scores", dependencies.listRoomScores)
	app.Post(basePath+"/reload", dependencies.reload)
	app.Get(basePath+"/metrics", dependencies.metrics)
}

// metrics returns one lock-free operational Games snapshot.
func (dependencies Dependencies) metrics(ctx *fiber.Ctx) error {
	if err := dependencies.readActor(ctx, true); err != nil {
		return err
	}
	return ctx.JSON(fiber.Map{"rooms": dependencies.Metrics.Snapshot(), "pollsResponded": dependencies.Polls.DatabaseResponses(), "gameCenterLaunches": dependencies.Lobby.Launches()})
}

// authorize checks one registered game permission.
func (dependencies Dependencies) authorize(ctx *fiber.Ctx, actorID int64, center bool) error {
	if actorID <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid administrative actor")
	}
	node := gamecenterpolicy.ManagePolls
	if center {
		node = gamecenterpolicy.ManageCenter
	}
	allowed, err := dependencies.Permissions.HasPermission(ctx.Context(), actorID, node)
	if err != nil {
		return err
	}
	if !allowed {
		return fiber.NewError(fiber.StatusForbidden, "administrative actor lacks game permission")
	}
	return nil
}

// readActor parses an authorized administrative read header.
func (dependencies Dependencies) readActor(ctx *fiber.Ctx, center bool) error {
	actorID, err := strconv.ParseInt(ctx.Get("X-Actor-Player-ID"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "X-Actor-Player-ID is required")
	}
	return dependencies.authorize(ctx, actorID, center)
}

// mutate validates authorization and writes one audit in the shared transaction.
func (dependencies Dependencies) mutate(ctx *fiber.Ctx, audit AuditRequest, center bool, action string, entity string, work func(context.Context) error) error {
	if audit.ActorPlayerID <= 0 || strings.TrimSpace(audit.Reason) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "actorPlayerId and reason are required")
	}
	if err := dependencies.authorize(ctx, audit.ActorPlayerID, center); err != nil {
		return err
	}
	return dependencies.Progression.WithinTransaction(ctx.Context(), func(txCtx context.Context) error {
		if err := work(txCtx); err != nil {
			return err
		}
		return dependencies.Progression.InsertAudit(txCtx, audit.ActorPlayerID, action, entity, strings.TrimSpace(audit.Reason))
	})
}

// parseID parses one positive path identifier.
func parseID(ctx *fiber.Ctx, name string) (int64, error) {
	value, err := strconv.ParseInt(ctx.Params(name), 10, 64)
	if err != nil || value <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid "+name)
	}
	return value, nil
}

// parseBody parses one JSON request.
func parseBody(ctx *fiber.Ctx, target any) error {
	if err := ctx.BodyParser(target); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid games request")
	}
	return nil
}
