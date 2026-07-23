// Package wired exposes protected WIRED room administration routes.
package wired

import (
	"time"

	wiredrecord "github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/registry"
)

// ConfigRequest contains an optimistic WIRED configuration replacement.
type ConfigRequest struct {
	// ExpectedVersion stores zero for first save or the current durable version.
	ExpectedVersion int64 `json:"expectedVersion"`
	// IntParams stores behavior-specific integer editor fields.
	IntParams []int32 `json:"intParams"`
	// StringParam stores behavior-specific text.
	StringParam string `json:"stringParam"`
	// SelectionMode stores Nitro's selection mode from zero through three.
	SelectionMode int32 `json:"selectionMode"`
	// DelayPulses stores 500 millisecond action delay units.
	DelayPulses int32 `json:"delayPulses"`
	// TargetIDs stores ordered furniture identifiers from the same room.
	TargetIDs []int64 `json:"targetIds"`
}

// RewardRequest contains one normalized durable reward option.
type RewardRequest struct {
	// Kind identifies furniture, badge, credits, currency, respect, or catalog_offer.
	Kind string `json:"kind"`
	// Reference stores the reward capability identifier.
	Reference string `json:"reference"`
	// Amount stores the positive delivered quantity.
	Amount int64 `json:"amount"`
	// Weight stores the positive integer selection weight.
	Weight int32 `json:"weight"`
	// Stock stores optional remaining global stock.
	Stock *int64 `json:"stock,omitempty"`
}

// RewardsRequest contains an atomic reward-list replacement.
type RewardsRequest struct {
	// Items stores normalized reward options in selection order.
	Items []RewardRequest `json:"items"`
}

// ConfigResponse contains one persisted WIRED node and descriptor.
type ConfigResponse struct {
	// Config stores the durable settings.
	Config wiredrecord.Config `json:"config"`
	// Descriptor stores canonical runtime behavior metadata.
	Descriptor registry.Descriptor `json:"descriptor"`
}

// RoomResponse contains room WIRED nodes and compilation state.
type RoomResponse struct {
	// RoomID identifies the room.
	RoomID int64 `json:"roomId"`
	// Loaded reports whether a runtime generation is active.
	Loaded bool `json:"loaded"`
	// Items stores configured nodes.
	Items []ConfigResponse `json:"items"`
}

// TraceResponse contains a sanitized bounded execution trace.
type TraceResponse struct {
	// ID identifies the source event.
	ID uint64 `json:"id"`
	// Kind stores the trigger kind numeric code.
	Kind uint8 `json:"kind"`
	// Stacks stores visited stack count.
	Stacks int `json:"stacks"`
	// Effects stores attempted effect count.
	Effects int `json:"effects"`
	// BudgetExhausted reports a safety-budget stop.
	BudgetExhausted bool `json:"budgetExhausted"`
	// StartedAt stores trace start time.
	StartedAt time.Time `json:"startedAt"`
	// Duration stores elapsed execution duration.
	Duration time.Duration `json:"durationNanoseconds"`
}

// ActionResponse reports one completed administrative action.
type ActionResponse struct {
	// Success reports whether the requested transition completed.
	Success bool `json:"success"`
}

// VisibilityRequest contains an optimistic room-level WIRED display mutation.
type VisibilityRequest struct {
	// ExpectedVersion prevents overwriting another room settings mutation.
	ExpectedVersion int64 `json:"expectedVersion"`
	// HideBoxes controls whether Nitro receives WIRED configuration boxes on entry.
	HideBoxes bool `json:"hideBoxes"`
}

// VisibilityResponse reports the committed room-level WIRED display setting.
type VisibilityResponse struct {
	// RoomID identifies the updated room.
	RoomID int64 `json:"roomId"`
	// HideBoxes reports the durable visibility setting.
	HideBoxes bool `json:"hideBoxes"`
	// Version stores the new optimistic room version.
	Version int64 `json:"version"`
}

// RegistryResponse contains the immutable runtime manifest.
type RegistryResponse struct {
	// Source identifies the audited upstream inventory.
	Source string `json:"source"`
	// Total stores canonical descriptor count.
	Total int `json:"total"`
	// Items stores descriptors in stable key order.
	Items []registry.Descriptor `json:"items"`
}
