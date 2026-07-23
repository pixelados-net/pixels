package routes

import (
	"time"

	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
)

// AuditRequest stores required mutation attribution.
type AuditRequest struct {
	// ActorPlayerID identifies the administrative actor.
	ActorPlayerID int64 `json:"actorPlayerId"`
	// Reason stores the mandatory human-readable reason.
	Reason string `json:"reason"`
}

// BadgePartRequest stores one typed badge layer.
type BadgePartRequest struct {
	// Kind identifies base or symbol.
	Kind grouprecord.BadgeKind `json:"kind"`
	// ElementID selects an enabled registry element.
	ElementID int32 `json:"elementId"`
	// ColorID selects an enabled family color.
	ColorID int32 `json:"colorId"`
	// Position stores renderer placement.
	Position int32 `json:"position"`
}

// GroupResponse contains protected group state.
type GroupResponse struct {
	// Group stores durable metadata.
	Group grouprecord.Group `json:"group"`
	// BadgeParts stores normalized badge layers.
	BadgeParts []grouprecord.BadgePart `json:"badgeParts"`
}

// GroupListResponse contains one bounded group page.
type GroupListResponse struct {
	// Items stores ordered groups.
	Items []grouprecord.Group `json:"items"`
	// NextOffset stores the next offset or zero at the end.
	NextOffset int `json:"nextOffset"`
}

// CreateRequest stores one administrative group creation.
type CreateRequest struct {
	AuditRequest
	// OwnerPlayerID identifies the owner.
	OwnerPlayerID int64 `json:"ownerPlayerId"`
	// Name stores the group title.
	Name string `json:"name"`
	// Description stores public information.
	Description string `json:"description"`
	// HomeRoomID identifies headquarters.
	HomeRoomID int64 `json:"homeRoomId"`
	// ColorA identifies the primary color.
	ColorA int32 `json:"colorA"`
	// ColorB identifies the secondary color.
	ColorB int32 `json:"colorB"`
	// BadgeParts stores requested layers.
	BadgeParts []BadgePartRequest `json:"badgeParts"`
	// Charge optionally applies the configured creation price and defaults to true.
	Charge *bool `json:"charge"`
}

// UpdateRequest stores one optimistic group patch.
type UpdateRequest struct {
	AuditRequest
	// Version stores the required optimistic version.
	Version int64 `json:"version"`
	// Name optionally replaces the title.
	Name *string `json:"name"`
	// Description optionally replaces public information.
	Description *string `json:"description"`
	// State optionally replaces admission policy.
	State *grouprecord.State `json:"state"`
	// CanMembersDecorate optionally replaces decoration policy.
	CanMembersDecorate *bool `json:"canMembersDecorate"`
	// ColorA optionally replaces the primary color.
	ColorA *int32 `json:"colorA"`
	// ColorB optionally replaces the secondary color.
	ColorB *int32 `json:"colorB"`
}

// VersionRequest stores mutation attribution and optimistic version.
type VersionRequest struct {
	AuditRequest
	// Version stores the required optimistic version.
	Version int64 `json:"version"`
}

// MemberPageResponse stores roster pagination metadata.
type MemberPageResponse struct {
	// Page stores the domain roster result.
	Page grouprecord.MemberPage `json:"page"`
	// CanManage reports whether the actor may mutate it.
	CanManage bool `json:"canManage"`
}

// MutationResponse describes bounded mutation output.
type MutationResponse struct {
	// Group stores resulting group state when applicable.
	Group *grouprecord.Group `json:"group,omitempty"`
	// Membership stores resulting membership when applicable.
	Membership *grouprecord.Membership `json:"membership,omitempty"`
	// Count stores affected or returned rows.
	Count int `json:"count,omitempty"`
	// Created reports an idempotent insertion.
	Created bool `json:"created,omitempty"`
}

// ForumThreadResponse stores one thread with a bounded message page.
type ForumThreadResponse struct {
	// Thread stores retained thread metadata.
	Thread grouprecord.Thread `json:"thread"`
	// Posts stores ordered messages.
	Posts []grouprecord.Post `json:"posts"`
	// Total stores total visible messages.
	Total int32 `json:"total"`
}

// TimeResponse stores an operation timestamp.
type TimeResponse struct {
	// CompletedAt stores completion time.
	CompletedAt time.Time `json:"completedAt"`
}
