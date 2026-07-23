package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	craftingrecord "github.com/niflaot/pixels/internal/realm/crafting/record"
)

// CreateRecipe creates one recipe and its ingredients atomically.
func (repository *Repository) CreateRecipe(ctx context.Context, input craftingrecord.CreateRecipe) (craftingrecord.Recipe, error) {
	var recipeID int64
	err := repository.WithinTransaction(ctx, func(txCtx context.Context) error {
		executor := repository.executorFor(txCtx)
		query := `insert into crafting_recipes(altar_definition_id,name,reward_definition_id,secret,limited,remaining,achievement_code) values($1,$2,$3,$4,$5,$6,nullif($7,'')) returning id`
		if err := executor.QueryRow(txCtx, query, input.AltarDefinitionID, input.Name, input.RewardDefinitionID, input.Secret, input.Limited, input.Remaining, input.AchievementCode).Scan(&recipeID); err != nil {
			return err
		}
		for _, ingredient := range input.Ingredients {
			if _, err := executor.Exec(txCtx, `insert into crafting_recipe_ingredients(recipe_id,ingredient_definition_id,amount) values($1,$2,$3)`, recipeID, ingredient.DefinitionID, ingredient.Amount); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return craftingrecord.Recipe{}, err
	}
	recipe, _, err := repository.Recipe(ctx, recipeID, true)
	return recipe, err
}

// UpdateRecipe applies one optimistic recipe patch.
func (repository *Repository) UpdateRecipe(ctx context.Context, id int64, patch craftingrecord.RecipePatch) (craftingrecord.Recipe, bool, error) {
	updated := false
	err := repository.WithinTransaction(ctx, func(txCtx context.Context) error {
		current, found, err := repository.Recipe(txCtx, id, true)
		if err != nil || !found {
			return err
		}
		if current.Version != patch.Version {
			return craftingrecord.ErrConflict
		}
		reward, secret, limited, remaining, achievement := current.RewardDefinitionID, current.Secret, current.Limited, current.Remaining, current.AchievementCode
		if patch.RewardDefinitionID != nil {
			reward = *patch.RewardDefinitionID
		}
		if patch.Secret != nil {
			secret = *patch.Secret
		}
		if patch.Limited != nil {
			limited = *patch.Limited
		}
		if patch.Remaining != nil {
			remaining = patch.Remaining
		}
		if patch.ClearRemaining {
			remaining = nil
		}
		if patch.AchievementCode != nil {
			achievement = *patch.AchievementCode
		}
		tag, err := repository.executorFor(txCtx).Exec(txCtx, `update crafting_recipes set reward_definition_id=$1,secret=$2,limited=$3,remaining=$4,achievement_code=nullif($5,''),updated_at=now(),version=version+1 where id=$6 and version=$7`, reward, secret, limited, remaining, achievement, id, patch.Version)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return craftingrecord.ErrConflict
		}
		if patch.ReplaceIngredients {
			if _, err = repository.executorFor(txCtx).Exec(txCtx, `delete from crafting_recipe_ingredients where recipe_id=$1`, id); err != nil {
				return err
			}
			for _, ingredient := range patch.Ingredients {
				if _, err = repository.executorFor(txCtx).Exec(txCtx, `insert into crafting_recipe_ingredients(recipe_id,ingredient_definition_id,amount) values($1,$2,$3)`, id, ingredient.DefinitionID, ingredient.Amount); err != nil {
					return err
				}
			}
		}
		updated = true
		return nil
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return craftingrecord.Recipe{}, false, nil
	}
	if err != nil {
		return craftingrecord.Recipe{}, false, err
	}
	recipe, found, err := repository.Recipe(ctx, id, true)
	return recipe, updated && found, err
}

// DisableRecipe soft-disables one recipe.
func (repository *Repository) DisableRecipe(ctx context.Context, id int64) (bool, error) {
	tag, err := repository.executorFor(ctx).Exec(ctx, `update crafting_recipes set enabled=false,updated_at=now(),version=version+1 where id=$1 and enabled`, id)
	return tag.RowsAffected() > 0, err
}
