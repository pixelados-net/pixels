package database

import (
	"context"
	"fmt"

	craftingrecord "github.com/niflaot/pixels/internal/realm/crafting/record"
)

// Known reports whether one player discovered one recipe.
func (repository *Repository) Known(ctx context.Context, playerID int64, recipeID int64) (bool, error) {
	var known bool
	err := repository.executorFor(ctx).QueryRow(ctx, `select exists(select 1 from player_crafting_known_recipes where player_id=$1 and recipe_id=$2)`, playerID, recipeID).Scan(&known)
	return known, err
}

// KnownRecipes lists one player's durable discoveries.
func (repository *Repository) KnownRecipes(ctx context.Context, playerID int64) ([]craftingrecord.KnownRecipe, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, `select k.player_id,k.recipe_id,r.name,k.discovered_at from player_crafting_known_recipes k join crafting_recipes r on r.id=k.recipe_id where k.player_id=$1 order by k.discovered_at,k.recipe_id`, playerID)
	if err != nil {
		return nil, fmt.Errorf("list known crafting recipes: %w", err)
	}
	defer rows.Close()
	known := make([]craftingrecord.KnownRecipe, 0, 8)
	for rows.Next() {
		var item craftingrecord.KnownRecipe
		if err = rows.Scan(&item.PlayerID, &item.RecipeID, &item.RecipeName, &item.DiscoveredAt); err != nil {
			return nil, err
		}
		known = append(known, item)
	}
	return known, rows.Err()
}

// RememberRecipe inserts one discovery idempotently.
func (repository *Repository) RememberRecipe(ctx context.Context, playerID int64, recipeID int64) (bool, error) {
	tag, err := repository.executorFor(ctx).Exec(ctx, `insert into player_crafting_known_recipes(player_id,recipe_id) values($1,$2) on conflict do nothing`, playerID, recipeID)
	return tag.RowsAffected() > 0, err
}

// ForgetRecipe removes one discovery idempotently.
func (repository *Repository) ForgetRecipe(ctx context.Context, playerID int64, recipeID int64) (bool, error) {
	tag, err := repository.executorFor(ctx).Exec(ctx, `delete from player_crafting_known_recipes where player_id=$1 and recipe_id=$2`, playerID, recipeID)
	return tag.RowsAffected() > 0, err
}
