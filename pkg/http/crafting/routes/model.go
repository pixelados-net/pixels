package routes

import (
	craftingconfig "github.com/niflaot/pixels/internal/realm/crafting/config"
	craftingrecord "github.com/niflaot/pixels/internal/realm/crafting/record"
)

// AuditRequest stores required administrative attribution.
type AuditRequest struct {
	ActorPlayerID int64  `json:"actorPlayerId"`
	Reason        string `json:"reason"`
}

// AltarRequest registers one furniture definition as an altar.
type AltarRequest struct {
	AuditRequest
	DefinitionID int64 `json:"definitionId"`
}

// RecipeCreateRequest creates one complete recipe.
type RecipeCreateRequest struct {
	AuditRequest
	Name               string                      `json:"name"`
	RewardDefinitionID int64                       `json:"rewardDefinitionId"`
	Secret             bool                        `json:"secret"`
	Limited            bool                        `json:"limited"`
	Remaining          *int32                      `json:"remaining"`
	AchievementCode    string                      `json:"achievementCode"`
	Ingredients        []craftingrecord.Ingredient `json:"ingredients"`
}

// RecipeUpdateRequest applies an optimistic recipe patch.
type RecipeUpdateRequest struct {
	AuditRequest
	craftingrecord.RecipePatch
}

// RecyclerConfigRequest replaces mutable recycler policy.
type RecyclerConfigRequest struct {
	AuditRequest
	Enabled      bool          `json:"enabled"`
	BatchSize    int           `json:"batchSize"`
	RarityChance map[int32]int `json:"rarityChance"`
}

// PrizeRequest adds one recycler prize.
type PrizeRequest struct {
	AuditRequest
	Tier               int32 `json:"tier"`
	RewardDefinitionID int64 `json:"rewardDefinitionId"`
}

// AltarResponse contains one altar and its grouped recipes.
type AltarResponse struct {
	Altar   craftingrecord.Altar    `json:"altar"`
	Recipes []craftingrecord.Recipe `json:"recipes"`
}

// RecyclerConfigResponse exposes the active mutable recycler policy.
type RecyclerConfigResponse struct {
	Enabled      bool          `json:"enabled"`
	BatchSize    int           `json:"batchSize"`
	RarityChance map[int32]int `json:"rarityChance"`
}

func configResponse(config craftingconfig.Config) RecyclerConfigResponse {
	return RecyclerConfigResponse{Enabled: config.RecyclerEnabled, BatchSize: config.RecyclerBatchSize, RarityChance: config.RecyclerRarityChance}
}
