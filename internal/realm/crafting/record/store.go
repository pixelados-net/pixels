package record

import "context"

// Store persists crafting, recycler, discovery, and audit state.
type Store interface {
	// WithinTransaction runs work in one shared PostgreSQL transaction.
	WithinTransaction(context.Context, func(context.Context) error) error
	// Altar finds one altar registration.
	Altar(context.Context, int64, bool) (Altar, bool, error)
	// ListAltars lists altar registrations with optional disabled rows.
	ListAltars(context.Context, bool) ([]Altar, error)
	// Recipes returns grouped recipes and ingredients in one query.
	Recipes(context.Context, RecipeFilter) ([]Recipe, error)
	// Recipe finds one recipe by identifier.
	Recipe(context.Context, int64, bool) (Recipe, bool, error)
	// RecipeByName finds one enabled altar recipe by its wire name.
	RecipeByName(context.Context, int64, string) (Recipe, bool, error)
	// Known reports whether one player discovered one recipe.
	Known(context.Context, int64, int64) (bool, error)
	// KnownRecipes lists one player's durable discoveries.
	KnownRecipes(context.Context, int64) ([]KnownRecipe, error)
	// RememberRecipe inserts one discovery idempotently.
	RememberRecipe(context.Context, int64, int64) (bool, error)
	// ForgetRecipe removes one discovery idempotently.
	ForgetRecipe(context.Context, int64, int64) (bool, error)
	// ConsumeLimited decrements positive stock atomically.
	ConsumeLimited(context.Context, int64) (int32, bool, error)
	// Prizes lists enabled recycler pool entries in rarity order.
	Prizes(context.Context) ([]Prize, error)
	// UpsertAltar creates or re-enables an altar.
	UpsertAltar(context.Context, int64) (Altar, bool, error)
	// DisableAltar disables an altar and all of its recipes.
	DisableAltar(context.Context, int64) (bool, error)
	// CreateRecipe creates one recipe and its ingredients atomically.
	CreateRecipe(context.Context, CreateRecipe) (Recipe, error)
	// UpdateRecipe applies one optimistic recipe patch.
	UpdateRecipe(context.Context, int64, RecipePatch) (Recipe, bool, error)
	// DisableRecipe soft-disables one recipe.
	DisableRecipe(context.Context, int64) (bool, error)
	// AddPrize inserts one recycler prize idempotently.
	AddPrize(context.Context, int32, int64) (bool, error)
	// DeletePrize removes one recycler prize idempotently.
	DeletePrize(context.Context, int32, int64) (bool, error)
	// InsertAudit appends one administrative mutation record.
	InsertAudit(context.Context, Audit) error
}
