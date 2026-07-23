package openapi

import (
	"net/http"

	groupobservability "github.com/niflaot/pixels/internal/realm/group/observability"
	grouproutes "github.com/niflaot/pixels/pkg/http/group/routes"
)

// adminGroupOperations returns protected social-group administration operations.
func adminGroupOperations() []operation {
	create := adminGroup(http.MethodPost, "/api/admin/groups", "Create social group", &GroupCreateRequest{}, &grouproutes.MutationResponse{}, http.StatusCreated)
	create.responses = append([]response{jsonResponse(http.StatusOK, &grouproutes.MutationResponse{}, "Replay existing social group creation.")}, create.responses...)
	return []operation{
		adminGroup(http.MethodGet, "/api/admin/groups", "List social groups", &GroupListRequest{}, &grouproutes.GroupListResponse{}, http.StatusOK),
		create,
		adminGroup(http.MethodGet, "/api/admin/groups/metrics", "Read social group telemetry", &GroupActorRequest{}, &groupobservability.Snapshot{}, http.StatusOK),
		adminGroup(http.MethodGet, "/api/admin/groups/players/{playerId}", "List player social groups", &GroupPlayerReadRequest{}, &[]GroupMembershipResponse{}, http.StatusOK),
		adminGroup(http.MethodPut, "/api/admin/groups/players/{playerId}/favorite", "Set player favorite group", &GroupFavoriteRequest{}, nil, http.StatusNoContent),
		adminGroup(http.MethodGet, "/api/admin/groups/{groupId}", "Read social group", &GroupReadRequest{}, &grouproutes.GroupResponse{}, http.StatusOK),
		adminGroup(http.MethodPatch, "/api/admin/groups/{groupId}", "Update social group", &GroupUpdateRequest{}, &grouproutes.MutationResponse{}, http.StatusOK),
		adminGroup(http.MethodDelete, "/api/admin/groups/{groupId}", "Deactivate social group", &GroupVersionRequest{}, &grouproutes.MutationResponse{}, http.StatusOK),
		adminGroup(http.MethodPost, "/api/admin/groups/{groupId}/restore", "Restore social group", &GroupRoomRequest{}, &grouproutes.MutationResponse{}, http.StatusOK),
		adminGroup(http.MethodPut, "/api/admin/groups/{groupId}/badge", "Replace social group badge", &GroupBadgeRequest{}, &grouproutes.MutationResponse{}, http.StatusOK),
		adminGroup(http.MethodPut, "/api/admin/groups/{groupId}/home-room", "Rebind social group home room", &GroupRoomRequest{}, &grouproutes.MutationResponse{}, http.StatusOK),
		adminGroup(http.MethodPost, "/api/admin/groups/{groupId}/owner", "Transfer social group owner", &GroupOwnerRequest{}, &grouproutes.MutationResponse{}, http.StatusOK),
		adminGroup(http.MethodGet, "/api/admin/groups/{groupId}/members", "List social group members", &GroupMemberListRequest{}, &grouproutes.MemberPageResponse{}, http.StatusOK),
		adminGroup(http.MethodPost, "/api/admin/groups/{groupId}/members", "Add social group member", &GroupMemberAddRequest{}, &grouproutes.MutationResponse{}, http.StatusOK),
		adminGroup(http.MethodPatch, "/api/admin/groups/{groupId}/members/{playerId}", "Change social group member role", &GroupMemberRoleRequest{}, &grouproutes.MutationResponse{}, http.StatusOK),
		adminGroup(http.MethodDelete, "/api/admin/groups/{groupId}/members/{playerId}", "Remove social group member", &GroupMemberMutationRequest{}, &grouproutes.MutationResponse{}, http.StatusOK),
		adminGroup(http.MethodGet, "/api/admin/groups/{groupId}/requests", "List social group requests", &GroupRequestListRequest{}, &[]GroupRequestResponse{}, http.StatusOK),
		adminGroup(http.MethodPost, "/api/admin/groups/{groupId}/requests/{playerId}/accept", "Accept social group request", &GroupMemberMutationRequest{}, &grouproutes.MutationResponse{}, http.StatusOK),
		adminGroup(http.MethodDelete, "/api/admin/groups/{groupId}/requests/{playerId}", "Decline social group request", &GroupMemberMutationRequest{}, &grouproutes.MutationResponse{}, http.StatusOK),
		adminGroup(http.MethodPost, "/api/admin/groups/{groupId}/requests/approve-all", "Approve social group requests", &GroupMutationRequest{}, &grouproutes.MutationResponse{}, http.StatusOK),
		adminGroup(http.MethodGet, "/api/admin/groups/{groupId}/forum/settings", "Read social group forum settings", &GroupReadRequest{}, &GroupForumSettingsResponse{}, http.StatusOK),
		adminGroup(http.MethodPut, "/api/admin/groups/{groupId}/forum/settings", "Update social group forum settings", &GroupForumSettingsRequest{}, &grouproutes.MutationResponse{}, http.StatusOK),
		adminGroup(http.MethodGet, "/api/admin/groups/{groupId}/forum/threads", "List social group forum threads", &GroupForumPageRequest{}, &GroupForumThreadsResponse{}, http.StatusOK),
		adminGroup(http.MethodGet, "/api/admin/groups/{groupId}/forum/threads/{threadId}", "Read social group forum thread", &GroupForumThreadReadRequest{}, &grouproutes.ForumThreadResponse{}, http.StatusOK),
		adminGroup(http.MethodPatch, "/api/admin/groups/{groupId}/forum/threads/{threadId}", "Moderate social group forum thread", &GroupForumThreadRequest{}, &GroupForumThreadResponse{}, http.StatusOK),
		adminGroup(http.MethodPatch, "/api/admin/groups/{groupId}/forum/posts/{postId}", "Moderate social group forum post", &GroupForumPostRequest{}, &GroupForumPostResponse{}, http.StatusOK),
	}
}

// adminGroup creates one protected social-group operation.
func adminGroup(method string, path string, summary string, request any, body any, status int) operation {
	responses := errorResponses(http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusConflict, http.StatusUnprocessableEntity, http.StatusTooManyRequests, http.StatusInternalServerError)
	if body == nil {
		responses = append([]response{emptyResponse(status, summary+".")}, responses...)
	} else {
		responses = append([]response{jsonResponse(status, body, summary+".")}, responses...)
	}
	return operation{method: method, path: path, tag: "Groups", summary: summary, description: summary + ". Reads require X-Actor-Player-ID; mutations require actorPlayerId, reason, and optimistic version where applicable.", request: request, responses: responses, secured: true}
}
