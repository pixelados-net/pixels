package openapi

import "net/http"

// roomVoteOperations returns protected room vote operations.
func roomVoteOperations() []operation {
	return []operation{
		adminRoomVote(http.MethodGet, "/api/admin/room-votes/status", "Read room vote status", &RoomVoteStatusRequest{}, &RoomVoteStatusResponse{}),
		adminRoomVote(http.MethodGet, "/api/admin/room-votes/list", "List room votes", &RoomVoteListRequest{}, &RoomVoteListResponse{}),
		adminRoomVote(http.MethodPost, "/api/admin/room-votes/cast", "Cast room upvote", &RoomVoteCastRequest{}, &RoomVoteCastResponse{}),
	}
}

// adminRoomVote creates a protected room vote operation.
func adminRoomVote(method string, path string, summary string, request any, body any) operation {
	return operation{
		method: method, path: path, tag: "Admin Room Votes", summary: summary,
		description: summary + ". Votes are permanent and idempotent per player and room.", request: request,
		responses: append([]response{jsonResponse(http.StatusOK, body, summary+".")},
			errorResponses(http.StatusBadRequest, http.StatusUnauthorized, http.StatusNotFound, http.StatusInternalServerError)...),
		secured: true,
	}
}
