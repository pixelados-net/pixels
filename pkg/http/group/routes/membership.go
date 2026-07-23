package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/pixels/internal/permission"
	grouppolicy "github.com/niflaot/pixels/internal/realm/group/policy"
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
)

// members returns one filtered member or pending page.
func (dependencies Dependencies) members(ctx *fiber.Ctx) error {
	actorID, err := dependencies.readAuthorized(ctx, grouppolicy.MembersManageAny)
	if err != nil {
		return err
	}
	groupID, err := positiveParam(ctx, "groupId")
	if err != nil {
		return err
	}
	page, err := boundedInt(ctx.Query("page"), 0, 1000000)
	if err != nil {
		return err
	}
	level, err := boundedInt(ctx.Query("level"), 0, 2)
	if err != nil {
		return err
	}
	result, canManage, err := dependencies.Membership.MemberPage(ctx.Context(), actorID, groupID, int32(page), ctx.Query("query"), int32(level))
	if err != nil {
		return groupError(err)
	}
	return ctx.JSON(MemberPageResponse{Page: result, CanManage: canManage})
}

// addMember inserts or updates one non-owner membership idempotently.
func (dependencies Dependencies) addMember(ctx *fiber.Ctx) error {
	request := struct {
		AuditRequest
		PlayerID int64            `json:"playerId"`
		Role     grouprecord.Role `json:"role"`
	}{}
	groupID, err := dependencies.mutation(ctx, "groupId", &request.AuditRequest, grouppolicy.MembersManageAny, &request)
	if err != nil {
		return err
	}
	member, created, err := dependencies.Membership.Add(auditContext(ctx, request.AuditRequest), request.ActorPlayerID, groupID, request.PlayerID, request.Role)
	if err != nil {
		return groupError(err)
	}
	return ctx.JSON(MutationResponse{Membership: &member, Created: created})
}

// role changes one existing non-owner role.
func (dependencies Dependencies) role(ctx *fiber.Ctx) error {
	request := struct {
		AuditRequest
		Role grouprecord.Role `json:"role"`
	}{}
	groupID, err := dependencies.mutation(ctx, "groupId", &request.AuditRequest, grouppolicy.RolesManageAny, &request)
	if err != nil {
		return err
	}
	playerID, err := positiveParam(ctx, "playerId")
	if err != nil {
		return err
	}
	member, err := dependencies.Membership.ChangeRole(auditContext(ctx, request.AuditRequest), request.ActorPlayerID, groupID, playerID, request.Role)
	if err != nil {
		return groupError(err)
	}
	return ctx.JSON(MutationResponse{Membership: &member})
}

// removeMember removes one membership and atomically returns HQ furniture.
func (dependencies Dependencies) removeMember(ctx *fiber.Ctx) error {
	request := AuditRequest{}
	groupID, err := dependencies.mutation(ctx, "groupId", &request, grouppolicy.MembersManageAny, &request)
	if err != nil {
		return err
	}
	playerID, err := positiveParam(ctx, "playerId")
	if err != nil {
		return err
	}
	count, err := dependencies.Membership.Remove(auditContext(ctx, request), request.ActorPlayerID, groupID, playerID)
	if err != nil {
		return groupError(err)
	}
	return ctx.JSON(MutationResponse{Count: count})
}

// requests returns one bounded pending request page.
func (dependencies Dependencies) requests(ctx *fiber.Ctx) error {
	actorID, err := dependencies.readAuthorized(ctx, grouppolicy.MembersManageAny)
	if err != nil {
		return err
	}
	groupID, err := positiveParam(ctx, "groupId")
	if err != nil {
		return err
	}
	offset, err := boundedInt(ctx.Query("offset"), 0, 1000000)
	if err != nil {
		return err
	}
	limit, err := boundedInt(ctx.Query("limit"), 50, 100)
	if err != nil || limit < 1 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request list limit")
	}
	items, err := dependencies.Membership.Requests(ctx.Context(), actorID, groupID, offset, limit)
	if err != nil {
		return groupError(err)
	}
	return ctx.JSON(items)
}

// accept accepts one pending membership request.
func (dependencies Dependencies) accept(ctx *fiber.Ctx) error {
	request := AuditRequest{}
	groupID, err := dependencies.mutation(ctx, "groupId", &request, grouppolicy.MembersManageAny, &request)
	if err != nil {
		return err
	}
	playerID, err := positiveParam(ctx, "playerId")
	if err != nil {
		return err
	}
	member, err := dependencies.Membership.Accept(auditContext(ctx, request), request.ActorPlayerID, groupID, playerID)
	if err != nil {
		return groupError(err)
	}
	return ctx.JSON(MutationResponse{Membership: &member})
}

// decline removes one pending request.
func (dependencies Dependencies) decline(ctx *fiber.Ctx) error {
	request := AuditRequest{}
	groupID, err := dependencies.mutation(ctx, "groupId", &request, grouppolicy.MembersManageAny, &request)
	if err != nil {
		return err
	}
	playerID, err := positiveParam(ctx, "playerId")
	if err != nil {
		return err
	}
	removed, err := dependencies.Membership.Decline(auditContext(ctx, request), request.ActorPlayerID, groupID, playerID)
	if err != nil {
		return groupError(err)
	}
	count := 0
	if removed {
		count = 1
	}
	return ctx.JSON(MutationResponse{Count: count})
}

// approveAll accepts one configured ordered pending batch.
func (dependencies Dependencies) approveAll(ctx *fiber.Ctx) error {
	request := AuditRequest{}
	groupID, err := dependencies.mutation(ctx, "groupId", &request, grouppolicy.MembersManageAny, &request)
	if err != nil {
		return err
	}
	members, err := dependencies.Membership.ApproveAll(auditContext(ctx, request), request.ActorPlayerID, groupID)
	if err != nil {
		return groupError(err)
	}
	return ctx.JSON(MutationResponse{Count: len(members)})
}

// playerGroups returns memberships and favorite for one player.
func (dependencies Dependencies) playerGroups(ctx *fiber.Ctx) error {
	if _, err := dependencies.readAuthorized(ctx, grouppolicy.ManageAny); err != nil {
		return err
	}
	playerID, err := positiveParam(ctx, "playerId")
	if err != nil {
		return err
	}
	groups, err := dependencies.Membership.PlayerGroups(ctx.Context(), playerID)
	if err != nil {
		return groupError(err)
	}
	return ctx.JSON(groups)
}

// favorite sets or clears exactly one player's favorite membership.
func (dependencies Dependencies) favorite(ctx *fiber.Ctx) error {
	request := struct {
		AuditRequest
		GroupID *int64 `json:"groupId"`
	}{}
	if err := body(ctx, &request); err != nil || requireReason(request.AuditRequest) != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid favorite request")
	}
	if err := dependencies.authorize(ctx, request.ActorPlayerID, grouppolicy.MembersManageAny); err != nil {
		return err
	}
	playerID, err := positiveParam(ctx, "playerId")
	if err != nil {
		return err
	}
	if err = dependencies.Membership.SetFavorite(auditContext(ctx, request.AuditRequest), playerID, request.GroupID); err != nil {
		return groupError(err)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

// readAuthorized resolves and authorizes one read actor.
func (dependencies Dependencies) readAuthorized(ctx *fiber.Ctx, node permission.Node) (int64, error) {
	actorID, err := readActor(ctx)
	if err != nil {
		return 0, err
	}
	return actorID, dependencies.authorize(ctx, actorID, node)
}

// mutation parses, validates, and authorizes one attributed body.
func (dependencies Dependencies) mutation(ctx *fiber.Ctx, path string, audit *AuditRequest, node permission.Node, target any) (int64, error) {
	groupID, err := positiveParam(ctx, path)
	if err != nil {
		return 0, err
	}
	if err = body(ctx, target); err != nil {
		return 0, err
	}
	if err = requireReason(*audit); err != nil {
		return 0, err
	}
	return groupID, dependencies.authorize(ctx, audit.ActorPlayerID, node)
}
