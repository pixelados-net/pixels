package openapi

// GroupActorRequest attributes one read-only group administration request.
type GroupActorRequest struct {
	APIKeyRequest
	// ActorPlayerID identifies the permission subject.
	ActorPlayerID int64 `header:"X-Actor-Player-ID" required:"true" minimum:"1"`
}

// GroupAuditRequest stores mandatory mutation attribution.
type GroupAuditRequest struct {
	APIKeyRequest
	// ActorPlayerID identifies the administrative actor.
	ActorPlayerID int64 `json:"actorPlayerId" required:"true" minimum:"1"`
	// Reason stores durable audit context.
	Reason string `json:"reason" required:"true" minLength:"3" maxLength:"500"`
}

// GroupPathRequest identifies one social group.
type GroupPathRequest struct {
	APIKeyRequest
	// GroupID identifies the social group.
	GroupID int64 `path:"groupId" required:"true" minimum:"1"`
}

// GroupReadRequest identifies a group and read actor.
type GroupReadRequest struct {
	GroupActorRequest
	// GroupID identifies the social group.
	GroupID int64 `path:"groupId" required:"true" minimum:"1"`
}

// GroupListRequest stores bounded list filters.
type GroupListRequest struct {
	GroupActorRequest
	// Query filters case-insensitive group names.
	Query string `query:"query,omitempty" maxLength:"64"`
	// Offset stores the zero-based page offset.
	Offset int `query:"offset,omitempty" minimum:"0"`
	// Limit bounds returned groups.
	Limit int `query:"limit,omitempty" minimum:"1" maximum:"200" default:"50"`
	// OwnerPlayerID filters one owner.
	OwnerPlayerID int64 `query:"ownerPlayerId,omitempty" minimum:"1"`
	// HomeRoomID filters one headquarters.
	HomeRoomID int64 `query:"homeRoomId,omitempty" minimum:"1"`
	// State filters one admission state.
	State *int16 `query:"state,omitempty" minimum:"0" maximum:"2"`
	// ForumEnabled filters forum entitlement.
	ForumEnabled *bool `query:"forumEnabled,omitempty"`
	// Active selects active or deactivated rows.
	Active *bool `query:"active,omitempty"`
}

// GroupBadgePartRequest stores one validated badge layer.
type GroupBadgePartRequest struct {
	// Kind identifies base zero or symbol one.
	Kind int16 `json:"kind" required:"true" minimum:"0" maximum:"1"`
	// ElementID selects one enabled element.
	ElementID int32 `json:"elementId" required:"true" minimum:"1" maximum:"999"`
	// ColorID selects one enabled color.
	ColorID int32 `json:"colorId" required:"true" minimum:"1" maximum:"999"`
	// Position stores renderer position.
	Position int32 `json:"position" required:"true" minimum:"0" maximum:"9"`
}

// GroupCreateRequest stores one administrative group creation.
type GroupCreateRequest struct {
	GroupAuditRequest
	// IdempotencyKey prevents duplicate logical creation requests.
	IdempotencyKey string `header:"Idempotency-Key" required:"true"`
	// OwnerPlayerID identifies the owner.
	OwnerPlayerID int64 `json:"ownerPlayerId" required:"true" minimum:"1"`
	// Name stores the visible title.
	Name string `json:"name" required:"true" minLength:"1" maxLength:"29"`
	// Description stores visible information.
	Description string `json:"description" maxLength:"254"`
	// HomeRoomID identifies headquarters.
	HomeRoomID int64 `json:"homeRoomId" required:"true" minimum:"1"`
	// ColorA identifies the primary editor color.
	ColorA int32 `json:"colorA" required:"true" minimum:"1"`
	// ColorB identifies the secondary editor color.
	ColorB int32 `json:"colorB" required:"true" minimum:"1"`
	// BadgeParts stores one through five layers.
	BadgeParts []GroupBadgePartRequest `json:"badgeParts" required:"true" minItems:"1" maxItems:"5"`
	// Charge applies the configured price and defaults to true; false creates without charging.
	Charge *bool `json:"charge,omitempty" default:"true"`
}

// GroupVersionRequest stores a versioned group mutation.
type GroupVersionRequest struct {
	GroupPathRequest
	// ActorPlayerID identifies the administrative actor.
	ActorPlayerID int64 `json:"actorPlayerId" required:"true" minimum:"1"`
	// Reason stores durable audit context.
	Reason string `json:"reason" required:"true" minLength:"3" maxLength:"500"`
	// Version stores the optimistic token.
	Version int64 `json:"version" required:"true" minimum:"1"`
}

// GroupUpdateRequest stores optional identity and settings fields.
type GroupUpdateRequest struct {
	GroupVersionRequest
	// Name optionally replaces the title.
	Name *string `json:"name,omitempty" minLength:"1" maxLength:"29"`
	// Description optionally replaces public information.
	Description *string `json:"description,omitempty" maxLength:"254"`
	// State optionally replaces admission policy.
	State *int16 `json:"state,omitempty" minimum:"0" maximum:"2"`
	// CanMembersDecorate optionally replaces decoration policy.
	CanMembersDecorate *bool `json:"canMembersDecorate,omitempty"`
	// ColorA optionally replaces the primary color.
	ColorA *int32 `json:"colorA,omitempty" minimum:"1"`
	// ColorB optionally replaces the secondary color.
	ColorB *int32 `json:"colorB,omitempty" minimum:"1"`
}

// GroupBadgeRequest stores one optimistic badge replacement.
type GroupBadgeRequest struct {
	GroupVersionRequest
	// BadgeParts stores one through five normalized layers.
	BadgeParts []GroupBadgePartRequest `json:"badgeParts" required:"true" minItems:"1" maxItems:"5"`
}

// GroupRoomRequest stores one optimistic home-room replacement.
type GroupRoomRequest struct {
	GroupVersionRequest
	// HomeRoomID identifies the eligible owner room.
	HomeRoomID int64 `json:"homeRoomId" required:"true" minimum:"1"`
}

// GroupOwnerRequest stores one optimistic ownership transfer.
type GroupOwnerRequest struct {
	GroupVersionRequest
	// OwnerPlayerID identifies an existing member.
	OwnerPlayerID int64 `json:"ownerPlayerId" required:"true" minimum:"1"`
}

// GroupPlayerReadRequest identifies one player's group projection.
type GroupPlayerReadRequest struct {
	GroupActorRequest
	// PlayerID identifies the player.
	PlayerID int64 `path:"playerId" required:"true" minimum:"1"`
}

// GroupFavoriteRequest stores a favorite set or clear operation.
type GroupFavoriteRequest struct {
	APIKeyRequest
	// PlayerID identifies the target player.
	PlayerID int64 `path:"playerId" required:"true" minimum:"1"`
	// ActorPlayerID identifies the administrative actor in the body.
	ActorPlayerID int64 `json:"actorPlayerId" required:"true" minimum:"1"`
	// Reason stores durable audit context.
	Reason string `json:"reason" required:"true" minLength:"3" maxLength:"500"`
	// GroupID selects a membership or remains null to clear.
	FavoriteGroupID *int64 `json:"groupId,omitempty" minimum:"1"`
}

// GroupMemberListRequest stores roster filters.
type GroupMemberListRequest struct {
	GroupReadRequest
	// Page stores the zero-based Nitro page.
	Page int `query:"page,omitempty" minimum:"0"`
	// Query filters usernames.
	Query string `query:"query,omitempty" maxLength:"64"`
	// Level selects all, admins, or pending.
	Level int `query:"level,omitempty" minimum:"0" maximum:"2"`
}

// GroupMutationRequest stores one attributed group mutation.
type GroupMutationRequest struct {
	GroupPathRequest
	// ActorPlayerID identifies the actor.
	ActorPlayerID int64 `json:"actorPlayerId" required:"true" minimum:"1"`
	// Reason stores durable audit context.
	Reason string `json:"reason" required:"true" minLength:"3" maxLength:"500"`
}

// GroupMemberAddRequest stores one direct member insertion.
type GroupMemberAddRequest struct {
	GroupMutationRequest
	// PlayerID identifies the target.
	PlayerID int64 `json:"playerId" required:"true" minimum:"1"`
	// Role selects admin one or member two.
	Role int16 `json:"role" required:"true" minimum:"1" maximum:"2"`
}

// GroupMemberMutationRequest identifies one roster target.
type GroupMemberMutationRequest struct {
	GroupMutationRequest
	// PlayerID identifies the target.
	PlayerID int64 `path:"playerId" required:"true" minimum:"1"`
}

// GroupMemberRoleRequest stores one target role change.
type GroupMemberRoleRequest struct {
	GroupMemberMutationRequest
	// Role selects admin one or member two.
	Role int16 `json:"role" required:"true" minimum:"1" maximum:"2"`
}

// GroupRequestListRequest stores pending request pagination.
type GroupRequestListRequest struct {
	GroupReadRequest
	// Offset stores the zero-based offset.
	Offset int `query:"offset,omitempty" minimum:"0"`
	// Limit bounds pending requests.
	Limit int `query:"limit,omitempty" minimum:"1" maximum:"100" default:"50"`
}

// GroupMembershipResponse documents one membership projection.
type GroupMembershipResponse struct {
	GroupID int64 `json:"groupId"`
}

// GroupRequestResponse documents one pending request projection.
type GroupRequestResponse struct {
	PlayerID int64 `json:"playerId"`
}
