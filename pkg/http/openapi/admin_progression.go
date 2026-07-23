package openapi

import (
	"net/http"

	progressionobservability "github.com/niflaot/pixels/internal/realm/progression/observability"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
	progressionroutes "github.com/niflaot/pixels/pkg/http/progression/routes"
	progressionrequest "github.com/niflaot/pixels/pkg/http/progression/routes/request"
)

// adminProgressionOperations returns protected progression administration operations.
func adminProgressionOperations() []operation {
	return []operation{
		adminProgression(http.MethodGet, "/api/admin/progression/achievements", "List achievement definitions", &ProgressionActorRequest{}, &[]progressionrecord.AchievementDefinition{}, http.StatusOK),
		adminProgression(http.MethodPost, "/api/admin/progression/achievements", "Create achievement definition", &progressionrequest.AchievementCreate{}, &progressionrecord.AchievementDefinition{}, http.StatusCreated),
		adminProgression(http.MethodPatch, "/api/admin/progression/achievements/{id}", "Update achievement definition", &ProgressionAchievementUpdateRequest{}, &progressionrecord.AchievementDefinition{}, http.StatusOK),
		adminProgression(http.MethodDelete, "/api/admin/progression/achievements/{id}", "Disable achievement definition", &ProgressionIDMutationRequest{}, nil, http.StatusNoContent),
		adminProgression(http.MethodPost, "/api/admin/progression/achievements/{id}/levels", "Create achievement level", &ProgressionAchievementLevelCreateRequest{}, &progressionrecord.AchievementLevel{}, http.StatusOK),
		adminProgression(http.MethodPatch, "/api/admin/progression/achievements/{id}/levels/{level}", "Update achievement level", &ProgressionAchievementLevelUpdateRequest{}, &progressionrecord.AchievementLevel{}, http.StatusOK),
		adminProgression(http.MethodDelete, "/api/admin/progression/achievements/{id}/levels/{level}", "Delete highest achievement level", &ProgressionLevelMutationRequest{}, nil, http.StatusNoContent),
		adminProgression(http.MethodGet, "/api/admin/progression/players/{playerId}/achievements", "List player achievements", &ProgressionPlayerReadRequest{}, &[]progressionroutes.PlayerAchievementResponse{}, http.StatusOK),
		adminProgression(http.MethodPost, "/api/admin/progression/players/{playerId}/achievements/{definitionId}/progress", "Add player achievement progress", &ProgressionPlayerProgressRequest{}, &ProgressionObjectResponse{}, http.StatusOK),
		adminProgression(http.MethodPut, "/api/admin/progression/players/{playerId}/achievements/{definitionId}/level", "Force player achievement level", &ProgressionPlayerLevelRequest{}, &ProgressionObjectResponse{}, http.StatusOK),
		adminProgression(http.MethodDelete, "/api/admin/progression/players/{playerId}/achievements/{definitionId}", "Reset player achievement", &ProgressionPlayerDefinitionMutationRequest{}, nil, http.StatusNoContent),
		adminProgression(http.MethodPost, "/api/admin/progression/players/{playerId}/badges", "Grant player badge", &ProgressionPlayerBadgeRequest{}, &ProgressionObjectResponse{}, http.StatusOK),
		adminProgression(http.MethodDelete, "/api/admin/progression/players/{playerId}/badges/{code}", "Remove player badge", &ProgressionPlayerCodeMutationRequest{}, nil, http.StatusNoContent),
		adminProgression(http.MethodPut, "/api/admin/progression/talents/{track}/levels/{level}", "Upsert talent track level", &ProgressionTalentLevelRequest{}, &progressionrecord.TalentLevel{}, http.StatusOK),
		adminProgression(http.MethodGet, "/api/admin/progression/players/{playerId}/talents", "List player talent levels", &ProgressionPlayerReadRequest{}, &[]progressionrecord.PlayerTalent{}, http.StatusOK),
		adminProgression(http.MethodPut, "/api/admin/progression/players/{playerId}/talents/{track}", "Force player talent level", &ProgressionPlayerTalentRequest{}, &ProgressionObjectResponse{}, http.StatusOK),
		adminProgression(http.MethodGet, "/api/admin/progression/campaigns", "List quest campaigns", &ProgressionActorRequest{}, &[]progressionrecord.QuestCampaign{}, http.StatusOK),
		adminProgression(http.MethodPost, "/api/admin/progression/campaigns", "Create quest campaign", &progressionrequest.Campaign{}, &progressionrecord.QuestCampaign{}, http.StatusCreated),
		adminProgression(http.MethodPatch, "/api/admin/progression/campaigns/{code}", "Update quest campaign", &ProgressionCampaignRequest{}, &progressionrecord.QuestCampaign{}, http.StatusOK),
		adminProgression(http.MethodDelete, "/api/admin/progression/campaigns/{code}", "Disable quest campaign", &ProgressionCodeMutationRequest{}, nil, http.StatusNoContent),
		adminProgression(http.MethodGet, "/api/admin/progression/quests", "List quest definitions", &ProgressionActorRequest{}, &[]progressionrecord.QuestDefinition{}, http.StatusOK),
		adminProgression(http.MethodPost, "/api/admin/progression/quests", "Create quest definition", &progressionrequest.Quest{}, &progressionrecord.QuestDefinition{}, http.StatusCreated),
		adminProgression(http.MethodPatch, "/api/admin/progression/quests/{id}", "Update quest definition", &ProgressionQuestRequest{}, &progressionrecord.QuestDefinition{}, http.StatusOK),
		adminProgression(http.MethodDelete, "/api/admin/progression/quests/{id}", "Disable quest definition", &ProgressionIDMutationRequest{}, nil, http.StatusNoContent),
		adminProgression(http.MethodGet, "/api/admin/progression/players/{playerId}/quests", "Read player quest state", &ProgressionPlayerReadRequest{}, &progressionroutes.PlayerQuestResponse{}, http.StatusOK),
		adminProgression(http.MethodPost, "/api/admin/progression/players/{playerId}/quests/{questId}/complete", "Force-complete player quest", &ProgressionPlayerQuestMutationRequest{}, &ProgressionObjectResponse{}, http.StatusOK),
		adminProgression(http.MethodDelete, "/api/admin/progression/players/{playerId}/quests/active", "Cancel active player quest", &ProgressionPlayerMutationRequest{}, nil, http.StatusNoContent),
		adminProgression(http.MethodGet, "/api/admin/progression/quizzes", "List quizzes", &ProgressionActorRequest{}, &[]progressionrecord.Quiz{}, http.StatusOK),
		adminProgression(http.MethodPost, "/api/admin/progression/quizzes", "Create quiz", &progressionrequest.Quiz{}, &progressionrecord.Quiz{}, http.StatusCreated),
		adminProgression(http.MethodDelete, "/api/admin/progression/quizzes/{code}", "Disable quiz", &ProgressionCodeMutationRequest{}, nil, http.StatusNoContent),
		adminProgression(http.MethodPost, "/api/admin/progression/quizzes/{code}/questions", "Create quiz question", &ProgressionQuizQuestionRequest{}, &progressionrecord.QuizQuestion{}, http.StatusCreated),
		adminProgression(http.MethodPatch, "/api/admin/progression/quizzes/{code}/questions/{id}", "Update quiz question", &ProgressionQuizQuestionIDRequest{}, &progressionrecord.QuizQuestion{}, http.StatusOK),
		adminProgression(http.MethodDelete, "/api/admin/progression/quizzes/{code}/questions/{id}", "Delete quiz question", &ProgressionCodeIDMutationRequest{}, nil, http.StatusNoContent),
		adminProgression(http.MethodGet, "/api/admin/progression/players/{playerId}/quizzes/{code}", "Read player quiz result", &ProgressionPlayerCodeReadRequest{}, &progressionrecord.QuizResult{}, http.StatusOK),
		adminProgression(http.MethodPost, "/api/admin/progression/polls", "Start room word quiz", &ProgressionPollRequest{}, &ProgressionObjectResponse{}, http.StatusCreated),
		adminProgression(http.MethodGet, "/api/admin/progression/promos", "List badge promotions", &ProgressionActorRequest{}, &[]progressionrecord.PromoBadge{}, http.StatusOK),
		adminProgression(http.MethodPost, "/api/admin/progression/promos", "Create badge promotion", &progressionrequest.Promo{}, &progressionrecord.PromoBadge{}, http.StatusCreated),
		adminProgression(http.MethodPatch, "/api/admin/progression/promos/{code}", "Update badge promotion", &ProgressionPromoRequest{}, &progressionrecord.PromoBadge{}, http.StatusOK),
		adminProgression(http.MethodDelete, "/api/admin/progression/promos/{code}", "Disable badge promotion", &ProgressionCodeMutationRequest{}, nil, http.StatusNoContent),
		adminProgression(http.MethodGet, "/api/admin/progression/promos/{code}/claims", "List promotion claims", &ProgressionCodeReadRequest{}, &[]progressionrecord.PromoClaim{}, http.StatusOK),
		adminProgression(http.MethodPost, "/api/admin/progression/players/{playerId}/promos/{code}/claim", "Claim promotion for player", &ProgressionPlayerPromoRequest{}, &ProgressionObjectResponse{}, http.StatusOK),
		adminProgression(http.MethodPost, "/api/admin/progression/reload", "Reload progression catalog", &progressionroutes.AuditRequest{}, &ProgressionObjectResponse{}, http.StatusOK),
		adminProgression(http.MethodGet, "/api/admin/progression/metrics", "Read progression metrics", &ProgressionActorRequest{}, &progressionobservability.Snapshot{}, http.StatusOK),
	}
}

// adminProgression creates one protected progression administration operation.
func adminProgression(method string, path string, summary string, request any, body any, status int) operation {
	responses := errorResponses(http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusConflict, http.StatusUnprocessableEntity, http.StatusInternalServerError)
	if body == nil {
		responses = append([]response{emptyResponse(status, summary+".")}, responses...)
	} else {
		responses = append([]response{jsonResponse(status, body, summary+".")}, responses...)
	}
	return operation{method: method, path: path, tag: "Progression", summary: summary, description: summary + ". Mutations require actorPlayerId and reason and are audited atomically.", request: request, responses: responses, secured: true}
}

// ProgressionActorRequest documents administrative read attribution.
type ProgressionActorRequest struct {
	ActorPlayerID int64 `header:"X-Actor-Player-ID" required:"true"`
}

// ProgressionIDMutationRequest documents one numeric path mutation.
type ProgressionIDMutationRequest struct {
	ID int64 `path:"id"`
	progressionroutes.AuditRequest
}

// ProgressionLevelMutationRequest documents one achievement level deletion.
type ProgressionLevelMutationRequest struct {
	ID    int64 `path:"id"`
	Level int32 `path:"level"`
	progressionroutes.AuditRequest
}

// ProgressionAchievementUpdateRequest documents one achievement update.
type ProgressionAchievementUpdateRequest struct {
	ID int64 `path:"id"`
	progressionrequest.AchievementUpdate
}

// ProgressionAchievementLevelCreateRequest documents one achievement level creation.
type ProgressionAchievementLevelCreateRequest struct {
	ID int64 `path:"id"`
	progressionrequest.AchievementLevel
}

// ProgressionAchievementLevelUpdateRequest documents one achievement level update.
type ProgressionAchievementLevelUpdateRequest struct {
	ID    int64 `path:"id"`
	Level int32 `path:"level"`
	progressionrequest.AchievementLevel
}

// ProgressionPlayerReadRequest documents one player read.
type ProgressionPlayerReadRequest struct {
	PlayerID int64 `path:"playerId"`
	ProgressionActorRequest
}

// ProgressionPlayerDefinitionMutationRequest documents one player achievement mutation.
type ProgressionPlayerDefinitionMutationRequest struct {
	PlayerID     int64 `path:"playerId"`
	DefinitionID int64 `path:"definitionId"`
	progressionroutes.AuditRequest
}

// ProgressionPlayerProgressRequest documents one player progress mutation.
type ProgressionPlayerProgressRequest struct {
	PlayerID     int64 `path:"playerId"`
	DefinitionID int64 `path:"definitionId"`
	progressionrequest.Progress
}

// ProgressionPlayerLevelRequest documents one exact achievement level mutation.
type ProgressionPlayerLevelRequest struct {
	PlayerID     int64 `path:"playerId"`
	DefinitionID int64 `path:"definitionId"`
	progressionrequest.ForceLevel
}

// ProgressionPlayerBadgeRequest documents one direct badge grant.
type ProgressionPlayerBadgeRequest struct {
	PlayerID int64 `path:"playerId"`
	progressionrequest.Badge
}

// ProgressionPlayerCodeMutationRequest documents one player code deletion.
type ProgressionPlayerCodeMutationRequest struct {
	PlayerID int64  `path:"playerId"`
	Code     string `path:"code"`
	progressionroutes.AuditRequest
}

// ProgressionTalentLevelRequest documents one talent level definition.
type ProgressionTalentLevelRequest struct {
	Track string `path:"track"`
	Level int32  `path:"level"`
	progressionrequest.TalentLevel
}

// ProgressionPlayerTalentRequest documents one exact player talent mutation.
type ProgressionPlayerTalentRequest struct {
	PlayerID int64  `path:"playerId"`
	Track    string `path:"track"`
	progressionrequest.TalentForce
}

// ProgressionCampaignRequest documents one campaign update.
type ProgressionCampaignRequest struct {
	Code string `path:"code"`
	progressionrequest.Campaign
}

// ProgressionCodeMutationRequest documents one string path mutation.
type ProgressionCodeMutationRequest struct {
	Code string `path:"code"`
	progressionroutes.AuditRequest
}

// ProgressionQuestRequest documents one quest update.
type ProgressionQuestRequest struct {
	ID int64 `path:"id"`
	progressionrequest.Quest
}

// ProgressionPlayerQuestMutationRequest documents one forced quest completion.
type ProgressionPlayerQuestMutationRequest struct {
	PlayerID int64 `path:"playerId"`
	QuestID  int64 `path:"questId"`
	progressionroutes.AuditRequest
}

// ProgressionPlayerMutationRequest documents one player mutation.
type ProgressionPlayerMutationRequest struct {
	PlayerID int64 `path:"playerId"`
	progressionroutes.AuditRequest
}

// ProgressionQuizQuestionRequest documents one quiz question creation.
type ProgressionQuizQuestionRequest struct {
	Code string `path:"code"`
	progressionrequest.QuizQuestion
}

// ProgressionQuizQuestionIDRequest documents one quiz question update.
type ProgressionQuizQuestionIDRequest struct {
	Code string `path:"code"`
	ID   int64  `path:"id"`
	progressionrequest.QuizQuestion
}

// ProgressionCodeIDMutationRequest documents one string and numeric path mutation.
type ProgressionCodeIDMutationRequest struct {
	Code string `path:"code"`
	ID   int64  `path:"id"`
	progressionroutes.AuditRequest
}

// ProgressionPlayerCodeReadRequest documents one player code read.
type ProgressionPlayerCodeReadRequest struct {
	PlayerID int64  `path:"playerId"`
	Code     string `path:"code"`
	ProgressionActorRequest
}

// ProgressionPromoRequest documents one promotion update.
type ProgressionPromoRequest struct {
	Code string `path:"code"`
	progressionrequest.Promo
}

// ProgressionCodeReadRequest documents one code read.
type ProgressionCodeReadRequest struct {
	Code string `path:"code"`
	ProgressionActorRequest
}

// ProgressionPlayerPromoRequest documents one player promotion claim.
type ProgressionPlayerPromoRequest struct {
	PlayerID int64  `path:"playerId"`
	Code     string `path:"code"`
	progressionrequest.PromoClaim
}
