package openapi

import (
	"net/http"

	craftingroutes "github.com/niflaot/pixels/pkg/http/crafting/routes"
)

// adminCraftingOperations returns protected crafting administration operations.
func adminCraftingOperations() []operation {
	return []operation{
		adminCrafting(http.MethodGet, "/api/admin/crafting/altars", "List crafting altars", &CraftingActorRequest{}, &[]craftingroutes.AltarResponse{}, http.StatusOK),
		adminCrafting(http.MethodPost, "/api/admin/crafting/altars", "Register crafting altar", &craftingroutes.AltarRequest{}, &CraftingObjectResponse{}, http.StatusCreated),
		adminCrafting(http.MethodDelete, "/api/admin/crafting/altars/{definitionId}", "Disable crafting altar", &CraftingDeleteRequest{}, nil, http.StatusNoContent),
		adminCrafting(http.MethodPost, "/api/admin/crafting/altars/{definitionId}/recipes", "Create crafting recipe", &CraftingRecipeCreateRequest{}, &CraftingObjectResponse{}, http.StatusCreated),
		adminCrafting(http.MethodPatch, "/api/admin/crafting/recipes/{recipeId}", "Update crafting recipe", &CraftingRecipeUpdateRequest{}, &CraftingObjectResponse{}, http.StatusOK),
		adminCrafting(http.MethodDelete, "/api/admin/crafting/recipes/{recipeId}", "Disable crafting recipe", &CraftingRecipeDeleteRequest{}, nil, http.StatusNoContent),
		adminCrafting(http.MethodGet, "/api/admin/crafting/players/{playerId}/recipes", "List known secret recipes", &CraftingPlayerRequest{}, &CraftingObjectResponse{}, http.StatusOK),
		adminCrafting(http.MethodPost, "/api/admin/crafting/players/{playerId}/recipes/{recipeId}", "Grant recipe knowledge", &CraftingKnowledgeRequest{}, &CraftingObjectResponse{}, http.StatusOK),
		adminCrafting(http.MethodDelete, "/api/admin/crafting/players/{playerId}/recipes/{recipeId}", "Revoke recipe knowledge", &CraftingKnowledgeRequest{}, &CraftingObjectResponse{}, http.StatusOK),
		adminCrafting(http.MethodGet, "/api/admin/crafting/recycler/config", "Read recycler config", &CraftingActorRequest{}, &craftingroutes.RecyclerConfigResponse{}, http.StatusOK),
		adminCrafting(http.MethodPut, "/api/admin/crafting/recycler/config", "Update recycler config", &craftingroutes.RecyclerConfigRequest{}, &craftingroutes.RecyclerConfigResponse{}, http.StatusOK),
		adminCrafting(http.MethodGet, "/api/admin/crafting/recycler/prizes", "List recycler prizes", &CraftingActorRequest{}, &CraftingObjectResponse{}, http.StatusOK),
		adminCrafting(http.MethodPost, "/api/admin/crafting/recycler/prizes", "Add recycler prize", &craftingroutes.PrizeRequest{}, &CraftingObjectResponse{}, http.StatusCreated),
		adminCrafting(http.MethodDelete, "/api/admin/crafting/recycler/prizes/{tier}/{definitionId}", "Delete recycler prize", &CraftingPrizeDeleteRequest{}, nil, http.StatusNoContent),
	}
}

func adminCrafting(method string, path string, summary string, request any, body any, status int) operation {
	responses := errorResponses(http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusConflict, http.StatusUnprocessableEntity, http.StatusInternalServerError)
	if body == nil {
		responses = append([]response{emptyResponse(status, summary+".")}, responses...)
	} else {
		responses = append([]response{jsonResponse(status, body, summary+".")}, responses...)
	}
	return operation{method: method, path: path, tag: "Crafting", summary: summary, description: summary + ". Mutations require actorPlayerId and reason for durable audit.", request: request, responses: responses, secured: true}
}

// CraftingActorRequest documents read attribution.
type CraftingActorRequest struct {
	ActorPlayerID int64 `header:"X-Actor-Player-ID" required:"true"`
}

// CraftingDeleteRequest documents one path mutation.
type CraftingDeleteRequest struct {
	DefinitionID int64 `path:"definitionId"`
	craftingroutes.AuditRequest
}

// CraftingRecipeCreateRequest documents one altar recipe creation.
type CraftingRecipeCreateRequest struct {
	DefinitionID int64 `path:"definitionId"`
	craftingroutes.RecipeCreateRequest
}

// CraftingRecipeUpdateRequest documents one optimistic recipe update.
type CraftingRecipeUpdateRequest struct {
	RecipeID int64 `path:"recipeId"`
	craftingroutes.RecipeUpdateRequest
}

// CraftingRecipeDeleteRequest documents one recipe disable mutation.
type CraftingRecipeDeleteRequest struct {
	RecipeID int64 `path:"recipeId"`
	craftingroutes.AuditRequest
}

// CraftingPlayerRequest documents one player read.
type CraftingPlayerRequest struct {
	PlayerID int64 `path:"playerId"`
	CraftingActorRequest
}

// CraftingKnowledgeRequest documents player-recipe paths and audit body.
type CraftingKnowledgeRequest struct {
	PlayerID int64 `path:"playerId"`
	RecipeID int64 `path:"recipeId"`
	craftingroutes.AuditRequest
}

// CraftingPrizeDeleteRequest documents recycler prize deletion.
type CraftingPrizeDeleteRequest struct {
	Tier         int32 `path:"tier"`
	DefinitionID int64 `path:"definitionId"`
	craftingroutes.AuditRequest
}

// CraftingObjectResponse documents heterogeneous crafting records.
type CraftingObjectResponse map[string]any
