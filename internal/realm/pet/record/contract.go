package record

import (
	"context"
	"time"
)

// GrantParams stores one idempotent pet grant.
type GrantParams struct {
	// OwnerPlayerID identifies the receiving owner.
	OwnerPlayerID int64
	// Name stores the validated pet name.
	Name string
	// TypeID identifies the species.
	TypeID int32
	// BreedID identifies the breed.
	BreedID int32
	// PaletteID identifies the renderer palette.
	PaletteID int32
	// Color stores the normalized hexadecimal color.
	Color string
	// Parts stores ordered custom renderer genetics.
	Parts []AppearancePart
	// OperationKey makes retries idempotent.
	OperationKey string
}

// Store persists pet aggregates and immutable reference data.
type Store interface {
	// WithinTransaction runs related mutations atomically.
	WithinTransaction(context.Context, func(context.Context) error) error
	// Find returns one live pet.
	Find(context.Context, int64) (Pet, bool, error)
	// ListAdmin returns a bounded protected pet page.
	ListAdmin(context.Context, AdminFilter) ([]Pet, error)
	// FindByOperation returns the pet produced by one completed idempotent operation.
	FindByOperation(context.Context, string) (Pet, bool, error)
	// Inventory lists one owner's inventory pets in identifier order.
	Inventory(context.Context, int64) ([]Pet, error)
	// Room lists all pets placed in one room without N+1 reads.
	Room(context.Context, int64) ([]Pet, error)
	// CountInventory counts inventory pets for one owner.
	CountInventory(context.Context, int64) (int, error)
	// Grant creates or returns one idempotently granted pet.
	Grant(context.Context, GrantParams) (Pet, bool, error)
	// Place compare-and-swaps an owned inventory pet into a room.
	Place(context.Context, int64, int64, int64, int, int, float64, int16, int64) (Pet, bool, error)
	// Pickup compare-and-swaps a placed pet back to inventory.
	Pickup(context.Context, int64, int64, int64, int64) (Pet, bool, error)
	// SavePosition persists a placed pet position by version.
	SavePosition(context.Context, int64, int64, int, int, float64, int16, int64) (int64, bool, error)
	// Respect applies one daily idempotent respect and experience grant.
	Respect(context.Context, int64, int64, int32, int) (Pet, bool, error)
	// UpdateFlags replaces public riding and breeding flags.
	UpdateFlags(context.Context, int64, int64, bool, bool, int64) (Pet, bool, error)
	// UpdateStats applies bounded stat deltas and recomputes level.
	UpdateStats(context.Context, int64, int32, int32, int32, int64) (Pet, bool, error)
	// SetSaddle replaces equipment state.
	SetSaddle(context.Context, int64, int64, bool, int64) (Pet, bool, error)
	// SetBreedingEligibility replaces one pet's consumable breeding eligibility.
	SetBreedingEligibility(context.Context, int64, bool, int64) (Pet, bool, error)
	// UpdateLifecycle replaces absolute monsterplant deadlines.
	UpdateLifecycle(context.Context, int64, int64, *time.Time, *time.Time, int64) (Pet, bool, error)
	// ConsumePlant atomically soft-deletes one eligible monsterplant state.
	ConsumePlant(context.Context, int64, int64, int64, string, int64) (bool, error)
	// SaveBreedingSession creates or confirms one nest-owned breeding session.
	SaveBreedingSession(context.Context, BreedingSession, int64) (BreedingSession, bool, error)
	// FindBreedingSession returns one active durable session.
	FindBreedingSession(context.Context, int64) (BreedingSession, bool, error)
	// SetBreedingSessionState compare-and-swaps one session state.
	SetBreedingSessionState(context.Context, int64, string, string, int64) (bool, error)
	// CancelBreedingRoom releases every active reservation in one closing room.
	CancelBreedingRoom(context.Context, int64) error
	// CancelBreedingPet releases every active reservation involving one pet.
	CancelBreedingPet(context.Context, int64, int64) error
	// SoftDelete removes one pet from active reads.
	SoftDelete(context.Context, int64, int64) (bool, error)
	// UpdateAdmin replaces protected mutable fields optimistically.
	UpdateAdmin(context.Context, int64, AdminPatch) (Pet, bool, error)
	// TransferOwner moves one inventory pet to another owner optimistically.
	TransferOwner(context.Context, int64, int64, int64) (Pet, bool, error)
	// DeleteAdmin soft-deletes one pet in any location optimistically.
	DeleteAdmin(context.Context, int64, int64) (bool, error)
	// AppendAudit records one protected mutation inside the caller transaction.
	AppendAudit(context.Context, int64, int64, string, string) error
	// AppendGlobalAudit records one protected reference mutation.
	AppendGlobalAudit(context.Context, int64, string, string) error
	// Species lists all species reference rows.
	Species(context.Context) ([]Species, error)
	// Breeds lists all breed reference rows.
	Breeds(context.Context) ([]Breed, error)
	// Commands lists every command and species association.
	Commands(context.Context) ([]Command, map[int32][]int32, error)
	// ProductRules lists typed pet product rules.
	ProductRules(context.Context) ([]ProductRule, error)
	// Vocals lists localized weighted species speech.
	Vocals(context.Context) ([]Vocal, error)
	// BreedingRules lists compatibility and weighted result appearance rules.
	BreedingRules(context.Context) ([]BreedingRule, []BreedingRace, error)
}
