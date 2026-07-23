package record

import (
	"context"
	"time"
)

// GroupFilter stores bounded administration list filters.
type GroupFilter struct {
	// Query filters names case-insensitively.
	Query string
	// OwnerPlayerID filters one owner when positive.
	OwnerPlayerID int64
	// HomeRoomID filters one headquarters when positive.
	HomeRoomID int64
	// State filters one admission state when supplied.
	State *State
	// ForumEnabled filters forum entitlement when supplied.
	ForumEnabled *bool
	// Active filters active or deactivated rows when supplied.
	Active *bool
	// Offset stores the zero-based row offset.
	Offset int
	// Limit caps returned rows.
	Limit int
}

// CreateParams contains one validated atomic group insertion.
type CreateParams struct {
	// OwnerPlayerID identifies the owner.
	OwnerPlayerID int64
	// Name stores the visible title.
	Name string
	// Description stores filtered plain text.
	Description string
	// HomeRoomID identifies the locked eligible room.
	HomeRoomID int64
	// State stores membership admission policy.
	State State
	// ColorA identifies the primary color.
	ColorA int32
	// ColorB identifies the secondary color.
	ColorB int32
	// BadgeCode stores the compiled badge.
	BadgeCode string
	// BadgeParts stores normalized badge layers.
	BadgeParts []BadgePart
}

// GroupPatch contains optional identity and settings mutations.
type GroupPatch struct {
	// Name replaces the title when supplied.
	Name *string
	// Description replaces public information when supplied.
	Description *string
	// State replaces admission policy when supplied.
	State *State
	// CanMembersDecorate replaces group room decoration policy when supplied.
	CanMembersDecorate *bool
	// ColorA replaces the primary editor color when supplied.
	ColorA *int32
	// ColorB replaces the secondary editor color when supplied.
	ColorB *int32
	// ForumEnabled replaces forum entitlement when supplied.
	ForumEnabled *bool
	// ReadPolicy replaces forum read policy when supplied.
	ReadPolicy *Policy
	// PostMessagePolicy replaces reply policy when supplied.
	PostMessagePolicy *Policy
	// PostThreadPolicy replaces thread-creation policy when supplied.
	PostThreadPolicy *Policy
	// ModeratePolicy replaces forum moderation policy when supplied.
	ModeratePolicy *Policy
}

// MemberPage contains one bounded roster page.
type MemberPage struct {
	// Group stores list metadata.
	Group Group `json:"group"`
	// Members stores active members and projected pending requests.
	Members []Membership `json:"members"`
	// Total stores filtered result count.
	Total int32 `json:"total"`
	// Page stores the zero-based requested page.
	Page int32 `json:"page"`
	// PageSize stores the applied page size.
	PageSize int32 `json:"pageSize"`
	// Level stores Nitro's filter value.
	Level int32 `json:"level"`
	// Query stores normalized search input.
	Query string `json:"query"`
}

// Store persists social groups, memberships, badges, and forums.
type Store interface {
	// WithinTransaction runs work in one shared PostgreSQL transaction scope.
	WithinTransaction(context.Context, func(context.Context) error) error
	// EligibleRooms lists active unbound non-template rooms owned by a player.
	EligibleRooms(context.Context, int64) ([]EligibleRoom, error)
	// LockEligibleRoom locks and validates one unbound owned room.
	LockEligibleRoom(context.Context, int64, int64) error
	// CountOwned counts active groups owned by one player.
	CountOwned(context.Context, int64) (int, error)
	// CountMemberships counts active memberships held by one player.
	CountMemberships(context.Context, int64) (int, error)
	// ClaimCreateOperation claims or replays one administrative creation key.
	ClaimCreateOperation(context.Context, string, string) (int64, bool, error)
	// CompleteCreateOperation binds a claimed creation key to its committed group.
	CompleteCreateOperation(context.Context, string, int64) error
	// InsertGroup creates group, owner membership, parts, and binding.
	InsertGroup(context.Context, CreateParams) (Group, error)
	// Group returns one group and resolved display metadata.
	Group(context.Context, int64, bool) (Group, bool, error)
	// Groups lists groups for administration.
	Groups(context.Context, GroupFilter) ([]Group, error)
	// PopularGroups lists active groups by descending member count.
	PopularGroups(context.Context, int) ([]Group, error)
	// UpdateGroup applies an optimistic metadata patch.
	UpdateGroup(context.Context, int64, int64, GroupPatch) (Group, error)
	// ReplaceBadge atomically replaces normalized parts and compiled code.
	ReplaceBadge(context.Context, int64, int64, string, []BadgePart) (Group, error)
	// BadgeParts returns stored normalized group badge layers.
	BadgeParts(context.Context, int64) ([]BadgePart, error)
	// BadgeRegistry returns every enabled editor element and color.
	BadgeRegistry(context.Context) ([]BadgeElement, []BadgeColor, error)
	// DeactivateGroup soft-deactivates group state and dependent preferences.
	DeactivateGroup(context.Context, int64, int64) (Group, error)
	// RestoreGroup validates and restores a retained group.
	RestoreGroup(context.Context, int64, int64, int64) (Group, error)
	// TransferOwner atomically exchanges owner and target roles.
	TransferOwner(context.Context, int64, int64, int64) (Group, error)
	// RebindRoom atomically replaces the home-room binding.
	RebindRoom(context.Context, int64, int64, int64) (Group, error)
	// RoomGroup returns one active room binding and immutable group metadata.
	RoomGroup(context.Context, int64) (Group, bool, error)
	// RoomGroups returns active group bindings for one bounded room batch.
	RoomGroups(context.Context, []int64) ([]RoomBinding, error)
	// RoomFurnitureLinks returns linked furniture for one room generation.
	RoomFurnitureLinks(context.Context, int64) ([]GroupFurnitureLink, error)
	// PlayerInventoryFurnitureLinks returns active linked furniture held in one player's inventory.
	PlayerInventoryFurnitureLinks(context.Context, int64) ([]GroupFurnitureLink, error)
	// LinkFurniture links granted inventory instances to one active group.
	LinkFurniture(context.Context, int64, []int64) error
	// EnableForum activates one group forum entitlement idempotently.
	EnableForum(context.Context, int64) (bool, error)
	// Membership returns one active role.
	Membership(context.Context, int64, int64) (Membership, bool, error)
	// Pending reports whether one membership request exists.
	Pending(context.Context, int64, int64) (bool, error)
	// PlayerGroups lists active memberships with favorite state.
	PlayerGroups(context.Context, int64) ([]PlayerGroup, error)
	// MemberPage returns a bounded filtered roster page.
	MemberPage(context.Context, int64, int32, int32, string, int32) (MemberPage, error)
	// Join inserts a member or pending request and reports pending and changed state.
	Join(context.Context, int64, int64, int, int, int) (Membership, bool, bool, error)
	// AddMember administratively inserts or replaces one non-owner role.
	AddMember(context.Context, int64, int64, Role, int, int) (Membership, bool, error)
	// AcceptRequest promotes one locked request to membership.
	AcceptRequest(context.Context, int64, int64, int) (Membership, error)
	// DeclineRequest removes one request idempotently.
	DeclineRequest(context.Context, int64, int64) (bool, error)
	// ApproveAll accepts a bounded ordered request batch.
	ApproveAll(context.Context, int64, int, int) ([]Membership, error)
	// ChangeRole changes one target role while preserving owner invariants.
	ChangeRole(context.Context, int64, int64, Role) (Membership, error)
	// RemoveMember removes membership, favorite, and returns HQ furniture atomically.
	RemoveMember(context.Context, int64, int64, int) (FurnitureReturn, error)
	// FurnitureCount counts target-owned furniture in the group headquarters.
	FurnitureCount(context.Context, int64, int64) (int, error)
	// SetFavorite validates membership and replaces one player preference.
	SetFavorite(context.Context, int64, *int64) error
	// Requests lists bounded pending requests.
	Requests(context.Context, int64, int, int) ([]Request, error)
	// ForumSummaries lists visible forum summaries by protocol mode.
	ForumSummaries(context.Context, int64, int32, int, int, bool, time.Time) ([]ForumSummary, int32, error)
	// ForumSummary returns one viewer-specific forum summary.
	ForumSummary(context.Context, int64, int64) (ForumSummary, bool, error)
	// Threads lists one bounded forum thread page.
	Threads(context.Context, int64, int64, int, int, bool) ([]Thread, int32, error)
	// Thread returns one forum thread.
	Thread(context.Context, int64, int64, bool) (Thread, bool, error)
	// Posts lists one bounded thread message page.
	Posts(context.Context, int64, int64, int64, int, int, bool) ([]Post, int32, error)
	// Post returns one retained post for report context.
	Post(context.Context, int64, int64) (Post, bool, error)
	// CreateThread atomically inserts a thread and first post.
	CreateThread(context.Context, int64, int64, string, string, string, string) (Thread, Post, error)
	// CreatePost atomically inserts one reply and advances counters.
	CreatePost(context.Context, int64, int64, int64, string, string, string) (Post, error)
	// UpdateThread changes pin, lock, or moderation state optimistically.
	UpdateThread(context.Context, int64, int64, int64, *bool, *bool, *ThreadState, int64, string) (Thread, error)
	// UpdatePost changes retained post moderation state optimistically.
	UpdatePost(context.Context, int64, int64, int64, PostState, int64, string) (Post, error)
	// UpdateReadMarker advances one marker monotonically.
	UpdateReadMarker(context.Context, ReadMarker) (ReadMarker, error)
	// UnreadCount returns authorized hotel-wide forum unread total.
	UnreadCount(context.Context, int64, bool) (int32, error)
}
