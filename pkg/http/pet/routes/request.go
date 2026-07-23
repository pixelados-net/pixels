package routes

// AuditRequest stores required mutation attribution.
type AuditRequest struct {
	// ActorPlayerID identifies the administrative actor.
	ActorPlayerID int64 `json:"actorPlayerId"`
	// Reason stores the human-readable audit reason.
	Reason string `json:"reason"`
}

// CreateRequest stores one idempotent pet grant.
type CreateRequest struct {
	AuditRequest
	// OwnerPlayerID identifies the receiving owner.
	OwnerPlayerID int64 `json:"ownerPlayerId"`
	// Name stores the visible name.
	Name string `json:"name"`
	// TypeID identifies species.
	TypeID int32 `json:"typeId"`
	// BreedID identifies breed.
	BreedID int32 `json:"breedId"`
	// PaletteID identifies palette.
	PaletteID int32 `json:"paletteId"`
	// Color stores renderer color.
	Color string `json:"color"`
}

// UpdateRequest stores one optimistic pet patch.
type UpdateRequest struct {
	AuditRequest
	// Name optionally replaces the visible name.
	Name *string `json:"name"`
	// BreedID optionally replaces renderer breed.
	BreedID *int32 `json:"breedId"`
	// PaletteID optionally replaces renderer palette.
	PaletteID *int32 `json:"paletteId"`
	// Color optionally replaces renderer color.
	Color *string `json:"color"`
	// PublicRide optionally replaces public riding access.
	PublicRide *bool `json:"publicRide"`
	// PublicBreed optionally replaces public breeding access.
	PublicBreed *bool `json:"publicBreed"`
	// Version stores the required optimistic version.
	Version int64 `json:"version"`
}

// OwnerRequest stores one explicit inventory ownership transfer.
type OwnerRequest struct {
	AuditRequest
	// OwnerPlayerID identifies the new owner.
	OwnerPlayerID int64 `json:"ownerPlayerId"`
	// Version stores the required optimistic version.
	Version int64 `json:"version"`
}

// LocationRequest stores one administrative place or pickup request.
type LocationRequest struct {
	AuditRequest
	// RoomID identifies placement or remains nil for pickup.
	RoomID *int64 `json:"roomId"`
	// X stores the optional placement tile.
	X *int `json:"x"`
	// Y stores the optional placement tile.
	Y *int `json:"y"`
}

// StatsRequest stores bounded stat deltas.
type StatsRequest struct {
	AuditRequest
	// EnergyDelta changes materialized energy.
	EnergyDelta int32 `json:"energyDelta"`
	// HappinessDelta changes materialized happiness.
	HappinessDelta int32 `json:"happinessDelta"`
	// ExperienceDelta changes accumulated experience.
	ExperienceDelta int32 `json:"experienceDelta"`
	// Version stores the required optimistic version.
	Version int64 `json:"version"`
}

// DeleteRequest stores one optimistic soft deletion.
type DeleteRequest struct {
	AuditRequest
	// Version stores the required optimistic version.
	Version int64 `json:"version"`
}
