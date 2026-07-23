// Package record defines durable social-group records and persistence boundaries.
package record

import "time"

// State controls public membership admission.
type State int16

const (
	// Regular admits users immediately.
	Regular State = iota
	// Exclusive creates membership requests.
	Exclusive
	// Private rejects public membership requests.
	Private
)

// Valid reports whether state is a Nitro-supported group state.
func (state State) Valid() bool { return state >= Regular && state <= Private }

// Role identifies one active social-group membership role.
type Role int16

const (
	// Owner is the single protected group owner.
	Owner Role = iota
	// Admin may manage ordinary members and requests.
	Admin
	// Member is an ordinary active member.
	Member
	// Requested is a wire-only pending projection rank.
	Requested
)

// Valid reports whether role is durable.
func (role Role) Valid() bool { return role >= Owner && role <= Member }

// Policy identifies a group-forum access threshold.
type Policy int16

const (
	// Everyone allows authenticated hotel users.
	Everyone Policy = iota
	// Members allows active group members.
	Members
	// Admins allows group owners and admins.
	Admins
	// Owners allows only the group owner.
	Owners
)

// Valid reports whether policy is supported.
func (policy Policy) Valid() bool { return policy >= Everyone && policy <= Owners }

// BadgeKind identifies a base or symbol element.
type BadgeKind int16

const (
	// BadgeBase is the mandatory first badge part.
	BadgeBase BadgeKind = iota
	// BadgeSymbol is an optional overlay.
	BadgeSymbol
)

// ColorFamily identifies one badge-editor color collection.
type ColorFamily int16

const (
	// BaseColor colors badge base elements.
	BaseColor ColorFamily = iota
	// SymbolColor colors badge symbol elements.
	SymbolColor
	// BackgroundColor colors both group furniture surfaces.
	BackgroundColor
)

// Group contains active or retained social-group metadata.
type Group struct {
	// ID identifies the group.
	ID int64 `json:"id"`
	// OwnerPlayerID identifies the exactly-one owner.
	OwnerPlayerID int64 `json:"ownerPlayerId"`
	// OwnerName stores the current owner display name.
	OwnerName string `json:"ownerName"`
	// Name stores the visible group title.
	Name string `json:"name"`
	// Description stores plain-text group information.
	Description string `json:"description"`
	// HomeRoomID identifies the authoritative headquarters.
	HomeRoomID int64 `json:"homeRoomId"`
	// HomeRoomName stores the current room display name.
	HomeRoomName string `json:"homeRoomName"`
	// State controls join behavior.
	State State `json:"state"`
	// CanMembersDecorate grants effective furniture rights to members.
	CanMembersDecorate bool `json:"canMembersDecorate"`
	// ColorA identifies the primary editor color.
	ColorA int32 `json:"colorA"`
	// ColorB identifies the secondary editor color.
	ColorB int32 `json:"colorB"`
	// ColorAHex stores the resolved primary RGB value.
	ColorAHex string `json:"colorAHex"`
	// ColorBHex stores the resolved secondary RGB value.
	ColorBHex string `json:"colorBHex"`
	// BadgeCode stores the validated compiled badge cache.
	BadgeCode string `json:"badgeCode"`
	// ForumEnabled reports whether the forum entitlement is active.
	ForumEnabled bool `json:"forumEnabled"`
	// ReadPolicy controls forum reads.
	ReadPolicy Policy `json:"readPolicy"`
	// PostMessagePolicy controls replies.
	PostMessagePolicy Policy `json:"postMessagePolicy"`
	// PostThreadPolicy controls thread creation.
	PostThreadPolicy Policy `json:"postThreadPolicy"`
	// ModeratePolicy controls group-role moderation.
	ModeratePolicy Policy `json:"moderatePolicy"`
	// MemberCount stores the transactionally maintained active count.
	MemberCount int32 `json:"memberCount"`
	// PendingCount stores the transactionally maintained request count.
	PendingCount int32 `json:"pendingCount"`
	// ThreadCount stores retained forum thread count.
	ThreadCount int32 `json:"threadCount"`
	// PostCount stores retained forum post count.
	PostCount int32 `json:"postCount"`
	// CreatedAt stores creation time.
	CreatedAt time.Time `json:"createdAt"`
	// UpdatedAt stores the latest durable mutation time.
	UpdatedAt time.Time `json:"updatedAt"`
	// DeactivatedAt stores soft-deletion time when inactive.
	DeactivatedAt *time.Time `json:"deactivatedAt,omitempty"`
	// Version stores the optimistic concurrency token.
	Version int64 `json:"version"`
}

// Active reports whether the group accepts normal operations.
func (group Group) Active() bool { return group.ID > 0 && group.DeactivatedAt == nil }

// Membership stores one active player role.
type Membership struct {
	// GroupID identifies the group.
	GroupID int64 `json:"groupId"`
	// PlayerID identifies the member.
	PlayerID int64 `json:"playerId"`
	// Username stores the current visible name for lists.
	Username string `json:"username"`
	// Figure stores the current avatar figure for lists.
	Figure string `json:"figure"`
	// Role stores the durable role.
	Role Role `json:"role"`
	// JoinedAt stores membership creation time.
	JoinedAt time.Time `json:"joinedAt"`
	// UpdatedAt stores the latest role mutation time.
	UpdatedAt time.Time `json:"updatedAt"`
	// Version stores the optimistic concurrency token.
	Version int64 `json:"version"`
}

// Request stores one pending exclusive-group request.
type Request struct {
	// GroupID identifies the group.
	GroupID int64 `json:"groupId"`
	// PlayerID identifies the requester.
	PlayerID int64 `json:"playerId"`
	// Username stores the current visible name.
	Username string `json:"username"`
	// Figure stores the current avatar figure.
	Figure string `json:"figure"`
	// RequestedAt stores request creation time.
	RequestedAt time.Time `json:"requestedAt"`
}

// PlayerGroup combines membership and immutable group list data.
type PlayerGroup struct {
	// Group stores group metadata.
	Group Group `json:"group"`
	// Role stores the player's active role.
	Role Role `json:"role"`
	// Favorite reports whether this is the player's active favorite.
	Favorite bool `json:"favorite"`
}

// BadgeElement describes one editor base or symbol.
type BadgeElement struct {
	// Kind identifies base or symbol.
	Kind BadgeKind
	// ID identifies the protocol editor entry.
	ID int32
	// ValueA stores renderer asset metadata.
	ValueA string
	// ValueB stores renderer asset metadata.
	ValueB string
	// Order stores stable editor ordering.
	Order int32
}

// BadgeColor describes one editor color.
type BadgeColor struct {
	// Family identifies its editor collection.
	Family ColorFamily
	// ID identifies the protocol color.
	ID int32
	// Hex stores an uppercase six-digit RGB value.
	Hex string
	// Order stores stable editor ordering.
	Order int32
}

// BadgePart stores one normalized badge layer.
type BadgePart struct {
	// Ordinal stores zero-based layer order.
	Ordinal int16 `json:"ordinal"`
	// Kind identifies base or symbol.
	Kind BadgeKind `json:"kind"`
	// ElementID selects an enabled editor element.
	ElementID int32 `json:"elementId"`
	// ColorID selects an enabled part color.
	ColorID int32 `json:"colorId"`
	// Position stores the renderer overlay position.
	Position int32 `json:"position"`
}

// EligibleRoom describes a creator home-room choice.
type EligibleRoom struct {
	// ID identifies the room.
	ID int64
	// Name stores the visible room name.
	Name string
}

// RoomBinding pairs one room identifier with active group metadata.
type RoomBinding struct {
	// RoomID identifies the linked room.
	RoomID int64
	// Group stores the immutable active group record.
	Group Group
}
