// Package record contains messenger persistence contracts and protocol-neutral records.
package record

import "time"

// Relation identifies one player's unilateral relationship marker.
type Relation int16

const (
	// RelationNone clears a relationship marker.
	RelationNone Relation = iota
	// RelationHeart displays Nitro's heart marker.
	RelationHeart
	// RelationSmile displays Nitro's smile marker.
	RelationSmile
	// RelationBobba displays Nitro's bobba marker.
	RelationBobba
)

// Valid reports whether the relation is supported by Nitro.
func (relation Relation) Valid() bool { return relation >= RelationNone && relation <= RelationBobba }

// Friendship stores one directional friendship record.
type Friendship struct {
	// PlayerID stores the player whose list contains the friendship.
	PlayerID int64
	// FriendPlayerID stores the listed friend.
	FriendPlayerID int64
	// Relation stores the owner's unilateral marker.
	Relation Relation
	// CategoryID stores a future optional friend-folder identifier.
	CategoryID *int64
	// CreatedAt stores when the friendship was accepted.
	CreatedAt time.Time
}

// Request stores one pending friend request.
type Request struct {
	// FromPlayerID stores the requester.
	FromPlayerID int64
	// ToPlayerID stores the recipient.
	ToPlayerID int64
	// CreatedAt stores when the request was made.
	CreatedAt time.Time
}

// IgnoredPlayer stores one directional ignored-user projection.
type IgnoredPlayer struct {
	// PlayerID identifies the ignored player.
	PlayerID int64
	// Username stores the protocol-visible username.
	Username string
}

// RelationshipSummary stores one public relationship category summary.
type RelationshipSummary struct {
	// Relation identifies the relationship category.
	Relation Relation
	// Count stores friends assigned to this category.
	Count int32
	// Sample stores one representative friend.
	Sample IgnoredPlayer
	// SampleLook stores the representative friend's figure.
	SampleLook string
}

// Card stores one Nitro messenger friend projection.
type Card struct {
	// ID stores the friend player id.
	ID int64
	// Username stores the visible username.
	Username string
	// Gender stores Nitro's numeric gender value.
	Gender int32
	// Online reports whether the player has an active session.
	Online bool
	// FollowingAllowed reports whether the player may be followed to a room.
	FollowingAllowed bool
	// Look stores the avatar figure.
	Look string
	// CategoryID stores the friend folder id or zero.
	CategoryID int32
	// Motto stores the public motto.
	Motto string
	// Relation stores the viewer's unilateral marker.
	Relation Relation
	// BlockFollowing stores the durable privacy decision used during live projection.
	BlockFollowing bool
}

// SearchResult stores one public player search projection.
type SearchResult struct {
	// PlayerID stores the matching player id.
	PlayerID int64
	// Username stores the matching username.
	Username string
	// Motto stores the public motto.
	Motto string
	// Look stores the avatar figure.
	Look string
	// Gender stores Nitro's numeric gender value.
	Gender int32
	// BlockFollowing stores the durable privacy decision used during projection.
	BlockFollowing bool
	// Online reports current live presence after projection.
	Online bool
	// CanFollow reports current follow availability after projection.
	CanFollow bool
}
