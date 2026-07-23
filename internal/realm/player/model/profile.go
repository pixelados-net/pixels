package model

import (
	"time"

	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// Gender is the persisted avatar gender code expected by Nitro.
type Gender string

const (
	// GenderMale is the male avatar gender code.
	GenderMale Gender = "M"

	// GenderFemale is the female avatar gender code.
	GenderFemale Gender = "F"
)

// Valid reports whether the gender is supported by the current player model.
func (gender Gender) Valid() bool {
	return gender == GenderMale || gender == GenderFemale
}

// Profile contains durable player presentation fields.
type Profile struct {
	// PlayerID is the owning player identifier.
	PlayerID int64

	// Look is the Nitro avatar figure string.
	Look string

	// Gender is the Nitro avatar gender code.
	Gender Gender

	// Motto is the public player motto.
	Motto string

	// HomeRoomID is the optional default home room identifier.
	HomeRoomID *int64

	// AllowNameChange reports whether the player can change username.
	AllowNameChange bool

	// BubbleStyle stores the validated Nitro chat bubble style.
	BubbleStyle int32

	// BlockFriendRequests reports whether incoming friend requests are disabled.
	BlockFriendRequests bool

	// BlockRoomInvites reports whether incoming messenger room invites are disabled.
	BlockRoomInvites bool

	// BlockFollowing reports whether friends may follow this player to a room.
	BlockFollowing bool

	// Timestamps contains durable record timestamps.
	sharedmodel.Timestamps

	// Version contains optimistic locking state.
	sharedmodel.Version
}

// Created reports whether the profile has an owning player id.
func (profile Profile) Created() bool {
	return profile.PlayerID > 0
}

// UpdatedAfter reports whether the profile was updated after a time.
func (profile Profile) UpdatedAfter(value time.Time) bool {
	return profile.UpdatedAt.After(value)
}
