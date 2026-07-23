package routes

import (
	"time"

	"github.com/niflaot/pixels/internal/permission"
	permissionmodel "github.com/niflaot/pixels/internal/permission/model"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
)

// NodeResponse contains one registered permission node.
type NodeResponse struct {
	// Node identifies the registered capability.
	Node permission.Node `json:"node"`
	// PerkName stores the optional Nitro perk code.
	PerkName string `json:"perkName,omitempty"`
	// Package stores the declaring domain package.
	Package string `json:"package"`
	// Description explains runtime plugin nodes to administrators.
	Description string `json:"description,omitempty"`
}

// GroupResponse contains one permission group record.
type GroupResponse struct {
	// ID identifies the group.
	ID int64 `json:"id"`
	// Name stores the unique group name.
	Name string `json:"name"`
	// Weight stores group resolution priority.
	Weight int32 `json:"weight"`
	// Prefix stores a future chat prefix.
	Prefix string `json:"prefix"`
	// PrefixColor stores a future chat prefix color.
	PrefixColor string `json:"prefixColor"`
	// RoomEffectID stores the synthetic room effect.
	RoomEffectID *int32 `json:"roomEffectId,omitempty"`
	// ParentGroupID identifies the optional inherited group.
	ParentGroupID *int64 `json:"parentGroupId,omitempty"`
	// Version stores optimistic locking state.
	Version int64 `json:"version"`
	// UpdatedAt stores the last group mutation time.
	UpdatedAt time.Time `json:"updatedAt"`
}

// EffectiveNodeResponse contains one resolved permission decision.
type EffectiveNodeResponse struct {
	// Node identifies the queried registered capability.
	Node permission.Node `json:"node"`
	// Allowed stores the resolved decision.
	Allowed bool `json:"allowed"`
	// Source identifies the deciding player override or group.
	Source string `json:"source"`
}

// CheckResponse contains one point permission decision.
type CheckResponse struct {
	// PlayerID identifies the checked player.
	PlayerID int64 `json:"playerId"`
	// Node identifies the checked capability.
	Node permission.Node `json:"node"`
	// Allowed stores the resolved decision.
	Allowed bool `json:"allowed"`
}

// MutationResponse acknowledges one permission mutation.
type MutationResponse struct {
	// Updated reports successful persistence.
	Updated bool `json:"updated"`
}

// groupResponse maps one persistent group into HTTP output.
func groupResponse(group permissionmodel.Group) GroupResponse {
	return GroupResponse{ID: group.ID, Name: group.Name, Weight: group.Weight, Prefix: group.Prefix,
		PrefixColor: group.PrefixColor, RoomEffectID: group.RoomEffectID, ParentGroupID: group.ParentGroupID,
		Version: group.Version.Version, UpdatedAt: group.UpdatedAt}
}

// effectiveResponse maps one resolved node into HTTP output.
func effectiveResponse(resolved permissionservice.ResolvedNode) EffectiveNodeResponse {
	return EffectiveNodeResponse{Node: resolved.Node, Allowed: resolved.Allowed, Source: resolved.Source}
}
