package record

// RecipeFilter controls one grouped recipe read.
type RecipeFilter struct {
	// AltarDefinitionID restricts one altar.
	AltarDefinitionID int64
	// PlayerID enables player-specific secret visibility.
	PlayerID int64
	// IncludeUnknownSecrets includes undiscovered secret recipes.
	IncludeUnknownSecrets bool
	// IncludeDisabled includes disabled recipes and altars.
	IncludeDisabled bool
}

// CreateRecipe stores validated recipe creation input.
type CreateRecipe struct {
	// AltarDefinitionID identifies the owning altar.
	AltarDefinitionID int64
	// Name stores the unique wire-facing name.
	Name string
	// RewardDefinitionID identifies the granted furniture definition.
	RewardDefinitionID int64
	// Secret reports whether discovery is required.
	Secret bool
	// Limited reports whether remaining stock is consumed.
	Limited bool
	// Remaining stores initial limited stock.
	Remaining *int32
	// AchievementCode stores an optional achievement hook.
	AchievementCode string
	// Ingredients stores the complete exact bag.
	Ingredients []Ingredient
}

// RecipePatch stores one optimistic administrative update.
type RecipePatch struct {
	// RewardDefinitionID optionally replaces the reward.
	RewardDefinitionID *int64
	// Secret optionally replaces discovery policy.
	Secret *bool
	// Limited optionally replaces limited policy.
	Limited *bool
	// Remaining optionally replaces limited stock.
	Remaining *int32
	// ClearRemaining removes limited stock.
	ClearRemaining bool
	// AchievementCode optionally replaces the achievement hook.
	AchievementCode *string
	// Ingredients optionally replaces the complete exact bag.
	Ingredients []Ingredient
	// ReplaceIngredients reports whether Ingredients is authoritative.
	ReplaceIngredients bool
	// Version stores the required optimistic version.
	Version int64
}
