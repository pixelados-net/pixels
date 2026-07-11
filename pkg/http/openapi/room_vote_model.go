package openapi

import "time"

// RoomVoteStatusRequest contains room vote status query parameters.
type RoomVoteStatusRequest struct {
	APIKeyRequest
	// RoomID identifies the rated room.
	RoomID int64 `query:"roomId" required:"true" minimum:"1"`
	// PlayerID identifies the inspected player.
	PlayerID int64 `query:"playerId" required:"true" minimum:"1"`
}

// RoomVoteListRequest contains room vote list filters.
type RoomVoteListRequest struct {
	APIKeyRequest
	// RoomID identifies the rated room.
	RoomID int64 `query:"roomId" required:"true" minimum:"1"`
	// PlayerID optionally filters one voter.
	PlayerID *int64 `query:"playerId" minimum:"1"`
	// Before optionally filters votes before an RFC3339 timestamp.
	Before *time.Time `query:"before"`
	// Limit bounds returned votes.
	Limit int `query:"limit" minimum:"1" maximum:"200"`
}

// RoomVoteCastRequest contains one administrative upvote.
type RoomVoteCastRequest struct {
	APIKeyRequest
	// RoomID identifies the rated room.
	RoomID int64 `json:"roomId" required:"true" minimum:"1"`
	// PlayerID identifies the voter.
	PlayerID int64 `json:"playerId" required:"true" minimum:"1"`
}

// RoomVoteStatusResponse describes score and eligibility.
type RoomVoteStatusResponse struct {
	// RoomID identifies the rated room.
	RoomID int64 `json:"roomId"`
	// PlayerID identifies the inspected player.
	PlayerID int64 `json:"playerId"`
	// Score stores current room score.
	Score int `json:"score"`
	// CanVote reports whether the player may vote.
	CanVote bool `json:"canVote"`
	// Voted reports whether the player already voted.
	Voted bool `json:"voted"`
}

// RoomVoteCastResponse describes a room vote mutation.
type RoomVoteCastResponse struct {
	// RoomID identifies the rated room.
	RoomID int64 `json:"roomId"`
	// PlayerID identifies the voter.
	PlayerID int64 `json:"playerId"`
	// Score stores resulting room score.
	Score int `json:"score"`
	// Inserted reports whether a new vote was created.
	Inserted bool `json:"inserted"`
}

// RoomVoteResponse describes one durable vote.
type RoomVoteResponse struct {
	// RoomID identifies the rated room.
	RoomID int64 `json:"roomId"`
	// PlayerID identifies the voter.
	PlayerID int64 `json:"playerId"`
	// CreatedAt stores when the vote was cast.
	CreatedAt time.Time `json:"createdAt"`
}

// RoomVoteListResponse contains a bounded vote page.
type RoomVoteListResponse struct {
	// Total stores returned vote count.
	Total int `json:"total"`
	// Items stores durable votes.
	Items []RoomVoteResponse `json:"items"`
}
