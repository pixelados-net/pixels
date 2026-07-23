// Package record defines crafting persistence records and domain contracts.
package record

import "time"

// Altar registers one furniture definition as a crafting surface.
type Altar struct {
	// DefinitionID identifies the altar furniture definition.
	DefinitionID int64
	// Enabled reports whether players may use the altar.
	Enabled bool
	// CreatedAt stores the registration time.
	CreatedAt time.Time
	// UpdatedAt stores the last mutation time.
	UpdatedAt time.Time
	// Version stores optimistic administration state.
	Version int64
}

// Ingredient stores one aggregated recipe requirement.
type Ingredient struct {
	// DefinitionID identifies the required furniture definition.
	DefinitionID int64
	// Name stores the Nitro furniture product name.
	Name string
	// Amount stores the exact required instance count.
	Amount int32
}

// Recipe stores one altar recipe and its complete ingredient set.
type Recipe struct {
	// ID identifies the recipe.
	ID int64
	// AltarDefinitionID identifies the owning altar definition.
	AltarDefinitionID int64
	// Name stores the unique wire-facing recipe name.
	Name string
	// RewardDefinitionID identifies the granted furniture definition.
	RewardDefinitionID int64
	// RewardName stores the Nitro reward product name.
	RewardName string
	// Secret reports whether discovery is required for listing and named crafting.
	Secret bool
	// Limited reports whether successful crafts consume remaining stock.
	Limited bool
	// Remaining stores limited stock or nil for unlimited recipes.
	Remaining *int32
	// AchievementCode stores the optional post-commit achievement hook.
	AchievementCode string
	// Enabled reports whether the recipe accepts operations.
	Enabled bool
	// Ingredients stores the complete exact bag.
	Ingredients []Ingredient
	// CreatedAt stores the creation time.
	CreatedAt time.Time
	// UpdatedAt stores the last mutation time.
	UpdatedAt time.Time
	// Version stores optimistic administration state.
	Version int64
}

// Product stores one Nitro crafting list entry.
type Product struct {
	// RecipeName stores the selectable recipe name.
	RecipeName string
	// RewardName stores the resulting furniture product name.
	RewardName string
}

// KnownRecipe stores one durable player discovery.
type KnownRecipe struct {
	// PlayerID identifies the discovering player.
	PlayerID int64
	// RecipeID identifies the discovered recipe.
	RecipeID int64
	// RecipeName stores the administration display name.
	RecipeName string
	// DiscoveredAt stores the first successful discovery time.
	DiscoveredAt time.Time
}

// Prize stores one recycler pool entry and its wire metadata.
type Prize struct {
	// Tier stores the rarity tier from one through five.
	Tier int32
	// RewardDefinitionID identifies the granted definition.
	RewardDefinitionID int64
	// RewardName stores the furniture product name.
	RewardName string
	// TypeCode stores Nitro's lower-case floor or wall discriminator.
	TypeCode string
	// SpriteID stores the Nitro furniture sprite identifier.
	SpriteID int32
}

// Audit stores durable administrative attribution.
type Audit struct {
	// ActorPlayerID identifies the administrative actor.
	ActorPlayerID int64
	// Action stores the bounded mutation name.
	Action string
	// EntityKind stores the mutated aggregate kind.
	EntityKind string
	// EntityID stores an optional aggregate identifier.
	EntityID int64
	// Reason stores the required human-readable attribution.
	Reason string
}
