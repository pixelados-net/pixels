// Package routes exposes protected social-group administration.
package routes

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/pixels/internal/permission"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	"github.com/niflaot/pixels/internal/realm/group/forum"
	"github.com/niflaot/pixels/internal/realm/group/identity"
	"github.com/niflaot/pixels/internal/realm/group/membership"
	groupobservability "github.com/niflaot/pixels/internal/realm/group/observability"
	grouppolicy "github.com/niflaot/pixels/internal/realm/group/policy"
	"go.uber.org/fx"
)

const basePath = "/api/admin/groups"
const actorHeader = "X-Actor-Player-ID"

// Dependencies contains social-group administration behavior.
type Dependencies struct {
	fx.In

	// Identity manages group identity and lifecycle.
	Identity *identity.Service
	// Membership manages group rosters and favorites.
	Membership *membership.Service
	// Forum manages retained forum state.
	Forum *forum.Service
	// Permissions authorizes the attributed administrative actor.
	Permissions permissionservice.Checker
	// Metrics exposes bounded process-wide group telemetry.
	Metrics *groupobservability.Metrics
}

// Register mounts protected social-group administration routes.
func Register(app *fiber.App, dependencies Dependencies) {
	app.Get(basePath, dependencies.list)
	app.Post(basePath, dependencies.create)
	app.Get(basePath+"/metrics", dependencies.metrics)
	app.Get(basePath+"/players/:playerId", dependencies.playerGroups)
	app.Put(basePath+"/players/:playerId/favorite", dependencies.favorite)
	app.Get(basePath+"/:groupId", dependencies.read)
	app.Patch(basePath+"/:groupId", dependencies.update)
	app.Delete(basePath+"/:groupId", dependencies.deactivate)
	app.Post(basePath+"/:groupId/restore", dependencies.restore)
	app.Put(basePath+"/:groupId/badge", dependencies.badge)
	app.Put(basePath+"/:groupId/home-room", dependencies.homeRoom)
	app.Post(basePath+"/:groupId/owner", dependencies.owner)
	app.Get(basePath+"/:groupId/members", dependencies.members)
	app.Post(basePath+"/:groupId/members", dependencies.addMember)
	app.Patch(basePath+"/:groupId/members/:playerId", dependencies.role)
	app.Delete(basePath+"/:groupId/members/:playerId", dependencies.removeMember)
	app.Get(basePath+"/:groupId/requests", dependencies.requests)
	app.Post(basePath+"/:groupId/requests/:playerId/accept", dependencies.accept)
	app.Delete(basePath+"/:groupId/requests/:playerId", dependencies.decline)
	app.Post(basePath+"/:groupId/requests/approve-all", dependencies.approveAll)
	app.Get(basePath+"/:groupId/forum/settings", dependencies.forumSettings)
	app.Put(basePath+"/:groupId/forum/settings", dependencies.updateForumSettings)
	app.Get(basePath+"/:groupId/forum/threads", dependencies.threads)
	app.Get(basePath+"/:groupId/forum/threads/:threadId", dependencies.thread)
	app.Patch(basePath+"/:groupId/forum/threads/:threadId", dependencies.updateThread)
	app.Patch(basePath+"/:groupId/forum/posts/:postId", dependencies.updatePost)
}

// metrics returns bounded process-wide group telemetry.
func (dependencies Dependencies) metrics(ctx *fiber.Ctx) error {
	actorID, err := readActor(ctx)
	if err != nil {
		return err
	}
	if err = dependencies.authorize(ctx, actorID, grouppolicy.ManageAny); err != nil {
		return err
	}
	return ctx.JSON(dependencies.Metrics.Snapshot())
}

// authorize verifies one positive administrative actor and permission node.
func (dependencies Dependencies) authorize(ctx *fiber.Ctx, actorID int64, node permission.Node) error {
	if actorID <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid administrative actor")
	}
	allowed, err := dependencies.Permissions.HasPermission(ctx.Context(), actorID, node)
	if err != nil {
		return err
	}
	if !allowed {
		return fiber.NewError(fiber.StatusForbidden, "administrative actor lacks group permission")
	}
	return nil
}

// readActor parses the required read attribution header.
func readActor(ctx *fiber.Ctx) (int64, error) {
	value, err := strconv.ParseInt(ctx.Get(actorHeader), 10, 64)
	if err != nil || value <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid administrative actor header")
	}
	return value, nil
}
