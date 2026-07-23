package routes

import (
	"time"

	petobservability "github.com/niflaot/pixels/internal/realm/pet/observability"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	petruntime "github.com/niflaot/pixels/internal/realm/pet/runtime"
	petdata "github.com/niflaot/pixels/networking/pet/data"
)

// MetricsResponse aliases the low-cardinality pet telemetry response.
type MetricsResponse = petobservability.Snapshot

// PetResponse contains one complete protected pet record.
type PetResponse struct {
	// ID identifies the pet.
	ID int64 `json:"id"`
	// OwnerPlayerID identifies the owner.
	OwnerPlayerID int64 `json:"ownerPlayerId"`
	// OwnerName stores the owner display name.
	OwnerName string `json:"ownerName"`
	// Name stores the visible pet name.
	Name string `json:"name"`
	// TypeID identifies species.
	TypeID int32 `json:"typeId"`
	// BreedID identifies breed.
	BreedID int32 `json:"breedId"`
	// PaletteID identifies palette.
	PaletteID int32 `json:"paletteId"`
	// Color stores renderer color.
	Color string `json:"color"`
	// Figure stores the complete renderer figure string.
	Figure string `json:"figure"`
	// Rarity stores breed rarity.
	Rarity int32 `json:"rarity"`
	// Level stores current level.
	Level int32 `json:"level"`
	// Experience stores accumulated experience.
	Experience int32 `json:"experience"`
	// Energy stores materialized energy.
	Energy int32 `json:"energy"`
	// Happiness stores materialized happiness.
	Happiness int32 `json:"happiness"`
	// Respect stores accumulated respects.
	Respect int32 `json:"respect"`
	// RoomID identifies placement.
	RoomID *int64 `json:"roomId"`
	// X stores placed x.
	X *int `json:"x"`
	// Y stores placed y.
	Y *int `json:"y"`
	// Z stores placed height.
	Z *float64 `json:"z"`
	// Rotation stores placed rotation.
	Rotation *int16 `json:"rotation"`
	// State stores durable lifecycle state.
	State string `json:"state"`
	// HasSaddle reports equipment state.
	HasSaddle bool `json:"hasSaddle"`
	// PublicRide reports public riding access.
	PublicRide bool `json:"publicRide"`
	// PublicBreed reports public breeding access.
	PublicBreed bool `json:"publicBreed"`
	// GrowAt stores monsterplant maturity.
	GrowAt *time.Time `json:"growAt"`
	// DieAt stores monsterplant death.
	DieAt *time.Time `json:"dieAt"`
	// CreatedAt stores creation time.
	CreatedAt time.Time `json:"createdAt"`
	// UpdatedAt stores mutation time.
	UpdatedAt time.Time `json:"updatedAt"`
	// DeletedAt stores optional soft deletion time.
	DeletedAt *time.Time `json:"deletedAt"`
	// Version stores optimistic state.
	Version int64 `json:"version"`
}

// PetListResponse stores one keyset page.
type PetListResponse struct {
	// Items stores ordered records.
	Items []PetResponse `json:"items"`
	// NextCursor stores the next identifier or zero at the end.
	NextCursor int64 `json:"nextCursor"`
}

// SpeciesResponse describes one renderer species slot.
type SpeciesResponse struct {
	// TypeID identifies the renderer slot.
	TypeID int32 `json:"typeId"`
	// Slug stores the stable key.
	Slug string `json:"slug"`
	// DisplayKey stores the i18n key.
	DisplayKey string `json:"displayKey"`
	// BehaviorKind selects runtime behavior.
	BehaviorKind string `json:"behaviorKind"`
	// MaximumLevel stores the level cap.
	MaximumLevel int32 `json:"maximumLevel"`
	// Rideable reports riding support.
	Rideable bool `json:"rideable"`
	// Breedable reports breeding support.
	Breedable bool `json:"breedable"`
	// Plant reports monsterplant lifecycle support.
	Plant bool `json:"plant"`
	// Enabled reports creation support.
	Enabled bool `json:"enabled"`
}

// BreedResponse describes one renderer appearance option.
type BreedResponse struct {
	// TypeID identifies species.
	TypeID int32 `json:"typeId"`
	// BreedID identifies breed.
	BreedID int32 `json:"breedId"`
	// PaletteID identifies palette.
	PaletteID int32 `json:"paletteId"`
	// Color stores default color.
	Color string `json:"color"`
	// Sellable reports catalog availability.
	Sellable bool `json:"sellable"`
	// Rarity stores rarity category.
	Rarity int32 `json:"rarity"`
	// Enabled reports runtime availability.
	Enabled bool `json:"enabled"`
}

// CommandResponse describes one trainable command.
type CommandResponse struct {
	// ID identifies the protocol command.
	ID int32 `json:"id"`
	// NameKey stores the localized name key.
	NameKey string `json:"nameKey"`
	// RequiredLevel stores the unlock level.
	RequiredLevel int32 `json:"requiredLevel"`
	// Family stores behavior grouping.
	Family string `json:"family"`
	// EnergyCost stores energy cost.
	EnergyCost int32 `json:"energyCost"`
	// HappinessCost stores happiness cost.
	HappinessCost int32 `json:"happinessCost"`
	// ExperienceReward stores experience reward.
	ExperienceReward int32 `json:"experienceReward"`
	// DurationMilliseconds stores action duration.
	DurationMilliseconds int64 `json:"durationMilliseconds"`
	// CooldownMilliseconds stores reuse cooldown.
	CooldownMilliseconds int64 `json:"cooldownMilliseconds"`
	// Enabled reports runtime availability.
	Enabled bool `json:"enabled"`
}

// CommandsResponse stores commands and per-species mappings.
type CommandsResponse struct {
	// Items stores ordered commands.
	Items []CommandResponse `json:"items"`
	// Species stores command identifiers by species.
	Species map[int32][]int32 `json:"species"`
}

// MutationResponse stores one successful administrative mutation.
type MutationResponse struct {
	// Pet stores the resulting pet when applicable.
	Pet *PetResponse `json:"pet,omitempty"`
	// Created reports whether an idempotent grant created a row.
	Created bool `json:"created,omitempty"`
}

// ReferenceRefreshResponse stores publication status.
type ReferenceRefreshResponse struct {
	// Refreshed reports successful atomic publication.
	Refreshed bool `json:"refreshed"`
}

// petResponse maps one durable aggregate.
func petResponse(pet petrecord.Pet) PetResponse {
	figure := petdata.FigureString(petruntime.InventoryPet(pet).Figure)
	return PetResponse{ID: pet.ID, OwnerPlayerID: pet.OwnerPlayerID, OwnerName: pet.OwnerName, Name: pet.Name, TypeID: pet.TypeID,
		BreedID: pet.BreedID, PaletteID: pet.PaletteID, Color: pet.Color, Figure: figure, Rarity: pet.Rarity, Level: pet.Level,
		Experience: pet.Experience, Energy: pet.Energy, Happiness: pet.Happiness, Respect: pet.Respect, RoomID: pet.RoomID,
		X: pet.X, Y: pet.Y, Z: pet.Z, Rotation: pet.Rotation, State: pet.State, HasSaddle: pet.HasSaddle, PublicRide: pet.PublicRide,
		PublicBreed: pet.PublicBreed, GrowAt: pet.GrowAt, DieAt: pet.DieAt, CreatedAt: pet.CreatedAt, UpdatedAt: pet.UpdatedAt,
		DeletedAt: pet.DeletedAt, Version: pet.Version}
}
