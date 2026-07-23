// Package record defines WIRED persistence records and domain contracts.
package record

import "context"

// Snapshot stores a selected furniture state captured for match/apply behavior.
type Snapshot struct {
	// State stores furniture extra data.
	State string
	// X stores the captured room tile coordinate.
	X int
	// Y stores the captured room tile coordinate.
	Y int
	// Z stores the captured room elevation.
	Z float64
	// Rotation stores the captured furniture rotation.
	Rotation int
	// Present reports whether snapshot fields were captured.
	Present bool
}

// Target stores one ordered selected furniture item.
type Target struct {
	// ItemID identifies selected furniture.
	ItemID int64
	// SpriteID identifies the selected furniture definition for type matching.
	SpriteID int32
	// Snapshot stores optional captured furniture state.
	Snapshot Snapshot
}

// Config stores one durable WIRED node configuration.
type Config struct {
	// ItemID identifies the WIRED furniture item.
	ItemID int64
	// RoomID identifies the room containing the item.
	RoomID int64
	// Interaction stores the canonical registered behavior key.
	Interaction string
	// SpriteID stores the client furniture sprite identifier.
	SpriteID int32
	// X stores the WIRED item's stack tile coordinate.
	X int
	// Y stores the WIRED item's stack tile coordinate.
	Y int
	// IntParams stores protocol-facing integer settings.
	IntParams []int32
	// StringParam stores protocol-facing text settings.
	StringParam string
	// SelectionMode stores the target matching policy.
	SelectionMode int32
	// DelayPulses stores action delay in 500 millisecond pulses.
	DelayPulses int32
	// Version stores optimistic configuration version.
	Version int64
	// Targets stores ordered selected furniture.
	Targets []Target
}

// Store persists and loads WIRED configurations.
type Store interface {
	// LoadRoom loads every configured node placed in one room.
	LoadRoom(context.Context, int64) ([]Config, error)
	// Find loads one configured or default node in one room.
	Find(context.Context, int64, int64) (Config, bool, error)
	// Save atomically replaces settings and selected items.
	Save(context.Context, Config, int64) (Config, error)
	// Capture atomically replaces selected-item snapshots.
	Capture(context.Context, int64, int64) ([]Target, error)
	// SaveRewardConfig atomically replaces settings, targets, and reward definitions.
	SaveRewardConfig(context.Context, Config, int64, []Reward) (Config, error)
	// CleanupItem removes a picked-up WIRED box configuration and target references.
	CleanupItem(context.Context, int64) error
}

// Reward stores one normalized durable reward option.
type Reward struct {
	// ID identifies the durable reward row.
	ID int64
	// Ordinal stabilizes editor and selection order.
	Ordinal int
	// Kind identifies the delivery capability.
	Kind string
	// Reference stores a definition, badge, currency, or offer identifier.
	Reference string
	// Amount stores the positive delivered quantity.
	Amount int64
	// Weight stores the positive integer selection weight.
	Weight int32
	// Stock stores optional remaining global stock.
	Stock *int64
}

// ClaimStatus classifies one durable reward claim.
type ClaimStatus uint8

const (
	// ClaimUnavailable reports absent or invalid reward configuration.
	ClaimUnavailable ClaimStatus = iota
	// ClaimAlreadyReceived reports an existing claim in the active period.
	ClaimAlreadyReceived
	// ClaimOutOfStock reports exhausted reward stock.
	ClaimOutOfStock
	// ClaimMissed reports a weighted probability outcome without a prize.
	ClaimMissed
	// ClaimDelivered reports a committed reward delivery.
	ClaimDelivered
)

// RewardStore persists reward definitions and atomic claims.
type RewardStore interface {
	// ListRewards lists normalized rewards for one WIRED effect.
	ListRewards(context.Context, int64) ([]Reward, error)
	// ReplaceRewards atomically replaces normalized reward definitions.
	ReplaceRewards(context.Context, int64, []Reward) error
	// ClaimReward atomically selects, delivers, and records one period claim.
	ClaimReward(context.Context, int64, int64, string, bool, uint64, func(context.Context, Reward) error) (ClaimStatus, Reward, error)
}
