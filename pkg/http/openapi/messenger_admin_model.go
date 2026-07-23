package openapi

// MessengerPlayerRequest identifies one messenger player.
type MessengerPlayerRequest struct {
	APIKeyRequest
	// PlayerID identifies the messenger owner.
	PlayerID int64 `path:"playerId" required:"true" minimum:"1"`
}

// MessengerFriendRequest identifies one friendship.
type MessengerFriendRequest struct {
	MessengerPlayerRequest
	// FriendID identifies the friend.
	FriendID int64 `path:"friendId" required:"true" minimum:"1"`
}

// MessengerPrivacyRequest patches messenger privacy.
type MessengerPrivacyRequest struct {
	MessengerPlayerRequest
	// BlockFriendRequests changes friend-request privacy.
	BlockFriendRequests *bool `json:"blockFriendRequests,omitempty"`
	// BlockRoomInvites changes room-invite privacy.
	BlockRoomInvites *bool `json:"blockRoomInvites,omitempty"`
	// BlockFollowing changes follow privacy.
	BlockFollowing *bool `json:"blockFollowing,omitempty"`
}

// MessengerFriend describes one friend card.
type MessengerFriend struct {
	// PlayerID identifies the friend.
	PlayerID int64 `json:"playerId" required:"true"`
	// Username stores the visible name.
	Username string `json:"username" required:"true"`
	// Look stores the avatar figure string.
	Look string `json:"look" required:"true"`
	// Gender stores the avatar gender code.
	Gender string `json:"gender" required:"true" enum:"M,F"`
	// Online reports current presence.
	Online bool `json:"online" required:"true"`
	// InRoom reports current room presence.
	InRoom bool `json:"inRoom" required:"true"`
	// Relation stores the unilateral marker.
	Relation int16 `json:"relation" required:"true"`
}

// MessengerFriendsResponse contains one friend list.
type MessengerFriendsResponse struct {
	// PlayerID identifies the list owner.
	PlayerID int64 `json:"playerId" required:"true"`
	// Friends stores friend cards.
	Friends []MessengerFriend `json:"friends" required:"true"`
}

// MessengerPendingRequest describes one pending request.
type MessengerPendingRequest struct {
	// FromPlayerID identifies the requester.
	FromPlayerID int64 `json:"fromPlayerId" required:"true"`
	// ToPlayerID identifies the recipient.
	ToPlayerID int64 `json:"toPlayerId" required:"true"`
	// CreatedAt stores creation time.
	CreatedAt string `json:"createdAt" required:"true" format:"date-time"`
}

// MessengerRequestsResponse contains incoming and outgoing requests.
type MessengerRequestsResponse struct {
	// PlayerID identifies the queried player.
	PlayerID int64 `json:"playerId" required:"true"`
	// Incoming stores received requests.
	Incoming []MessengerPendingRequest `json:"incoming" required:"true"`
	// Outgoing stores sent requests.
	Outgoing []MessengerPendingRequest `json:"outgoing" required:"true"`
}

// MessengerPrivacyResponse contains persisted privacy state.
type MessengerPrivacyResponse struct {
	// PlayerID identifies the profile owner.
	PlayerID int64 `json:"playerId" required:"true"`
	// BlockFriendRequests stores friend-request privacy.
	BlockFriendRequests bool `json:"blockFriendRequests" required:"true"`
	// BlockRoomInvites stores room-invite privacy.
	BlockRoomInvites bool `json:"blockRoomInvites" required:"true"`
	// BlockFollowing stores follow privacy.
	BlockFollowing bool `json:"blockFollowing" required:"true"`
}

// MessengerMutationResponse confirms administrative removal.
type MessengerMutationResponse struct {
	// PlayerID identifies the friendship owner.
	PlayerID int64 `json:"playerId" required:"true"`
	// FriendPlayerID identifies the removed friend.
	FriendPlayerID int64 `json:"friendPlayerId" required:"true"`
	// Removed reports whether a friendship existed.
	Removed bool `json:"removed" required:"true"`
}
