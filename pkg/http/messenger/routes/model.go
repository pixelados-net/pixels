// Package routes contains protected messenger administration routes.
package routes

// FriendResponse describes one friend card.
type FriendResponse struct {
	// PlayerID identifies the friend.
	PlayerID int64 `json:"playerId"`
	// Username stores the visible username.
	Username string `json:"username"`
	// Look stores the avatar figure string.
	Look string `json:"look"`
	// Gender stores the avatar gender code.
	Gender string `json:"gender"`
	// Online reports current presence.
	Online bool `json:"online"`
	// InRoom reports follow availability.
	InRoom bool `json:"inRoom"`
	// Relation stores the unilateral relationship marker.
	Relation int16 `json:"relation"`
}

// FriendsResponse contains one player's friend list.
type FriendsResponse struct {
	// PlayerID identifies the list owner.
	PlayerID int64 `json:"playerId"`
	// Friends stores current friend cards.
	Friends []FriendResponse `json:"friends"`
}

// RequestResponse describes one pending friend request.
type RequestResponse struct {
	// FromPlayerID identifies the requester.
	FromPlayerID int64 `json:"fromPlayerId"`
	// ToPlayerID identifies the recipient.
	ToPlayerID int64 `json:"toPlayerId"`
	// CreatedAt stores the RFC3339 creation time.
	CreatedAt string `json:"createdAt"`
}

// RequestsResponse contains incoming and outgoing requests.
type RequestsResponse struct {
	// PlayerID identifies the queried player.
	PlayerID int64 `json:"playerId"`
	// Incoming stores received requests.
	Incoming []RequestResponse `json:"incoming"`
	// Outgoing stores sent requests.
	Outgoing []RequestResponse `json:"outgoing"`
}

// PrivacyRequest contains optional messenger privacy changes.
type PrivacyRequest struct {
	// BlockFriendRequests changes friend-request privacy when supplied.
	BlockFriendRequests *bool `json:"blockFriendRequests"`
	// BlockRoomInvites changes room-invite privacy when supplied.
	BlockRoomInvites *bool `json:"blockRoomInvites"`
	// BlockFollowing changes follow privacy when supplied.
	BlockFollowing *bool `json:"blockFollowing"`
}

// PrivacyResponse describes persisted messenger privacy.
type PrivacyResponse struct {
	// PlayerID identifies the profile owner.
	PlayerID int64 `json:"playerId"`
	// BlockFriendRequests stores friend-request privacy.
	BlockFriendRequests bool `json:"blockFriendRequests"`
	// BlockRoomInvites stores room-invite privacy.
	BlockRoomInvites bool `json:"blockRoomInvites"`
	// BlockFollowing stores follow privacy.
	BlockFollowing bool `json:"blockFollowing"`
}

// MutationResponse describes one administrative friendship removal.
type MutationResponse struct {
	// PlayerID identifies the friendship owner.
	PlayerID int64 `json:"playerId"`
	// FriendPlayerID identifies the removed friend.
	FriendPlayerID int64 `json:"friendPlayerId"`
	// Removed reports whether a friendship existed.
	Removed bool `json:"removed"`
}
