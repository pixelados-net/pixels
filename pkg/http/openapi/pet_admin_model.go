package openapi

import "time"

// PetActorRequest authenticates one read-only administrative actor.
type PetActorRequest struct {
	APIKeyRequest
	// ActorPlayerID identifies the actor whose permission node is resolved.
	ActorPlayerID int64 `header:"X-Actor-Player-ID" required:"true" minimum:"1"`
}

// PetAuditRequest stores protected mutation attribution.
type PetAuditRequest struct {
	APIKeyRequest
	// ActorPlayerID identifies the administrative actor.
	ActorPlayerID int64 `json:"actorPlayerId" required:"true" minimum:"1"`
	// Reason stores the human-readable audit reason.
	Reason string `json:"reason" required:"true" minLength:"1"`
}

// PetIDRequest identifies one pet.
type PetIDRequest struct {
	APIKeyRequest
	// ID identifies the pet.
	ID int64 `path:"id" required:"true" minimum:"1"`
}

// PetReadRequest identifies one pet and its read-only administrative actor.
type PetReadRequest struct {
	PetActorRequest
	// ID identifies the pet.
	ID int64 `path:"id" required:"true" minimum:"1"`
}

// PetListRequest stores protected pet-list filters.
type PetListRequest struct {
	PetActorRequest
	// OwnerPlayerID optionally restricts one owner.
	OwnerPlayerID *int64 `query:"ownerPlayerId,omitempty" minimum:"1"`
	// Name optionally matches a case-insensitive fragment.
	Name string `query:"name,omitempty"`
	// TypeID optionally restricts species.
	TypeID *int32 `query:"typeId,omitempty" minimum:"0" maximum:"35"`
	// RoomID optionally restricts placement.
	RoomID *int64 `query:"roomId,omitempty" minimum:"1"`
	// State optionally restricts lifecycle state.
	State string `query:"state,omitempty" enum:"inventory,room,breeding_reserved,harvested,composted"`
	// Deleted includes soft-deleted rows.
	Deleted bool `query:"deleted,omitempty"`
	// Cursor stores the exclusive keyset cursor.
	Cursor int64 `query:"cursor,omitempty" minimum:"0"`
	// Limit bounds returned rows.
	Limit int `query:"limit,omitempty" minimum:"1" maximum:"200" default:"50"`
}

// PetCreateRequest stores one idempotent pet grant.
type PetCreateRequest struct {
	PetAuditRequest
	// IdempotencyKey prevents duplicate grants.
	IdempotencyKey string `header:"Idempotency-Key" required:"true"`
	// OwnerPlayerID identifies the receiving owner.
	OwnerPlayerID int64 `json:"ownerPlayerId" required:"true" minimum:"1"`
	// Name stores the visible name.
	Name string `json:"name" required:"true" minLength:"2" maxLength:"16"`
	// TypeID identifies species.
	TypeID int32 `json:"typeId" required:"true" minimum:"0" maximum:"35"`
	// BreedID identifies breed.
	BreedID int32 `json:"breedId" required:"true" minimum:"0"`
	// PaletteID identifies palette.
	PaletteID int32 `json:"paletteId" required:"true" minimum:"0"`
	// Color stores renderer color.
	Color string `json:"color" required:"true" pattern:"^[0-9A-Fa-f]{6}$"`
}

// PetUpdateRequest stores one optimistic pet patch.
type PetUpdateRequest struct {
	PetIDRequest
	// ActorPlayerID identifies the administrative actor.
	ActorPlayerID int64 `json:"actorPlayerId" required:"true" minimum:"1"`
	// Reason stores the audit reason.
	Reason string `json:"reason" required:"true" minLength:"1"`
	// Name optionally replaces the visible name.
	Name *string `json:"name,omitempty" minLength:"2" maxLength:"16"`
	// BreedID optionally replaces renderer breed.
	BreedID *int32 `json:"breedId,omitempty" minimum:"0"`
	// PaletteID optionally replaces renderer palette.
	PaletteID *int32 `json:"paletteId,omitempty" minimum:"0"`
	// Color optionally replaces renderer color.
	Color *string `json:"color,omitempty" pattern:"^[0-9A-Fa-f]{6}$"`
	// PublicRide optionally replaces public riding access.
	PublicRide *bool `json:"publicRide,omitempty"`
	// PublicBreed optionally replaces public breeding access.
	PublicBreed *bool `json:"publicBreed,omitempty"`
	// Version stores the required optimistic version.
	Version int64 `json:"version" required:"true" minimum:"1"`
}

// PetDeleteRequest stores one optimistic deletion.
type PetDeleteRequest struct {
	PetIDRequest
	// ActorPlayerID identifies the administrative actor.
	ActorPlayerID int64 `json:"actorPlayerId" required:"true" minimum:"1"`
	// Reason stores the audit reason.
	Reason string `json:"reason" required:"true" minLength:"1"`
	// Version stores the required optimistic version.
	Version int64 `json:"version" required:"true" minimum:"1"`
}

// PetOwnerRequest stores one ownership transfer.
type PetOwnerRequest struct {
	PetDeleteRequest
	// OwnerPlayerID identifies the new owner.
	OwnerPlayerID int64 `json:"ownerPlayerId" required:"true" minimum:"1"`
}

// PetLocationRequest stores one validated place or pickup operation.
type PetLocationRequest struct {
	PetIDRequest
	// ActorPlayerID identifies the administrative actor.
	ActorPlayerID int64 `json:"actorPlayerId" required:"true" minimum:"1"`
	// Reason stores the audit reason.
	Reason string `json:"reason" required:"true" minLength:"1"`
	// RoomID identifies placement or remains null for pickup.
	RoomID *int64 `json:"roomId,omitempty" minimum:"1"`
	// X stores placement x when roomId is present.
	X *int `json:"x,omitempty" minimum:"0"`
	// Y stores placement y when roomId is present.
	Y *int `json:"y,omitempty" minimum:"0"`
}

// PetStatsRequest stores bounded stat deltas.
type PetStatsRequest struct {
	PetDeleteRequest
	// EnergyDelta changes materialized energy.
	EnergyDelta int32 `json:"energyDelta" required:"true"`
	// HappinessDelta changes materialized happiness.
	HappinessDelta int32 `json:"happinessDelta" required:"true"`
	// ExperienceDelta changes accumulated experience.
	ExperienceDelta int32 `json:"experienceDelta" required:"true"`
}

// PetBreedListRequest stores optional breed filters.
type PetBreedListRequest struct {
	PetActorRequest
	// TypeID optionally restricts species.
	TypeID *int32 `query:"typeId,omitempty" minimum:"0" maximum:"35"`
	// Sellable includes only catalog breeds.
	Sellable bool `query:"sellable,omitempty"`
}

// PetAdminResponse contains protected pet support state.
type PetAdminResponse struct {
	// ID identifies the pet.
	ID int64 `json:"id" required:"true"`
	// OwnerPlayerID identifies the owner.
	OwnerPlayerID int64 `json:"ownerPlayerId" required:"true"`
	// OwnerName stores the owner name.
	OwnerName string `json:"ownerName" required:"true"`
	// Name stores the visible name.
	Name string `json:"name" required:"true"`
	// TypeID identifies species.
	TypeID int32 `json:"typeId" required:"true"`
	// BreedID identifies breed.
	BreedID int32 `json:"breedId" required:"true"`
	// PaletteID identifies palette.
	PaletteID int32 `json:"paletteId" required:"true"`
	// Color stores renderer color.
	Color string `json:"color" required:"true"`
	// Figure stores the renderer figure.
	Figure string `json:"figure" required:"true"`
	// Level stores current level.
	Level int32 `json:"level" required:"true"`
	// Experience stores accumulated experience.
	Experience int32 `json:"experience" required:"true"`
	// Energy stores materialized energy.
	Energy int32 `json:"energy" required:"true"`
	// Happiness stores materialized happiness.
	Happiness int32 `json:"happiness" required:"true"`
	// Respect stores accumulated respects.
	Respect int32 `json:"respect" required:"true"`
	// RoomID identifies current placement.
	RoomID *int64 `json:"roomId,omitempty"`
	// State stores durable lifecycle state.
	State string `json:"state" required:"true"`
	// HasSaddle reports equipment state.
	HasSaddle bool `json:"hasSaddle" required:"true"`
	// PublicRide reports public riding access.
	PublicRide bool `json:"publicRide" required:"true"`
	// PublicBreed reports public breeding access.
	PublicBreed bool `json:"publicBreed" required:"true"`
	// GrowAt stores plant maturity.
	GrowAt *time.Time `json:"growAt,omitempty"`
	// DieAt stores plant death.
	DieAt *time.Time `json:"dieAt,omitempty"`
	// Version stores optimistic state.
	Version int64 `json:"version" required:"true"`
}

// PetListResponse stores one protected keyset page.
type PetListResponse struct {
	// Items stores ordered pet records.
	Items []PetAdminResponse `json:"items" required:"true"`
	// NextCursor stores the next identifier or zero.
	NextCursor int64 `json:"nextCursor" required:"true"`
}

// PetMutationResponse stores one successful mutation.
type PetMutationResponse struct {
	// Pet stores the resulting aggregate.
	Pet *PetAdminResponse `json:"pet,omitempty"`
	// Created reports whether a grant created a row.
	Created bool `json:"created,omitempty"`
}

// PetReferenceRefreshResponse stores publication status.
type PetReferenceRefreshResponse struct {
	// Refreshed reports successful publication.
	Refreshed bool `json:"refreshed" required:"true"`
}
