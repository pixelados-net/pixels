package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	craftingrecord "github.com/niflaot/pixels/internal/realm/crafting/record"
)

const recipeSelect = `select r.id,r.altar_definition_id,r.name,r.reward_definition_id,rewards.name,r.secret,r.limited,r.remaining,coalesce(r.achievement_code,''),r.enabled,r.created_at,r.updated_at,r.version,i.ingredient_definition_id,ingredients.name,i.amount from crafting_recipes r join furniture_definitions rewards on rewards.id=r.reward_definition_id left join crafting_recipe_ingredients i on i.recipe_id=r.id left join furniture_definitions ingredients on ingredients.id=i.ingredient_definition_id`

// Recipes returns grouped recipes and ingredients in one query.
func (repository *Repository) Recipes(ctx context.Context, filter craftingrecord.RecipeFilter) ([]craftingrecord.Recipe, error) {
	query := recipeSelect + ` join crafting_altars a on a.definition_id=r.altar_definition_id left join player_crafting_known_recipes known on known.recipe_id=r.id and known.player_id=$2 where ($1=0 or r.altar_definition_id=$1) and ($4 or (r.enabled and a.enabled)) and ($3 or not r.secret or known.recipe_id is not null) order by r.id,i.ingredient_definition_id`
	rows, err := repository.executorFor(ctx).Query(ctx, query, filter.AltarDefinitionID, filter.PlayerID, filter.IncludeUnknownSecrets, filter.IncludeDisabled)
	if err != nil {
		return nil, fmt.Errorf("list crafting recipes: %w", err)
	}
	defer rows.Close()
	return scanRecipes(rows)
}

// Recipe finds one recipe by identifier.
func (repository *Repository) Recipe(ctx context.Context, id int64, includeDisabled bool) (craftingrecord.Recipe, bool, error) {
	return repository.oneRecipe(ctx, recipeSelect+` where r.id=$1 and ($2 or r.enabled) order by i.ingredient_definition_id`, id, includeDisabled)
}

// RecipeByName finds one enabled altar recipe by its wire name.
func (repository *Repository) RecipeByName(ctx context.Context, altarID int64, name string) (craftingrecord.Recipe, bool, error) {
	return repository.oneRecipe(ctx, recipeSelect+` join crafting_altars a on a.definition_id=r.altar_definition_id where r.altar_definition_id=$1 and r.name=$2 and r.enabled and a.enabled order by i.ingredient_definition_id`, altarID, name)
}

func (repository *Repository) oneRecipe(ctx context.Context, query string, arguments ...any) (craftingrecord.Recipe, bool, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, query, arguments...)
	if err != nil {
		return craftingrecord.Recipe{}, false, err
	}
	defer rows.Close()
	recipes, err := scanRecipes(rows)
	if err != nil || len(recipes) == 0 {
		return craftingrecord.Recipe{}, false, err
	}
	return recipes[0], true, nil
}

func scanRecipes(rows pgx.Rows) ([]craftingrecord.Recipe, error) {
	recipes := make([]craftingrecord.Recipe, 0, 8)
	for rows.Next() {
		var recipe craftingrecord.Recipe
		var definitionID *int64
		var name *string
		var amount *int32
		err := rows.Scan(&recipe.ID, &recipe.AltarDefinitionID, &recipe.Name, &recipe.RewardDefinitionID, &recipe.RewardName, &recipe.Secret, &recipe.Limited, &recipe.Remaining, &recipe.AchievementCode, &recipe.Enabled, &recipe.CreatedAt, &recipe.UpdatedAt, &recipe.Version, &definitionID, &name, &amount)
		if err != nil {
			return nil, err
		}
		if len(recipes) == 0 || recipes[len(recipes)-1].ID != recipe.ID {
			recipes = append(recipes, recipe)
		}
		if definitionID != nil {
			recipes[len(recipes)-1].Ingredients = append(recipes[len(recipes)-1].Ingredients, craftingrecord.Ingredient{DefinitionID: *definitionID, Name: *name, Amount: *amount})
		}
	}
	return recipes, rows.Err()
}

// ConsumeLimited decrements positive stock atomically.
func (repository *Repository) ConsumeLimited(ctx context.Context, recipeID int64) (int32, bool, error) {
	var remaining int32
	err := repository.executorFor(ctx).QueryRow(ctx, `update crafting_recipes set remaining=remaining-1,updated_at=now(),version=version+1 where id=$1 and enabled and limited and remaining>0 returning remaining`, recipeID).Scan(&remaining)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil
	}
	return remaining, err == nil, err
}
