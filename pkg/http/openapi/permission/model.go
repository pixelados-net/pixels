// Package permission contains reflected permission administration API models.
package permission

// APIKeyRequest contains the API key header.
type APIKeyRequest struct {
	// APIKey stores the configured access key.
	APIKey string `header:"X-API-Key" required:"true" description:"Access key configured by PIXELS_ACCESS_KEY."`
}

// NodeResponse contains one process-registered permission node.
type NodeResponse struct {
	// Node identifies the registered capability.
	Node string `json:"node" required:"true" example:"catalog.admin.manage"`
	// PerkName stores the optional Nitro perk code.
	PerkName string `json:"perkName,omitempty" example:"HEIGHTMAP_EDITOR_BETA"`
	// Package stores the declaring domain package.
	Package string `json:"package" required:"true"`
}

// NodesResponse contains registered permission nodes.
type NodesResponse []NodeResponse

// GroupCreateRequest contains permission group creation fields.
type GroupCreateRequest struct {
	APIKeyRequest
	// Name stores the unique group name.
	Name string `json:"name" required:"true" example:"moderator"`
	// Weight stores group resolution priority.
	Weight int32 `json:"weight" example:"50"`
	// Prefix stores the future localized chat prefix.
	Prefix string `json:"prefix,omitempty"`
	// PrefixColor stores the future chat prefix color.
	PrefixColor string `json:"prefixColor,omitempty" example:"#ff0000"`
	// RoomEffectID stores the synthetic room effect.
	RoomEffectID *int32 `json:"roomEffectId,omitempty" minimum:"1"`
	// ParentGroupID identifies the optional inherited group.
	ParentGroupID *int64 `json:"parentGroupId,omitempty" minimum:"1"`
}

// GroupPatchRequest contains optional permission group changes.
type GroupPatchRequest struct {
	APIKeyRequest
	// ID identifies the permission group.
	ID int64 `path:"id" required:"true" minimum:"1"`
	// Name replaces the group name.
	Name *string `json:"name,omitempty"`
	// Weight replaces group resolution priority.
	Weight *int32 `json:"weight,omitempty"`
	// Prefix replaces the future chat prefix.
	Prefix *string `json:"prefix,omitempty"`
	// PrefixColor replaces the future chat prefix color.
	PrefixColor *string `json:"prefixColor,omitempty"`
	// RoomEffectID replaces the synthetic room effect.
	RoomEffectID *int32 `json:"roomEffectId,omitempty" minimum:"1"`
	// ClearRoomEffect removes the synthetic room effect.
	ClearRoomEffect bool `json:"clearRoomEffect,omitempty"`
	// ParentGroupID replaces the inherited group.
	ParentGroupID *int64 `json:"parentGroupId,omitempty" minimum:"1"`
	// ClearParent removes the inherited group.
	ClearParent bool `json:"clearParent,omitempty"`
}

// GroupNodeRequest grants or denies one node to a group.
type GroupNodeRequest struct {
	APIKeyRequest
	// ID identifies the permission group.
	ID int64 `path:"id" required:"true" minimum:"1"`
	// Node identifies a registered capability or wildcard.
	Node string `json:"node" required:"true" example:"room.moderation.*"`
	// Allowed reports whether the grant allows or denies the node.
	Allowed bool `json:"allowed" required:"true"`
}

// GroupNodeDeleteRequest identifies one group grant.
type GroupNodeDeleteRequest struct {
	APIKeyRequest
	// ID identifies the permission group.
	ID int64 `path:"id" required:"true" minimum:"1"`
	// Node identifies the URL-escaped capability or wildcard.
	Node string `path:"node" required:"true"`
}

// MembershipRequest identifies one player group membership.
type MembershipRequest struct {
	APIKeyRequest
	// PlayerID identifies the player.
	PlayerID int64 `path:"playerId" required:"true" minimum:"1"`
	// GroupID identifies the permission group.
	GroupID int64 `path:"groupId" required:"true" minimum:"1"`
}

// PlayerNodeRequest grants or denies one direct player node.
type PlayerNodeRequest struct {
	APIKeyRequest
	// PlayerID identifies the player.
	PlayerID int64 `path:"playerId" required:"true" minimum:"1"`
	// Node identifies a registered capability or wildcard.
	Node string `json:"node" required:"true" example:"catalog.admin.manage"`
	// Allowed reports whether the grant allows or denies the node.
	Allowed bool `json:"allowed" required:"true"`
}

// PlayerNodeDeleteRequest identifies one direct player grant.
type PlayerNodeDeleteRequest struct {
	APIKeyRequest
	// PlayerID identifies the player.
	PlayerID int64 `path:"playerId" required:"true" minimum:"1"`
	// Node identifies the URL-escaped capability or wildcard.
	Node string `path:"node" required:"true"`
}

// PlayerRequest identifies one player permission projection.
type PlayerRequest struct {
	APIKeyRequest
	// PlayerID identifies the player.
	PlayerID int64 `path:"playerId" required:"true" minimum:"1"`
}

// CheckRequest identifies one point permission query.
type CheckRequest struct {
	PlayerRequest
	// Node identifies the queried concrete permission node.
	Node string `query:"node" required:"true" example:"catalog.admin.manage"`
}

// GroupResponse contains one permission group record.
type GroupResponse struct {
	// ID identifies the group.
	ID int64 `json:"id" required:"true"`
	// Name stores the unique group name.
	Name string `json:"name" required:"true"`
	// Weight stores group resolution priority.
	Weight int32 `json:"weight" required:"true"`
	// Prefix stores the future chat prefix.
	Prefix string `json:"prefix" required:"true"`
	// PrefixColor stores the future prefix color.
	PrefixColor string `json:"prefixColor" required:"true"`
	// RoomEffectID stores the synthetic room effect.
	RoomEffectID *int32 `json:"roomEffectId,omitempty"`
	// ParentGroupID identifies the optional inherited group.
	ParentGroupID *int64 `json:"parentGroupId,omitempty"`
	// Version stores optimistic locking state.
	Version int64 `json:"version" required:"true"`
	// UpdatedAt stores the last mutation time.
	UpdatedAt string `json:"updatedAt" required:"true" format:"date-time"`
}

// GroupsResponse contains permission groups.
type GroupsResponse []GroupResponse

// EffectiveNodeResponse contains one resolved permission decision.
type EffectiveNodeResponse struct {
	// Node identifies the registered capability.
	Node string `json:"node" required:"true"`
	// Allowed stores the resolved decision.
	Allowed bool `json:"allowed" required:"true"`
	// Source identifies the deciding player override or group.
	Source string `json:"source" required:"true"`
}

// EffectiveResponse contains effective permission decisions.
type EffectiveResponse []EffectiveNodeResponse

// CheckResponse contains one point permission decision.
type CheckResponse struct {
	// PlayerID identifies the checked player.
	PlayerID int64 `json:"playerId" required:"true"`
	// Node identifies the checked capability.
	Node string `json:"node" required:"true"`
	// Allowed stores the resolved decision.
	Allowed bool `json:"allowed" required:"true"`
}

// MutationResponse acknowledges one permission mutation.
type MutationResponse struct {
	// Updated reports successful persistence.
	Updated bool `json:"updated" required:"true"`
}
