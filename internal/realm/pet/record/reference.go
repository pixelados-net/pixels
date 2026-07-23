package record

import "time"

// Species stores immutable pet species reference data.
type Species struct {
	// TypeID identifies the renderer species.
	TypeID int32
	// Slug stores the stable species key.
	Slug string
	// DisplayKey stores the localized name key.
	DisplayKey string
	// BehaviorKind selects runtime behavior.
	BehaviorKind string
	// MaximumLevel stores the level cap.
	MaximumLevel int32
	// Rideable reports horse-like support.
	Rideable bool
	// Breedable reports breeding support.
	Breedable bool
	// Plant reports monsterplant lifecycle support.
	Plant bool
	// Enabled reports whether creation is allowed.
	Enabled bool
}

// Breed stores one species appearance option.
type Breed struct {
	// TypeID identifies the species.
	TypeID int32
	// BreedID identifies the breed.
	BreedID int32
	// PaletteID identifies the base palette.
	PaletteID int32
	// Color stores the default normalized color.
	Color string
	// Sellable reports catalog availability.
	Sellable bool
	// Rarity stores the rarity category.
	Rarity int32
	// Enabled reports runtime availability.
	Enabled bool
}

// Command stores one trainable pet command.
type Command struct {
	// ID identifies the protocol command.
	ID int32
	// NameKey stores the localized command key.
	NameKey string
	// RequiredLevel stores the unlock level.
	RequiredLevel int32
	// Family stores behavior grouping.
	Family string
	// EnergyCost stores energy consumed on success.
	EnergyCost int32
	// HappinessCost stores happiness consumed on success.
	HappinessCost int32
	// ExperienceReward stores experience gained on success.
	ExperienceReward int32
	// Duration stores the bounded action duration.
	Duration time.Duration
	// Cooldown stores the bounded reuse cooldown.
	Cooldown time.Duration
	// Enabled reports runtime availability.
	Enabled bool
}

// ProductRule stores one typed furniture-to-pet product action.
type ProductRule struct {
	// DefinitionID identifies the furniture definition.
	DefinitionID int64
	// Kind identifies food, drink, toy, saddle, supplement, seed, or nest.
	Kind string
	// TypeID optionally restricts one species and uses -1 for any.
	TypeID int32
	// EnergyDelta stores the energy mutation.
	EnergyDelta int32
	// HappinessDelta stores the happiness mutation.
	HappinessDelta int32
	// ExperienceDelta stores the experience mutation.
	ExperienceDelta int32
	// Consumable reports whether the furniture item is consumed.
	Consumable bool
	// Enabled reports runtime availability.
	Enabled bool
}

// Vocal stores one weighted localized species vocalization.
type Vocal struct {
	// TypeID identifies the species.
	TypeID int32
	// Mood identifies the bounded behavior context.
	Mood string
	// TextKey identifies localized hotel-facing speech.
	TextKey string
	// Weight stores deterministic selection weight.
	Weight int32
	// Cooldown stores the minimum autonomous repeat interval.
	Cooldown time.Duration
	// Enabled reports runtime availability.
	Enabled bool
}

// BreedingRule stores one parent compatibility and result species rule.
type BreedingRule struct {
	// ParentOneTypeID identifies the lower canonical parent type.
	ParentOneTypeID int32
	// ParentTwoTypeID identifies the higher canonical parent type.
	ParentTwoTypeID int32
	// ResultTypeID identifies the offspring species.
	ResultTypeID int32
	// Enabled reports runtime availability.
	Enabled bool
}

// BreedingRace stores one weighted offspring appearance.
type BreedingRace struct {
	// ResultTypeID identifies the offspring species.
	ResultTypeID int32
	// BreedID identifies the result breed.
	BreedID int32
	// PaletteID identifies the result palette.
	PaletteID int32
	// Weight stores the bounded selection weight.
	Weight int32
	// Mutation reports whether the row represents a mutation.
	Mutation bool
	// Enabled reports runtime availability.
	Enabled bool
}
