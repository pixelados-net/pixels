package recipe

import (
	"context"
	"errors"
	craftingrecord "github.com/niflaot/pixels/internal/realm/crafting/record"
)

// Open validates one placed altar and returns player-visible recipes.
func (service *Service) Open(ctx context.Context, playerID, roomID, altarItemID int64) ([]craftingrecord.Recipe, error) {
	altarID, err := service.resolveAltar(ctx, roomID, altarItemID)
	if err != nil {
		return nil, err
	}
	recipes, err := service.store.Recipes(ctx, craftingrecord.RecipeFilter{AltarDefinitionID: altarID, PlayerID: playerID})
	if err != nil {
		return nil, err
	}
	service.mutex.Lock()
	service.altars[playerID] = altarID
	service.mutex.Unlock()
	return recipes, nil
}

// Recipe returns one visible recipe from the player's currently opened altar.
func (service *Service) Recipe(ctx context.Context, playerID int64, name string) (craftingrecord.Recipe, error) {
	service.mutex.RLock()
	altarID, found := service.altars[playerID]
	service.mutex.RUnlock()
	if !found {
		return craftingrecord.Recipe{}, craftingrecord.ErrAltarNotFound
	}
	recipe, found, err := service.store.RecipeByName(ctx, altarID, name)
	if err != nil {
		return craftingrecord.Recipe{}, err
	}
	if !found {
		return craftingrecord.Recipe{}, craftingrecord.ErrRecipeNotFound
	}
	if recipe.Secret {
		known, knownErr := service.store.Known(ctx, playerID, recipe.ID)
		if knownErr != nil {
			return craftingrecord.Recipe{}, knownErr
		}
		if !known {
			return craftingrecord.Recipe{}, craftingrecord.ErrRecipeNotFound
		}
	}
	return recipe, nil
}

// Craft crafts one visible named recipe atomically.
func (service *Service) Craft(ctx context.Context, playerID, roomID, altarItemID int64, name string) (Result, error) {
	altarID, err := service.resolveAltar(ctx, roomID, altarItemID)
	if err != nil {
		return Result{}, err
	}
	recipe, found, err := service.store.RecipeByName(ctx, altarID, name)
	if err != nil {
		return Result{}, err
	}
	if !found {
		return Result{}, craftingrecord.ErrRecipeNotFound
	}
	if recipe.Secret {
		known, knownErr := service.store.Known(ctx, playerID, recipe.ID)
		if knownErr != nil {
			return Result{}, knownErr
		}
		if !known {
			return Result{}, craftingrecord.ErrRecipeNotFound
		}
	}
	return service.commit(ctx, playerID, recipe, nil)
}

// CraftSecret matches and crafts one exact free-form inventory bag.
func (service *Service) CraftSecret(ctx context.Context, playerID, roomID, altarItemID int64, itemIDs []int64) (Result, error) {
	altarID, err := service.resolveAltar(ctx, roomID, altarItemID)
	if err != nil {
		return Result{}, err
	}
	bag, err := service.explicitBag(ctx, playerID, itemIDs)
	if err != nil {
		return Result{}, err
	}
	recipes, err := service.store.Recipes(ctx, craftingrecord.RecipeFilter{AltarDefinitionID: altarID, PlayerID: playerID, IncludeUnknownSecrets: true})
	if err != nil {
		return Result{}, err
	}
	for _, match := range MatchRecipes(recipes, bag) {
		if match.Exact {
			return service.commit(ctx, playerID, match.Recipe, itemIDs)
		}
	}
	return Result{}, craftingrecord.ErrIngredients
}

// Hint reports partial and exact unknown-secret matches.
func (service *Service) Hint(ctx context.Context, playerID, roomID, altarItemID int64, itemIDs []int64) (int32, bool, error) {
	altarID, err := service.resolveAltar(ctx, roomID, altarItemID)
	if err != nil {
		return 0, false, err
	}
	bag, err := service.explicitBag(ctx, playerID, itemIDs)
	if err != nil {
		return 0, false, err
	}
	recipes, err := service.store.Recipes(ctx, craftingrecord.RecipeFilter{AltarDefinitionID: altarID, PlayerID: playerID, IncludeUnknownSecrets: true})
	if err != nil {
		return 0, false, err
	}
	unknown := recipes[:0]
	for _, candidate := range recipes {
		if !candidate.Secret {
			continue
		}
		known, knownErr := service.store.Known(ctx, playerID, candidate.ID)
		if knownErr != nil {
			return 0, false, knownErr
		}
		if !known {
			unknown = append(unknown, candidate)
		}
	}
	count, exact := Hint(unknown, bag)
	return count, exact, nil
}

// Close releases one player's ephemeral altar selection.
func (service *Service) Close(playerID int64) {
	service.mutex.Lock()
	delete(service.altars, playerID)
	service.mutex.Unlock()
}
func (service *Service) resolveAltar(ctx context.Context, roomID, itemID int64) (int64, error) {
	if !service.config.Enabled {
		return 0, craftingrecord.ErrDisabled
	}
	item, found, err := service.furniture.FindItemByID(ctx, itemID)
	if err != nil {
		return 0, err
	}
	if !found || item.RoomID == nil || *item.RoomID != roomID {
		return 0, craftingrecord.ErrAltarNotFound
	}
	_, found, err = service.store.Altar(ctx, item.DefinitionID, false)
	if err != nil {
		return 0, err
	}
	if !found {
		return 0, craftingrecord.ErrAltarNotFound
	}
	return item.DefinitionID, nil
}
func (service *Service) explicitBag(ctx context.Context, playerID int64, itemIDs []int64) (map[int64]int32, error) {
	if len(itemIDs) == 0 {
		return nil, craftingrecord.ErrIngredients
	}
	bag := make(map[int64]int32, len(itemIDs))
	seen := make(map[int64]struct{}, len(itemIDs))
	for _, itemID := range itemIDs {
		if _, duplicate := seen[itemID]; duplicate {
			return nil, craftingrecord.ErrIngredients
		}
		seen[itemID] = struct{}{}
		item, found, err := service.furniture.FindItemByID(ctx, itemID)
		if err != nil {
			return nil, err
		}
		if !found || item.OwnerPlayerID != playerID || !item.InInventory() || item.MarketplaceReserved {
			return nil, craftingrecord.ErrItemUnavailable
		}
		bag[item.DefinitionID]++
	}
	return bag, nil
}

// Expected reports whether an error is a player-domain rejection.
func Expected(err error) bool {
	return errors.Is(err, craftingrecord.ErrDisabled) || errors.Is(err, craftingrecord.ErrAltarNotFound) || errors.Is(err, craftingrecord.ErrRecipeNotFound) || errors.Is(err, craftingrecord.ErrRecipeSoldOut) || errors.Is(err, craftingrecord.ErrIngredients) || errors.Is(err, craftingrecord.ErrItemUnavailable)
}
