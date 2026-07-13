package openapi

import (
	"net/http"
	"time"
)

// AdminPlayerPathRequest documents one player id path.
type AdminPlayerPathRequest struct {
	APIKeyRequest
	// ID identifies the player.
	ID int64 `path:"id" required:"true" minimum:"1"`
}

// AdminPlayerIDRequest documents one player id lookup.
type AdminPlayerIDRequest struct {
	AdminPlayerPathRequest
	// IfNoneMatch supports conditional profile reads.
	IfNoneMatch string `header:"If-None-Match,omitempty"`
}

// AdminPlayerUsernameRequest documents one exact username lookup.
type AdminPlayerUsernameRequest struct {
	APIKeyRequest
	// Username stores the URL-escaped visible player name.
	Username string `path:"username" required:"true" minLength:"1" maxLength:"64" example:"alice"`
	// IfNoneMatch supports conditional profile reads.
	IfNoneMatch string `header:"If-None-Match,omitempty"`
}

// AdminPlayerCreateRequest documents player onboarding input.
type AdminPlayerCreateRequest struct {
	APIKeyRequest
	// IdempotencyKey makes retries safe for 24 hours.
	IdempotencyKey string `header:"Idempotency-Key" required:"true" minLength:"1" maxLength:"128"`
	// Username is the unique visible player name.
	Username string `json:"username" required:"true" minLength:"1" maxLength:"64" example:"alice"`
	// Look is the Nitro avatar figure string.
	Look string `json:"look,omitempty" maxLength:"512" example:"hr-828-61.hd-180-1"`
	// Gender is the Nitro avatar gender code.
	Gender string `json:"gender,omitempty" enum:"M,F" default:"M"`
	// Motto is the initial public player motto.
	Motto string `json:"motto,omitempty" maxLength:"256"`
	// HomeRoomID is the optional initial home room identifier.
	HomeRoomID *int64 `json:"homeRoomId,omitempty" minimum:"1"`
	// AllowNameChange reports whether the player may change username.
	AllowNameChange bool `json:"allowNameChange,omitempty"`
}

// AdminPlayerUpdateRequest documents optional identity and profile changes.
type AdminPlayerUpdateRequest struct {
	AdminPlayerPathRequest
	// Username replaces the visible player name.
	Username *string `json:"username,omitempty" minLength:"1" maxLength:"64"`
	// Look replaces the Nitro avatar figure string.
	Look *string `json:"look,omitempty" maxLength:"512"`
	// Gender replaces the Nitro avatar gender code.
	Gender *string `json:"gender,omitempty" enum:"M,F"`
	// Motto replaces the public player motto.
	Motto *string `json:"motto,omitempty" maxLength:"256"`
	// HomeRoomID replaces the home room identifier.
	HomeRoomID *int64 `json:"homeRoomId,omitempty" minimum:"1"`
	// ClearHomeRoom removes the current home room identifier.
	ClearHomeRoom bool `json:"clearHomeRoom,omitempty"`
	// AllowNameChange replaces the username-change flag.
	AllowNameChange *bool `json:"allowNameChange,omitempty"`
	// BubbleStyle replaces the selected Nitro chat bubble style.
	BubbleStyle *int32 `json:"bubbleStyle,omitempty" minimum:"0"`
	// BlockFriendRequests replaces the friend-request privacy flag.
	BlockFriendRequests *bool `json:"blockFriendRequests,omitempty"`
	// BlockRoomInvites replaces the room-invite privacy flag.
	BlockRoomInvites *bool `json:"blockRoomInvites,omitempty"`
	// BlockFollowing replaces the follow privacy flag.
	BlockFollowing *bool `json:"blockFollowing,omitempty"`
}

// AdminPlayerResponse documents player identity, profile, activity, and club data.
type AdminPlayerResponse struct {
	// ID is the durable player identifier.
	ID int64 `json:"id" required:"true"`
	// Username is the canonical player name.
	Username string `json:"username" required:"true"`
	// Look is the Nitro avatar figure string.
	Look string `json:"look" required:"true"`
	// Gender is the Nitro avatar gender code.
	Gender string `json:"gender" required:"true" enum:"M,F"`
	// Motto is the public player motto.
	Motto string `json:"motto" required:"true"`
	// HomeRoomID is the optional home room identifier.
	HomeRoomID *int64 `json:"homeRoomId,omitempty"`
	// AllowNameChange reports whether username changes are enabled.
	AllowNameChange bool `json:"allowNameChange" required:"true"`
	// BubbleStyle stores the selected chat bubble style.
	BubbleStyle int32 `json:"bubbleStyle" required:"true"`
	// BlockFriendRequests stores friend-request privacy.
	BlockFriendRequests bool `json:"blockFriendRequests" required:"true"`
	// BlockRoomInvites stores room-invite privacy.
	BlockRoomInvites bool `json:"blockRoomInvites" required:"true"`
	// BlockFollowing stores follow privacy.
	BlockFollowing bool `json:"blockFollowing" required:"true"`
	// ClubLevel stores the derived subscription tier.
	ClubLevel int16 `json:"clubLevel" required:"true" minimum:"0" maximum:"2"`
	// ClubExpiresAt stores the optional membership expiration.
	ClubExpiresAt *time.Time `json:"clubExpiresAt,omitempty"`
	// LastLoginAt stores the latest successful login.
	LastLoginAt *time.Time `json:"lastLoginAt,omitempty"`
	// LastSeenAt stores the latest profile activity.
	LastSeenAt *time.Time `json:"lastSeenAt,omitempty"`
	// CreatedAt stores identity creation time.
	CreatedAt time.Time `json:"createdAt" required:"true"`
	// UpdatedAt stores the newest identity or profile mutation time.
	UpdatedAt time.Time `json:"updatedAt" required:"true"`
	// Version combines identity and profile optimistic-lock versions.
	Version string `json:"version" required:"true" example:"2.4"`
}

// adminPlayerOperations returns protected player identity and profile operations.
func adminPlayerOperations() []operation {
	return []operation{
		adminPlayer(http.MethodPost, "/api/admin/players", "Create player and profile", &AdminPlayerCreateRequest{}, &AdminPlayerResponse{}, http.StatusCreated),
		adminPlayer(http.MethodGet, "/api/admin/players/by-username/{username}", "Find player by username", &AdminPlayerUsernameRequest{}, &AdminPlayerResponse{}, http.StatusOK),
		adminPlayer(http.MethodGet, "/api/admin/players/{id}", "Read player profile", &AdminPlayerIDRequest{}, &AdminPlayerResponse{}, http.StatusOK),
		adminPlayer(http.MethodPatch, "/api/admin/players/{id}", "Update player profile", &AdminPlayerUpdateRequest{}, &AdminPlayerResponse{}, http.StatusOK),
		adminPlayer(http.MethodDelete, "/api/admin/players/{id}", "Soft delete player", &AdminPlayerPathRequest{}, nil, http.StatusNoContent),
	}
}

// adminPlayer creates one protected player administration operation.
func adminPlayer(method string, path string, summary string, request any, body any, status int) operation {
	statuses := []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusInternalServerError}
	switch method {
	case http.MethodGet:
		statuses = append(statuses, http.StatusNotFound)
	case http.MethodPost:
		statuses = append(statuses, http.StatusConflict)
	case http.MethodPatch, http.MethodDelete:
		statuses = append(statuses, http.StatusNotFound, http.StatusConflict)
	}
	responses := errorResponses(statuses...)
	if method == http.MethodGet {
		responses = append(responses, emptyResponse(http.StatusNotModified, "Player representation is unchanged."))
	}
	if method == http.MethodPost {
		responses = append(responses, jsonResponse(http.StatusOK, body, "Previously completed idempotent request replayed."))
	}
	if body == nil {
		responses = append([]response{emptyResponse(status, summary+".")}, responses...)
	} else {
		responses = append([]response{jsonResponse(status, body, summary+".")}, responses...)
	}

	return operation{method: method, path: path, tag: "Admin Players", summary: summary,
		description: summary + ".", request: request, responses: responses, secured: true}
}
