package votes

import "time"

// CastRequest contains one administrative room vote.
type CastRequest struct {
	// RoomID identifies the rated room.
	RoomID int64 `json:"roomId"`
	// PlayerID identifies the voter.
	PlayerID int64 `json:"playerId"`
}

// CastResponse describes an administrative room vote result.
type CastResponse struct {
	// RoomID identifies the rated room.
	RoomID int64 `json:"roomId"`
	// PlayerID identifies the voter.
	PlayerID int64 `json:"playerId"`
	// Score stores the resulting room score.
	Score int `json:"score"`
	// Inserted reports whether the permanent vote was newly created.
	Inserted bool `json:"inserted"`
}

// StatusResponse describes room score and player eligibility.
type StatusResponse struct {
	// RoomID identifies the rated room.
	RoomID int64 `json:"roomId"`
	// PlayerID identifies the inspected player.
	PlayerID int64 `json:"playerId"`
	// Score stores the current room score.
	Score int `json:"score"`
	// CanVote reports whether the player may vote.
	CanVote bool `json:"canVote"`
	// Voted reports whether the player already voted.
	Voted bool `json:"voted"`
}

// VoteResponse describes one durable room vote.
type VoteResponse struct {
	// RoomID identifies the rated room.
	RoomID int64 `json:"roomId"`
	// PlayerID identifies the voter.
	PlayerID int64 `json:"playerId"`
	// CreatedAt stores when the vote was cast.
	CreatedAt time.Time `json:"createdAt"`
}

// ListResponse contains a bounded room vote page.
type ListResponse struct {
	// Total stores the returned record count.
	Total int `json:"total"`
	// Items stores durable room votes.
	Items []VoteResponse `json:"items"`
}
