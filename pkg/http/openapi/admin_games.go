package openapi

import (
	"net/http"

	gamecenterrecord "github.com/niflaot/pixels/internal/realm/gamecenter/record"
	progressionpoll "github.com/niflaot/pixels/internal/realm/progression/poll"
	roomgames "github.com/niflaot/pixels/internal/realm/room/world/games"
	gameroutes "github.com/niflaot/pixels/pkg/http/game/routes"
)

// adminGamesOperations returns protected game administration operations.
func adminGamesOperations() []operation {
	return []operation{
		adminGames(http.MethodGet, "/api/admin/games/center", "List external games", &GamesActorRequest{}, &[]gamecenterrecord.Game{}, http.StatusOK),
		adminGames(http.MethodPost, "/api/admin/games/center", "Create external game", &gameroutes.CenterRequest{}, &gamecenterrecord.Game{}, http.StatusCreated),
		adminGames(http.MethodPatch, "/api/admin/games/center/{id}", "Update external game", &GamesCenterMutationRequest{}, &gamecenterrecord.Game{}, http.StatusOK),
		adminGames(http.MethodDelete, "/api/admin/games/center/{id}", "Disable external game", &GamesIDMutationRequest{}, nil, http.StatusNoContent),
		adminGames(http.MethodPut, "/api/admin/games/center/{id}/scores/{playerId}", "Upsert weekly game score", &GamesScoreMutationRequest{}, nil, http.StatusNoContent),
		adminGames(http.MethodGet, "/api/admin/games/polls", "List room polls", &GamesActorRequest{}, &[]progressionpoll.Definition{}, http.StatusOK),
		adminGames(http.MethodPost, "/api/admin/games/polls", "Create room poll", &gameroutes.PollRequest{}, &progressionpoll.Definition{}, http.StatusCreated),
		adminGames(http.MethodPatch, "/api/admin/games/polls/{id}", "Update room poll", &GamesPollMutationRequest{}, &progressionpoll.Definition{}, http.StatusOK),
		adminGames(http.MethodDelete, "/api/admin/games/polls/{id}", "Disable room poll", &GamesIDMutationRequest{}, nil, http.StatusNoContent),
		adminGames(http.MethodPut, "/api/admin/games/polls/{id}/room", "Assign poll room", &GamesPollRoomMutationRequest{}, &GamesVersionResponse{}, http.StatusOK),
		adminGames(http.MethodGet, "/api/admin/games/rooms/{roomId}/scores", "List room game scores", &GamesRoomScoresRequest{}, &roomgames.ScorePage{}, http.StatusOK),
		adminGames(http.MethodPost, "/api/admin/games/reload", "Reload game caches", &gameroutes.AuditRequest{}, &GamesReloadResponse{}, http.StatusOK),
		adminGames(http.MethodGet, "/api/admin/games/metrics", "Read game metrics", &GamesActorRequest{}, &GamesMetricsResponse{}, http.StatusOK),
	}
}

// adminGames creates one protected game administration operation.
func adminGames(method string, path string, summary string, request any, body any, status int) operation {
	responses := errorResponses(http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusConflict, http.StatusUnprocessableEntity, http.StatusInternalServerError)
	if body == nil {
		responses = append([]response{emptyResponse(status, summary+".")}, responses...)
	} else {
		responses = append([]response{jsonResponse(status, body, summary+".")}, responses...)
	}
	return operation{method: method, path: path, tag: "Games", summary: summary, description: summary + ". Mutations require actorPlayerId and reason and are audited atomically.", request: request, responses: responses, secured: true}
}

// GamesActorRequest documents administrative read attribution.
type GamesActorRequest struct {
	// ActorPlayerID identifies the authorized administrator.
	ActorPlayerID int64 `header:"X-Actor-Player-ID" required:"true"`
}

// GamesIDMutationRequest documents one audited numeric mutation.
type GamesIDMutationRequest struct {
	// ID identifies the mutated entity.
	ID int64 `path:"id"`
	// AuditRequest documents administrative attribution.
	gameroutes.AuditRequest
}

// GamesCenterMutationRequest documents one center update.
type GamesCenterMutationRequest struct {
	// ID identifies the external game.
	ID int64 `path:"id"`
	// CenterRequest documents the replacement registration.
	gameroutes.CenterRequest
}

// GamesPollMutationRequest documents one poll update.
type GamesPollMutationRequest struct {
	// ID identifies the durable poll.
	ID int64 `path:"id"`
	// PollRequest documents the replacement poll.
	gameroutes.PollRequest
}

// GamesPollRoomMutationRequest documents one poll room assignment.
type GamesPollRoomMutationRequest struct {
	// ID identifies the durable poll.
	ID int64 `path:"id"`
	// PollRoomRequest documents the room assignment.
	gameroutes.PollRoomRequest
}

// GamesScoreMutationRequest documents one manual weekly score.
type GamesScoreMutationRequest struct {
	// ID identifies the external game.
	ID int64 `path:"id"`
	// PlayerID identifies the scored player.
	PlayerID int64 `path:"playerId"`
	// ScoreRequest documents the weekly result.
	gameroutes.ScoreRequest
}

// GamesRoomScoresRequest documents one paginated score read.
type GamesRoomScoresRequest struct {
	// GamesActorRequest documents administrative attribution.
	GamesActorRequest
	// RoomID identifies the queried room.
	RoomID int64 `path:"roomId"`
	// BeforeID stores the exclusive pagination cursor.
	BeforeID int64 `query:"beforeId"`
	// Limit bounds the requested result page.
	Limit int `query:"limit"`
}

// GamesVersionResponse documents one optimistic version response.
type GamesVersionResponse struct {
	// Version stores the updated optimistic version.
	Version int64 `json:"version"`
	// RoomID stores the assigned room or zero.
	RoomID int64 `json:"roomId"`
}

// GamesReloadResponse documents a successful cache reload.
type GamesReloadResponse struct {
	// Reloaded reports a successful atomic cache replacement.
	Reloaded bool `json:"reloaded"`
}

// GamesMetricsResponse documents lock-free room, poll, and Game Center metrics.
type GamesMetricsResponse struct {
	// Rooms stores room-game counters.
	Rooms roomgames.MetricsSnapshot `json:"rooms"`
	// PollsResponded stores durable completed poll count.
	PollsResponded uint64 `json:"pollsResponded"`
	// GameCenterLaunches stores successful external launches.
	GameCenterLaunches uint64 `json:"gameCenterLaunches"`
}
