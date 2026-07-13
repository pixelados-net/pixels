// Package routes exposes the private administrative player HTTP contract.
package routes

import "time"

// CreateRequest contains player onboarding input from an administrative client.
type CreateRequest struct {
	// Username is the unique visible player name.
	Username string `json:"username"`
	// Look is the Nitro avatar figure string.
	Look string `json:"look"`
	// Gender is the Nitro avatar gender code.
	Gender string `json:"gender"`
	// Motto is the initial public player motto.
	Motto string `json:"motto,omitempty"`
	// HomeRoomID is the optional initial home room identifier.
	HomeRoomID *int64 `json:"homeRoomId,omitempty"`
	// AllowNameChange reports whether the player may change username.
	AllowNameChange bool `json:"allowNameChange,omitempty"`
}

// UpdateRequest contains optional administrative identity and profile changes.
type UpdateRequest struct {
	// Username replaces the visible player name.
	Username *string `json:"username,omitempty"`
	// Look replaces the Nitro avatar figure string.
	Look *string `json:"look,omitempty"`
	// Gender replaces the Nitro avatar gender code.
	Gender *string `json:"gender,omitempty"`
	// Motto replaces the public player motto.
	Motto *string `json:"motto,omitempty"`
	// HomeRoomID replaces the home room identifier.
	HomeRoomID *int64 `json:"homeRoomId,omitempty"`
	// ClearHomeRoom removes the current home room identifier.
	ClearHomeRoom bool `json:"clearHomeRoom,omitempty"`
	// AllowNameChange replaces the username-change flag.
	AllowNameChange *bool `json:"allowNameChange,omitempty"`
	// BubbleStyle replaces the selected Nitro chat bubble style.
	BubbleStyle *int32 `json:"bubbleStyle,omitempty"`
	// BlockFriendRequests replaces the friend-request privacy flag.
	BlockFriendRequests *bool `json:"blockFriendRequests,omitempty"`
	// BlockRoomInvites replaces the room-invite privacy flag.
	BlockRoomInvites *bool `json:"blockRoomInvites,omitempty"`
	// BlockFollowing replaces the follow privacy flag.
	BlockFollowing *bool `json:"blockFollowing,omitempty"`
}

// Response contains player identity and profile data required by administrative clients.
type Response struct {
	// ID is the durable player identifier.
	ID int64 `json:"id"`
	// Username is the canonical player name.
	Username string `json:"username"`
	// Look is the Nitro avatar figure string.
	Look string `json:"look"`
	// Gender is the Nitro avatar gender code.
	Gender string `json:"gender"`
	// Motto is the public player motto.
	Motto string `json:"motto"`
	// HomeRoomID is the optional home room identifier.
	HomeRoomID *int64 `json:"homeRoomId,omitempty"`
	// AllowNameChange reports whether the player may change username.
	AllowNameChange bool `json:"allowNameChange"`
	// BubbleStyle stores the selected Nitro chat bubble style.
	BubbleStyle int32 `json:"bubbleStyle"`
	// BlockFriendRequests reports whether incoming friend requests are disabled.
	BlockFriendRequests bool `json:"blockFriendRequests"`
	// BlockRoomInvites reports whether incoming room invites are disabled.
	BlockRoomInvites bool `json:"blockRoomInvites"`
	// BlockFollowing reports whether friends may follow this player.
	BlockFollowing bool `json:"blockFollowing"`
	// ClubLevel stores the derived subscription tier.
	ClubLevel int16 `json:"clubLevel"`
	// AllowTrade reports whether direct trading is enabled.
	AllowTrade bool `json:"allowTrade"`
	// ClubExpiresAt stores the optional current subscription expiration.
	ClubExpiresAt *time.Time `json:"clubExpiresAt,omitempty"`
	// LastLoginAt is the most recent successful hotel login.
	LastLoginAt *time.Time `json:"lastLoginAt,omitempty"`
	// LastSeenAt is the most recent profile activity time.
	LastSeenAt *time.Time `json:"lastSeenAt,omitempty"`
	// CreatedAt stores player identity creation time.
	CreatedAt time.Time `json:"createdAt"`
	// UpdatedAt stores the newest identity or profile mutation time.
	UpdatedAt time.Time `json:"updatedAt"`
	// Version is the combined identity and profile representation version.
	Version string `json:"version"`
}

// idempotencyRecord stores one create request state in Redis.
type idempotencyRecord struct {
	// State reports whether the request is pending or complete.
	State string `json:"state"`
	// RequestHash binds the key to one canonical request body.
	RequestHash string `json:"requestHash"`
	// Username supports recovery after a lost HTTP response.
	Username string `json:"username"`
	// Response stores the replayable completed result.
	Response *Response `json:"response,omitempty"`
}
