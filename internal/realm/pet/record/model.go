// Package record defines durable pet records and persistence boundaries.
package record

import "time"

const (
	// StateInventory identifies a pet held outside a room.
	StateInventory = "inventory"
	// StateRoom identifies a pet placed in a room.
	StateRoom = "room"
	// StateBreedingReserved identifies a pet reserved by a breeding session.
	StateBreedingReserved = "breeding_reserved"
	// StateHarvested identifies a consumed monsterplant reward.
	StateHarvested = "harvested"
	// StateComposted identifies a composted monsterplant.
	StateComposted = "composted"
)

// AppearancePart stores one custom renderer part.
type AppearancePart struct {
	// LayerID identifies the renderer layer.
	LayerID int32
	// PartID identifies the custom part.
	PartID int32
	// PaletteID identifies the custom palette.
	PaletteID int32
}

// Pet stores one durable pet aggregate.
type Pet struct {
	// ID identifies the pet.
	ID int64
	// OwnerPlayerID identifies the current owner.
	OwnerPlayerID int64
	// OwnerName stores the joined owner display name.
	OwnerName string
	// Name stores the visible pet name.
	Name string
	// TypeID identifies the species.
	TypeID int32
	// BreedID identifies the breed.
	BreedID int32
	// PaletteID identifies the renderer palette.
	PaletteID int32
	// Color stores normalized hexadecimal color.
	Color string
	// Rarity stores the rarity category.
	Rarity int32
	// Parts stores ordered custom appearance parts.
	Parts []AppearancePart
	// Level stores the current level.
	Level int32
	// Experience stores accumulated experience.
	Experience int32
	// Energy stores materialized energy.
	Energy int32
	// Happiness stores materialized happiness.
	Happiness int32
	// Respect stores accumulated respect.
	Respect int32
	// StatsAt stores the energy and happiness materialization time.
	StatsAt time.Time
	// RoomID identifies placement and is nil in inventory.
	RoomID *int64
	// X stores the optional placed x coordinate.
	X *int
	// Y stores the optional placed y coordinate.
	Y *int
	// Z stores the optional placed height.
	Z *float64
	// Rotation stores the optional placed rotation.
	Rotation *int16
	// Posture stores the current renderer posture.
	Posture string
	// HasSaddle reports equipped saddle state.
	HasSaddle bool
	// CanBreed reports whether a completed breeding has not consumed eligibility.
	CanBreed bool
	// PublicRide reports public riding permission.
	PublicRide bool
	// PublicBreed reports public breeding permission.
	PublicBreed bool
	// GrowAt stores monsterplant maturity.
	GrowAt *time.Time
	// DieAt stores monsterplant death.
	DieAt *time.Time
	// State stores the lifecycle state.
	State string
	// CreatedAt stores creation time.
	CreatedAt time.Time
	// UpdatedAt stores last durable mutation time.
	UpdatedAt time.Time
	// DeletedAt stores soft deletion time.
	DeletedAt *time.Time
	// Version stores optimistic mutation state.
	Version int64
}

// AdminFilter stores bounded protected pet-list filters.
type AdminFilter struct {
	// OwnerPlayerID optionally restricts one owner.
	OwnerPlayerID *int64
	// Name optionally matches a case-insensitive name fragment.
	Name string
	// TypeID optionally restricts one species.
	TypeID *int32
	// RoomID optionally restricts one placed room.
	RoomID *int64
	// State optionally restricts one durable lifecycle state.
	State string
	// IncludeDeleted includes soft-deleted records.
	IncludeDeleted bool
	// Cursor returns identifiers strictly after this value.
	Cursor int64
	// Limit bounds returned rows.
	Limit int
}

// AdminPatch stores one optimistic protected pet mutation.
type AdminPatch struct {
	// Name optionally replaces the visible name.
	Name *string
	// BreedID optionally replaces the renderer breed.
	BreedID *int32
	// PaletteID optionally replaces the renderer palette.
	PaletteID *int32
	// Color optionally replaces the renderer color.
	Color *string
	// PublicRide optionally replaces public riding access.
	PublicRide *bool
	// PublicBreed optionally replaces public breeding access.
	PublicBreed *bool
	// Version stores the required optimistic version.
	Version int64
}

// Inventory reports whether the pet is held outside a room.
func (pet Pet) Inventory() bool { return pet.RoomID == nil && pet.State == StateInventory }

// PlantState stores derived monsterplant lifecycle flags.
type PlantState struct {
	// GrowthStage stores the renderer stage from one through seven.
	GrowthStage int32
	// FullyGrown reports maturity.
	FullyGrown bool
	// Dead reports expired lifetime.
	Dead bool
	// CanHarvest reports one available harvest.
	CanHarvest bool
	// CanRevive reports one available revival.
	CanRevive bool
	// RemainingGrowSeconds stores bounded time until maturity.
	RemainingGrowSeconds int32
	// RemainingLifeSeconds stores bounded time until death.
	RemainingLifeSeconds int32
}

// DerivePlantState computes lifecycle flags from absolute deadlines.
func (pet Pet) DerivePlantState(now time.Time, species Species) PlantState {
	if !species.Plant || pet.GrowAt == nil || pet.DieAt == nil {
		return PlantState{}
	}
	dead := !now.Before(*pet.DieAt)
	fullyGrown := !dead && !now.Before(*pet.GrowAt)
	return PlantState{
		GrowthStage: plantGrowthStage(pet.CreatedAt, *pet.GrowAt, now), FullyGrown: fullyGrown, Dead: dead, CanHarvest: fullyGrown && pet.State == StateRoom,
		CanRevive: dead && pet.State == StateRoom, RemainingGrowSeconds: remainingSeconds(now, *pet.GrowAt),
		RemainingLifeSeconds: remainingSeconds(now, *pet.DieAt),
	}
}

// plantGrowthStage derives one stable seven-stage renderer value.
func plantGrowthStage(createdAt time.Time, growAt time.Time, now time.Time) int32 {
	if !now.Before(growAt) {
		return 7
	}
	if createdAt.IsZero() || !createdAt.Before(growAt) || !now.After(createdAt) {
		return 1
	}
	total, elapsed := growAt.Unix()-createdAt.Unix(), now.Unix()-createdAt.Unix()
	stage := int32(1 + elapsed*6/total)
	if stage > 6 {
		return 6
	}
	return stage
}

// remainingSeconds returns a non-negative duration bounded to int32.
func remainingSeconds(now time.Time, deadline time.Time) int32 {
	seconds := int64(deadline.Sub(now) / time.Second)
	if seconds <= 0 {
		return 0
	}
	if seconds > int64(^uint32(0)>>1) {
		return int32(^uint32(0) >> 1)
	}
	return int32(seconds)
}
