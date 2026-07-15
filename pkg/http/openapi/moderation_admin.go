package openapi

import "net/http"

// moderationAdminOperations returns protected global moderation operations.
func moderationAdminOperations() []operation {
	return []operation{
		adminModeration(http.MethodGet, "/api/admin/players/{playerId}/punishments", "List player punishments", &ModerationPlayerRequest{}, &ModerationPunishmentListResponse{}, http.StatusOK),
		adminModeration(http.MethodPost, "/api/admin/players/{playerId}/punishments", "Apply player punishment", &ModerationPunishmentApplyRequest{}, &ModerationPunishmentResponse{}, http.StatusCreated),
		adminModeration(http.MethodDelete, "/api/admin/punishments/{id}", "Revoke punishment", &ModerationPunishmentRevokeRequest{}, &ModerationPunishmentResponse{}, http.StatusOK),
		adminModeration(http.MethodGet, "/api/admin/moderation/issues", "List moderation issues", &ModerationIssuesRequest{}, &ModerationIssueListResponse{}, http.StatusOK),
		adminModeration(http.MethodGet, "/api/admin/moderation/cfh-topics", "List call-for-help topics", &APIKeyRequest{}, &ModerationTopicListResponse{}, http.StatusOK),
		adminModeration(http.MethodPost, "/api/admin/moderation/cfh-topics", "Create call-for-help topic", &ModerationTopicRequest{}, &ModerationTopicResponse{}, http.StatusCreated),
		adminModeration(http.MethodPatch, "/api/admin/moderation/cfh-topics/{id}", "Update call-for-help topic", &ModerationTopicPatchRequest{}, &ModerationTopicResponse{}, http.StatusOK),
		adminModeration(http.MethodGet, "/api/admin/moderation/presets", "List moderator presets", &APIKeyRequest{}, &ModerationPresetListResponse{}, http.StatusOK),
		adminModeration(http.MethodPost, "/api/admin/moderation/presets", "Create moderator preset", &ModerationPresetRequest{}, &ModerationPresetResponse{}, http.StatusCreated),
		adminModeration(http.MethodPatch, "/api/admin/moderation/presets/{id}", "Update moderator preset", &ModerationPresetPatchRequest{}, &ModerationPresetResponse{}, http.StatusOK),
		adminModeration(http.MethodGet, "/api/admin/moderation/sanction-ladder", "List sanction ladder", &APIKeyRequest{}, &ModerationLadderResponse{}, http.StatusOK),
		adminModeration(http.MethodPut, "/api/admin/moderation/sanction-ladder", "Replace sanction ladder", &ModerationLadderRequest{}, &ModerationLadderResponse{}, http.StatusOK),
	}
}

// adminModeration creates one protected global moderation operation.
func adminModeration(method string, path string, summary string, request any, body any, status int) operation {
	return operation{method: method, path: path, tag: "Admin Moderation", summary: summary, description: summary + ". Requires X-API-Key; punishment mutations also identify the acting moderator.", request: request, responses: append([]response{jsonResponse(status, body, summary+".")}, errorResponses(http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusConflict, http.StatusInternalServerError)...), secured: true}
}
