package routes

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playerprofile "github.com/niflaot/pixels/internal/realm/player/profile"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	playersettings "github.com/niflaot/pixels/internal/realm/player/settings"
	playerwardrobe "github.com/niflaot/pixels/internal/realm/player/wardrobe"
)

// RemainingDependencies contains focused remaining user administration capabilities.
type RemainingDependencies struct {
	// Players updates existing player profile policy.
	Players playerservice.AdminManager
	// Settings reads and updates client settings.
	Settings *playersettings.Service
	// Profile manages public tags.
	Profile *playerprofile.Service
	// Wardrobe reads outfits and clothing unlocks.
	Wardrobe *playerwardrobe.Service
	// Live stores online projections.
	Live *playerlive.Registry
}

// SettingsPatchRequest contains one attributed optimistic settings mutation.
type SettingsPatchRequest struct {
	// ExpectedVersion stores the current settings version.
	ExpectedVersion int64 `json:"expectedVersion"`
	// ActorPlayerID identifies the administrative actor.
	ActorPlayerID int64 `json:"actorPlayerId"`
	// Reason explains the mutation.
	Reason string `json:"reason"`
	// VolumeSystem optionally replaces system volume.
	VolumeSystem *int32 `json:"volumeSystem"`
	// VolumeFurniture optionally replaces furniture volume.
	VolumeFurniture *int32 `json:"volumeFurniture"`
	// VolumeTrax optionally replaces music volume.
	VolumeTrax *int32 `json:"volumeTrax"`
	// OldChat optionally replaces legacy chat selection.
	OldChat *bool `json:"oldChat"`
	// CameraFollowBlocked optionally replaces camera-follow privacy.
	CameraFollowBlocked *bool `json:"cameraFollowBlocked"`
	// SafetyLocked optionally replaces server-controlled safety state.
	SafetyLocked *bool `json:"safetyLocked"`
}

// TagsRequest contains one attributed complete tag replacement.
type TagsRequest struct {
	// Tags stores the complete ordered public tag set.
	Tags []string `json:"tags"`
	// ActorPlayerID identifies the administrative actor.
	ActorPlayerID int64 `json:"actorPlayerId"`
	// Reason explains the mutation.
	Reason string `json:"reason"`
}

// AllowNameChangeRequest contains one attributed rename-policy grant.
type AllowNameChangeRequest struct {
	// ActorPlayerID identifies the administrative actor.
	ActorPlayerID int64 `json:"actorPlayerId"`
	// Reason explains the mutation.
	Reason string `json:"reason"`
}

// WardrobeResponse combines saved outfits and clothing unlocks.
type WardrobeResponse struct {
	// Outfits stores ordered saved slots.
	Outfits []playerwardrobe.Outfit `json:"outfits"`
	// Clothing stores the complete clothing unlock projection.
	Clothing playerwardrobe.ClothingSnapshot `json:"clothing"`
}

// RegisterRemaining mounts protected remaining user administration routes.
func RegisterRemaining(app *fiber.App, dependencies RemainingDependencies) {
	app.Get("/api/admin/players/:id/settings", readSettings(dependencies))
	app.Patch("/api/admin/players/:id/settings", patchSettings(dependencies))
	app.Put("/api/admin/players/:id/profile/tags", replaceTags(dependencies))
	app.Get("/api/admin/players/:id/wardrobe", readWardrobe(dependencies))
	app.Post("/api/admin/players/:id/name-change/allow", allowNameChange(dependencies))
}

// readSettings returns one durable settings snapshot.
func readSettings(dependencies RemainingDependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		id, err := remainingPlayerID(ctx)
		if err != nil {
			return err
		}
		record, err := dependencies.Settings.Find(ctx.Context(), id)
		if err != nil {
			return err
		}
		return ctx.JSON(record)
	}
}

// patchSettings applies one attributed optimistic settings mutation.
func patchSettings(dependencies RemainingDependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		id, err := remainingPlayerID(ctx)
		if err != nil {
			return err
		}
		var request SettingsPatchRequest
		if err = ctx.BodyParser(&request); err != nil || request.ExpectedVersion <= 0 || !attributed(request.ActorPlayerID, request.Reason) {
			return fiber.NewError(fiber.StatusBadRequest, "invalid player settings request")
		}
		record, err := dependencies.Settings.UpdateAdmin(ctx.Context(), id, request.ExpectedVersion, playersettings.AdminPatch{VolumeSystem: request.VolumeSystem, VolumeFurniture: request.VolumeFurniture, VolumeTrax: request.VolumeTrax, OldChat: request.OldChat, CameraFollowBlocked: request.CameraFollowBlocked, SafetyLocked: request.SafetyLocked})
		if err != nil {
			return err
		}
		if dependencies.Live != nil {
			if player, found := dependencies.Live.Find(id); found {
				player.SetClientSettings(record.VolumeSystem, record.VolumeFurniture, record.VolumeTrax, record.OldChat, record.CameraFollowBlocked, record.SafetyLocked)
			}
		}
		return ctx.JSON(record)
	}
}

// replaceTags replaces one player's complete public tag set.
func replaceTags(dependencies RemainingDependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		id, err := remainingPlayerID(ctx)
		if err != nil {
			return err
		}
		var request TagsRequest
		if err = ctx.BodyParser(&request); err != nil || !attributed(request.ActorPlayerID, request.Reason) {
			return fiber.NewError(fiber.StatusBadRequest, "invalid player tags request")
		}
		if err = dependencies.Profile.ReplaceTags(ctx.Context(), id, request.Tags); err != nil {
			return err
		}
		tags, err := dependencies.Profile.Tags(ctx.Context(), id)
		if err != nil {
			return err
		}
		return ctx.JSON(fiber.Map{"playerId": id, "tags": tags})
	}
}

// readWardrobe returns saved outfits and clothing unlocks.
func readWardrobe(dependencies RemainingDependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		id, err := remainingPlayerID(ctx)
		if err != nil {
			return err
		}
		outfits, err := dependencies.Wardrobe.Outfits(ctx.Context(), id)
		if err != nil {
			return err
		}
		clothing, err := dependencies.Wardrobe.Clothing(ctx.Context(), id)
		if err != nil {
			return err
		}
		return ctx.JSON(WardrobeResponse{Outfits: outfits, Clothing: clothing})
	}
}

// allowNameChange enables one one-shot username change through the player aggregate.
func allowNameChange(dependencies RemainingDependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		id, err := remainingPlayerID(ctx)
		if err != nil {
			return err
		}
		var request AllowNameChangeRequest
		if err = ctx.BodyParser(&request); err != nil || !attributed(request.ActorPlayerID, request.Reason) {
			return fiber.NewError(fiber.StatusBadRequest, "invalid name-change policy request")
		}
		allowed := true
		record, err := dependencies.Players.Update(ctx.Context(), id, playerservice.UpdateParams{AllowNameChange: &allowed})
		if err != nil {
			return playerError(err)
		}
		if dependencies.Live != nil {
			if player, found := dependencies.Live.Find(id); found {
				player.SetUsername(record.Player.Username, true)
			}
		}
		return ctx.JSON(playerResponse(record))
	}
}

// remainingPlayerID parses one protected player id path parameter.
func remainingPlayerID(ctx *fiber.Ctx) (int64, error) {
	id, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid player id")
	}
	return id, nil
}

// attributed validates administrative actor and reason fields.
func attributed(actorPlayerID int64, reason string) bool {
	return actorPlayerID > 0 && strings.TrimSpace(reason) != ""
}
