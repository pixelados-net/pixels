package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/niflaot/pixels/internal/permission"
)

// addPlayerGroupHandler adds one player to a permission group.
func addPlayerGroupHandler(dependencies Dependencies) fiber.Handler {
	return membershipHandler(dependencies, true)
}

// removePlayerGroupHandler removes one player from a permission group.
func removePlayerGroupHandler(dependencies Dependencies) fiber.Handler {
	return membershipHandler(dependencies, false)
}

// membershipHandler mutates one player group membership.
func membershipHandler(dependencies Dependencies, add bool) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		playerID, err := routeID(ctx, "playerId")
		if err != nil {
			return err
		}
		groupID, err := routeID(ctx, "groupId")
		if err != nil {
			return err
		}
		if add {
			err = dependencies.Permissions.AddPlayerToGroup(ctx.Context(), playerID, groupID)
		} else {
			err = dependencies.Permissions.RemovePlayerFromGroup(ctx.Context(), playerID, groupID)
		}
		if err != nil {
			return permissionError(err)
		}
		if !add {
			return ctx.SendStatus(fiber.StatusNoContent)
		}
		return ctx.JSON(MutationResponse{Updated: true})
	}
}

// grantPlayerNodeHandler creates or replaces one direct player grant.
func grantPlayerNodeHandler(dependencies Dependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		playerID, err := routeID(ctx, "playerId")
		if err != nil {
			return err
		}
		var request NodeRequest
		if err := ctx.BodyParser(&request); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid permission player node request body")
		}
		if err := dependencies.Permissions.GrantPlayerNode(ctx.Context(), playerID, request.Node, request.Allowed); err != nil {
			return permissionError(err)
		}
		return ctx.JSON(MutationResponse{Updated: true})
	}
}

// revokePlayerNodeHandler removes one direct player grant.
func revokePlayerNodeHandler(dependencies Dependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		playerID, err := routeID(ctx, "playerId")
		if err != nil {
			return err
		}
		node, err := routeNode(ctx)
		if err != nil {
			return err
		}
		if err := dependencies.Permissions.RevokePlayerNode(ctx.Context(), playerID, node); err != nil {
			return permissionError(err)
		}
		return ctx.SendStatus(fiber.StatusNoContent)
	}
}

// effectiveNodesHandler lists one player's resolved registered nodes.
func effectiveNodesHandler(dependencies Dependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		playerID, err := routeID(ctx, "playerId")
		if err != nil {
			return err
		}
		nodes, err := dependencies.Permissions.EffectiveNodes(ctx.Context(), playerID)
		if err != nil {
			return permissionError(err)
		}
		response := make([]EffectiveNodeResponse, 0, len(nodes))
		for _, node := range nodes {
			response = append(response, effectiveResponse(node))
		}
		return ctx.JSON(response)
	}
}

// checkHandler resolves one permission node for a player.
func checkHandler(dependencies Dependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		playerID, err := routeID(ctx, "playerId")
		if err != nil {
			return err
		}
		node := permission.Node(ctx.Query("node"))
		allowed, err := dependencies.Permissions.HasPermission(ctx.Context(), playerID, node)
		if err != nil {
			return permissionError(err)
		}
		return ctx.JSON(CheckResponse{PlayerID: playerID, Node: node, Allowed: allowed})
	}
}
