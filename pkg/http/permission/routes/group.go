package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/pixels/internal/permission"
)

// nodesHandler lists every process-registered permission node.
func nodesHandler() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		registered := permission.RegisteredNodes()
		response := make([]NodeResponse, 0, len(registered))
		for _, node := range registered {
			response = append(response, NodeResponse{Node: node.Node, PerkName: node.PerkName, Package: node.Package})
		}
		return ctx.JSON(response)
	}
}

// groupsHandler lists active permission groups.
func groupsHandler(dependencies Dependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		groups, err := dependencies.Permissions.Groups(ctx.Context())
		if err != nil {
			return err
		}
		response := make([]GroupResponse, 0, len(groups))
		for _, group := range groups {
			response = append(response, groupResponse(group))
		}
		return ctx.JSON(response)
	}
}

// createGroupHandler creates one permission group.
func createGroupHandler(dependencies Dependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var request GroupRequest
		if err := ctx.BodyParser(&request); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid permission group request body")
		}
		group, err := dependencies.Permissions.CreateGroup(ctx.Context(), groupParams(request))
		if err != nil {
			return permissionError(err)
		}
		return ctx.Status(fiber.StatusCreated).JSON(groupResponse(group))
	}
}

// updateGroupHandler updates one permission group.
func updateGroupHandler(dependencies Dependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		groupID, err := routeID(ctx, "id")
		if err != nil {
			return err
		}
		var request GroupPatchRequest
		if err := ctx.BodyParser(&request); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid permission group patch body")
		}
		group, err := dependencies.Permissions.UpdateGroup(ctx.Context(), groupID, groupPatch(request))
		if err != nil {
			return permissionError(err)
		}
		return ctx.JSON(groupResponse(group))
	}
}

// grantGroupNodeHandler creates or replaces one group grant.
func grantGroupNodeHandler(dependencies Dependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		groupID, err := routeID(ctx, "id")
		if err != nil {
			return err
		}
		var request NodeRequest
		if err := ctx.BodyParser(&request); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid permission group node request body")
		}
		if err := dependencies.Permissions.GrantGroupNode(ctx.Context(), groupID, request.Node, request.Allowed); err != nil {
			return permissionError(err)
		}
		return ctx.JSON(MutationResponse{Updated: true})
	}
}

// revokeGroupNodeHandler removes one group grant.
func revokeGroupNodeHandler(dependencies Dependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		groupID, err := routeID(ctx, "id")
		if err != nil {
			return err
		}
		node, err := routeNode(ctx)
		if err != nil {
			return err
		}
		if err := dependencies.Permissions.RevokeGroupNode(ctx.Context(), groupID, node); err != nil {
			return permissionError(err)
		}
		return ctx.SendStatus(fiber.StatusNoContent)
	}
}
