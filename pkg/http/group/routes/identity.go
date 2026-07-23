package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/pixels/internal/permission"
	"github.com/niflaot/pixels/internal/realm/group/identity"
	grouppolicy "github.com/niflaot/pixels/internal/realm/group/policy"
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
)

// list returns one bounded administration group page.
func (dependencies Dependencies) list(ctx *fiber.Ctx) error {
	actorID, err := readActor(ctx)
	if err != nil {
		return err
	}
	if err = dependencies.authorize(ctx, actorID, grouppolicy.ManageAny); err != nil {
		return err
	}
	offset, err := boundedInt(ctx.Query("offset"), 0, 1000000)
	if err != nil {
		return err
	}
	limit, err := boundedInt(ctx.Query("limit"), 50, 200)
	if err != nil || limit < 1 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid group list limit")
	}
	ownerID, err := optionalPositive(ctx.Query("ownerPlayerId"))
	if err != nil {
		return err
	}
	roomID, err := optionalPositive(ctx.Query("homeRoomId"))
	if err != nil {
		return err
	}
	var state *grouprecord.State
	if value := ctx.Query("state"); value != "" {
		parsed, parseErr := boundedInt(value, 0, 2)
		if parseErr != nil {
			return parseErr
		}
		resolved := grouprecord.State(parsed)
		state = &resolved
	}
	forumEnabled, err := optionalBool(ctx.Query("forumEnabled"))
	if err != nil {
		return err
	}
	active, err := optionalBool(ctx.Query("active"))
	if err != nil {
		return err
	}
	items, err := dependencies.Identity.Groups(ctx.Context(), grouprecord.GroupFilter{Query: ctx.Query("query"), OwnerPlayerID: ownerID, HomeRoomID: roomID, State: state, ForumEnabled: forumEnabled, Active: active, Offset: offset, Limit: limit})
	if err != nil {
		return groupError(err)
	}
	next := 0
	if len(items) == limit {
		next = offset + limit
	}
	return ctx.JSON(GroupListResponse{Items: items, NextOffset: next})
}

// read returns identity, settings, counts, badge parts, and version.
func (dependencies Dependencies) read(ctx *fiber.Ctx) error {
	actorID, err := readActor(ctx)
	if err != nil {
		return err
	}
	if err = dependencies.authorize(ctx, actorID, grouppolicy.ManageAny); err != nil {
		return err
	}
	groupID, err := positiveParam(ctx, "groupId")
	if err != nil {
		return err
	}
	group, found, err := dependencies.Identity.Group(ctx.Context(), groupID, ctx.QueryBool("deactivated", true))
	if err != nil {
		return groupError(err)
	}
	if !found {
		return fiber.NewError(fiber.StatusNotFound, grouprecord.ErrNotFound.Error())
	}
	badgeParts, err := dependencies.Identity.Parts(ctx.Context(), groupID)
	if err != nil {
		return groupError(err)
	}
	return ctx.JSON(GroupResponse{Group: group, BadgeParts: badgeParts})
}

// create validates and atomically purchases an administrative group.
func (dependencies Dependencies) create(ctx *fiber.Ctx) error {
	request := CreateRequest{}
	if err := body(ctx, &request); err != nil {
		return err
	}
	if err := requireReason(request.AuditRequest); err != nil {
		return err
	}
	if err := dependencies.authorize(ctx, request.ActorPlayerID, grouppolicy.ManageAny); err != nil {
		return err
	}
	charge := true
	if request.Charge != nil {
		charge = *request.Charge
	}
	group, replayed, err := dependencies.Identity.CreateAdministrative(auditContext(ctx, request.AuditRequest), identity.AdministrativeCreateParams{
		CreateParams:   identity.CreateParams{OwnerPlayerID: request.OwnerPlayerID, Name: request.Name, Description: request.Description, HomeRoomID: request.HomeRoomID, ColorA: request.ColorA, ColorB: request.ColorB, BadgeParts: parts(request.BadgeParts)},
		IdempotencyKey: ctx.Get("Idempotency-Key"), Charge: charge,
	})
	if err != nil {
		return groupError(err)
	}
	status := fiber.StatusCreated
	if replayed {
		status = fiber.StatusOK
	}
	return ctx.Status(status).JSON(MutationResponse{Group: &group, Created: !replayed})
}

// update applies one optimistic identity or settings patch.
func (dependencies Dependencies) update(ctx *fiber.Ctx) error {
	groupID, request, err := dependencies.updateRequest(ctx, grouppolicy.ManageAny)
	if err != nil {
		return err
	}
	patch := grouprecord.GroupPatch{Name: request.Name, Description: request.Description, State: request.State, CanMembersDecorate: request.CanMembersDecorate, ColorA: request.ColorA, ColorB: request.ColorB}
	group, err := dependencies.Identity.Update(auditContext(ctx, request.AuditRequest), request.ActorPlayerID, groupID, request.Version, patch)
	return groupMutation(ctx, group, err)
}

// deactivate soft-deactivates retained group state.
func (dependencies Dependencies) deactivate(ctx *fiber.Ctx) error {
	groupID, request, err := dependencies.versionRequest(ctx, grouppolicy.DeleteAny)
	if err != nil {
		return err
	}
	group, err := dependencies.Identity.Deactivate(auditContext(ctx, request.AuditRequest), request.ActorPlayerID, groupID, request.Version)
	return groupMutation(ctx, group, err)
}

// restore restores one retained group to an eligible home room.
func (dependencies Dependencies) restore(ctx *fiber.Ctx) error {
	request := struct {
		VersionRequest
		HomeRoomID int64 `json:"homeRoomId"`
	}{}
	groupID, err := positiveParam(ctx, "groupId")
	if err != nil || body(ctx, &request) != nil || requireReason(request.AuditRequest) != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid group restore request")
	}
	if err = dependencies.authorize(ctx, request.ActorPlayerID, grouppolicy.DeleteAny); err != nil {
		return err
	}
	group, err := dependencies.Identity.Restore(auditContext(ctx, request.AuditRequest), request.ActorPlayerID, groupID, request.Version, request.HomeRoomID)
	return groupMutation(ctx, group, err)
}

// badge validates and replaces normalized badge layers.
func (dependencies Dependencies) badge(ctx *fiber.Ctx) error {
	request := struct {
		VersionRequest
		BadgeParts []BadgePartRequest `json:"badgeParts"`
	}{}
	groupID, err := positiveParam(ctx, "groupId")
	if err != nil || body(ctx, &request) != nil || requireReason(request.AuditRequest) != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid group badge request")
	}
	if err = dependencies.authorize(ctx, request.ActorPlayerID, grouppolicy.BadgeManageAny); err != nil {
		return err
	}
	group, err := dependencies.Identity.SaveBadge(auditContext(ctx, request.AuditRequest), request.ActorPlayerID, groupID, request.Version, parts(request.BadgeParts))
	return groupMutation(ctx, group, err)
}

// homeRoom atomically replaces the group headquarters.
func (dependencies Dependencies) homeRoom(ctx *fiber.Ctx) error {
	request := struct {
		VersionRequest
		HomeRoomID int64 `json:"homeRoomId"`
	}{}
	groupID, err := positiveParam(ctx, "groupId")
	if err != nil || body(ctx, &request) != nil || requireReason(request.AuditRequest) != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid group home-room request")
	}
	if err = dependencies.authorize(ctx, request.ActorPlayerID, grouppolicy.HomeRoomRebind); err != nil {
		return err
	}
	group, err := dependencies.Identity.RebindRoom(auditContext(ctx, request.AuditRequest), request.ActorPlayerID, groupID, request.HomeRoomID, request.Version)
	return groupMutation(ctx, group, err)
}

// owner transfers the protected owner membership.
func (dependencies Dependencies) owner(ctx *fiber.Ctx) error {
	request := struct {
		VersionRequest
		OwnerPlayerID int64 `json:"ownerPlayerId"`
	}{}
	groupID, err := positiveParam(ctx, "groupId")
	if err != nil || body(ctx, &request) != nil || requireReason(request.AuditRequest) != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid group owner request")
	}
	if err = dependencies.authorize(ctx, request.ActorPlayerID, grouppolicy.RolesManageAny); err != nil {
		return err
	}
	group, err := dependencies.Identity.TransferOwner(auditContext(ctx, request.AuditRequest), request.ActorPlayerID, groupID, request.OwnerPlayerID, request.Version)
	return groupMutation(ctx, group, err)
}

// updateRequest parses and authorizes one optimistic patch.
func (dependencies Dependencies) updateRequest(ctx *fiber.Ctx, node permission.Node) (int64, UpdateRequest, error) {
	groupID, err := positiveParam(ctx, "groupId")
	request := UpdateRequest{}
	if err != nil || body(ctx, &request) != nil || requireReason(request.AuditRequest) != nil {
		return 0, request, fiber.NewError(fiber.StatusBadRequest, "invalid group update request")
	}
	err = dependencies.authorize(ctx, request.ActorPlayerID, node)
	return groupID, request, err
}

// versionRequest parses and authorizes one versioned mutation.
func (dependencies Dependencies) versionRequest(ctx *fiber.Ctx, node permission.Node) (int64, VersionRequest, error) {
	groupID, err := positiveParam(ctx, "groupId")
	request := VersionRequest{}
	if err != nil || body(ctx, &request) != nil || requireReason(request.AuditRequest) != nil {
		return 0, request, fiber.NewError(fiber.StatusBadRequest, "invalid group version request")
	}
	err = dependencies.authorize(ctx, request.ActorPlayerID, node)
	return groupID, request, err
}

// groupMutation writes one common group response.
func groupMutation(ctx *fiber.Ctx, group grouprecord.Group, err error) error {
	if err != nil {
		return groupError(err)
	}
	return ctx.JSON(MutationResponse{Group: &group})
}
