package routes

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/pixels/internal/permission"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	cameraadmin "github.com/niflaot/pixels/internal/realm/camera/admin"
	camerapolicy "github.com/niflaot/pixels/internal/realm/camera/policy"
	camerarecord "github.com/niflaot/pixels/internal/realm/camera/record"
	"go.uber.org/fx"
)

// basePath stores the protected camera route prefix.
const basePath = "/api/admin/camera"

// Dependencies contains protected camera administration behavior.
type Dependencies struct {
	fx.In
	// Camera coordinates protected camera workflows.
	Camera *cameraadmin.Service
	// Permissions authorizes administrative actors.
	Permissions permissionservice.Checker
}

// Register mounts every protected camera administration route.
func Register(app *fiber.App, dependencies Dependencies) {
	app.Get(basePath+"/settings", dependencies.getSettings)
	app.Put(basePath+"/settings", dependencies.putSettings)
	app.Get(basePath+"/captures/:playerId", dependencies.captures)
	app.Get(basePath+"/gallery", dependencies.gallery)
	app.Delete(basePath+"/gallery/:publicationId", dependencies.removePublication)
	app.Delete(basePath+"/photos/:itemId", dependencies.deletePhoto)
}

// getSettings returns current pricing and policy.
func (dependencies Dependencies) getSettings(ctx *fiber.Ctx) error {
	actorID, err := dependencies.actor(ctx)
	if err != nil {
		return err
	}
	if err = dependencies.authorize(ctx, actorID, camerapolicy.SettingsManage); err != nil {
		return err
	}
	settings, err := dependencies.Camera.Settings(ctx.Context())
	if err != nil {
		return err
	}
	return ctx.JSON(settingsResponse(settings))
}

// putSettings replaces current pricing and policy.
func (dependencies Dependencies) putSettings(ctx *fiber.Ctx) error {
	var request SettingsRequest
	if err := ctx.BodyParser(&request); err != nil || !validSettings(request) {
		return fiber.NewError(fiber.StatusBadRequest, "invalid camera settings request")
	}
	audit := camerarecord.Audit{ActorPlayerID: request.ActorPlayerID, Reason: request.Reason}
	if !cameraadmin.ValidAudit(audit) {
		return fiber.NewError(fiber.StatusBadRequest, "actorPlayerId and reason are required")
	}
	if err := dependencies.authorize(ctx, request.ActorPlayerID, camerapolicy.SettingsManage); err != nil {
		return err
	}
	settings, err := dependencies.Camera.UpdateSettings(ctx.Context(), settingsRecord(request), request.Version, audit)
	if errors.Is(err, cameraadmin.ErrConflict) {
		return fiber.ErrConflict
	}
	if err != nil {
		return err
	}
	return ctx.JSON(settingsResponse(settings))
}

// captures lists recent captures for one player.
func (dependencies Dependencies) captures(ctx *fiber.Ctx) error {
	actorID, err := dependencies.actor(ctx)
	if err != nil {
		return err
	}
	if err = dependencies.authorize(ctx, actorID, camerapolicy.GalleryModerate); err != nil {
		return err
	}
	playerID, err := pathID(ctx, "playerId")
	if err != nil {
		return err
	}
	values, err := dependencies.Camera.Captures(ctx.Context(), playerID, queryInt(ctx, "limit", 50))
	if err != nil {
		return err
	}
	return ctx.JSON(values)
}

// gallery lists public camera publications.
func (dependencies Dependencies) gallery(ctx *fiber.Ctx) error {
	actorID, err := dependencies.actor(ctx)
	if err != nil {
		return err
	}
	if err = dependencies.authorize(ctx, actorID, camerapolicy.GalleryModerate); err != nil {
		return err
	}
	values, err := dependencies.Camera.Publications(ctx.Context(), queryInt(ctx, "limit", 50), queryInt(ctx, "offset", 0), ctx.QueryBool("includeRemoved", false))
	if err != nil {
		return err
	}
	return ctx.JSON(values)
}

// removePublication soft-removes one gallery entry.
func (dependencies Dependencies) removePublication(ctx *fiber.Ctx) error {
	return dependencies.remove(ctx, "publicationId", true)
}

// deletePhoto removes one photo furniture item.
func (dependencies Dependencies) deletePhoto(ctx *fiber.Ctx) error {
	return dependencies.remove(ctx, "itemId", false)
}

// remove executes one audited gallery moderation mutation.
func (dependencies Dependencies) remove(ctx *fiber.Ctx, parameter string, publication bool) error {
	id, err := pathID(ctx, parameter)
	if err != nil {
		return err
	}
	var request AuditRequest
	if err = ctx.BodyParser(&request); err != nil {
		return fiber.ErrBadRequest
	}
	audit := camerarecord.Audit{ActorPlayerID: request.ActorPlayerID, Reason: request.Reason}
	if !cameraadmin.ValidAudit(audit) {
		return fiber.NewError(fiber.StatusBadRequest, "actorPlayerId and reason are required")
	}
	if err = dependencies.authorize(ctx, request.ActorPlayerID, camerapolicy.GalleryModerate); err != nil {
		return err
	}
	removed := false
	if publication {
		removed, err = dependencies.Camera.RemovePublication(ctx.Context(), id, audit)
	} else {
		removed, err = dependencies.Camera.DeletePhoto(ctx.Context(), id, audit)
	}
	if err != nil {
		return err
	}
	if !removed {
		return fiber.ErrNotFound
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

// authorize checks one camera administration node.
func (dependencies Dependencies) authorize(ctx *fiber.Ctx, actorID int64, node permission.Node) error {
	allowed, err := dependencies.Permissions.HasPermission(ctx.Context(), actorID, node)
	if err != nil {
		return err
	}
	if !allowed {
		return fiber.NewError(fiber.StatusForbidden, "administrative actor lacks camera permission")
	}
	return nil
}

// actor reads the administrative actor header.
func (dependencies Dependencies) actor(ctx *fiber.Ctx) (int64, error) {
	value, err := strconv.ParseInt(ctx.Get("X-Actor-Player-ID"), 10, 64)
	if err != nil || value <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid X-Actor-Player-ID")
	}
	return value, nil
}

// pathID parses one positive path identifier.
func pathID(ctx *fiber.Ctx, name string) (int64, error) {
	value, err := strconv.ParseInt(ctx.Params(name), 10, 64)
	if err != nil || value <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid "+name)
	}
	return value, nil
}

// queryInt parses one optional integer query parameter.
func queryInt(ctx *fiber.Ctx, name string, fallback int) int {
	value, err := strconv.Atoi(ctx.Query(name))
	if err != nil {
		return fallback
	}
	return value
}

// validSettings validates operational settings bounds.
func validSettings(request SettingsRequest) bool {
	return request.Version > 0 && request.CreditsPrice >= 0 && request.PointsPrice >= 0 && request.PublishPointsPrice >= 0 && request.PublishCooldownSeconds >= 0
}
