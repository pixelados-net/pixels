package recipe

import (
	"context"
	craftedevent "github.com/niflaot/pixels/internal/realm/crafting/recipe/events/crafted"
	discoveredevent "github.com/niflaot/pixels/internal/realm/crafting/recipe/events/discovered"
	exhaustedevent "github.com/niflaot/pixels/internal/realm/crafting/recipe/events/exhausted"
	craftingrecord "github.com/niflaot/pixels/internal/realm/crafting/record"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	"github.com/niflaot/pixels/pkg/bus"
	"sort"
)

func (service *Service) commit(ctx context.Context, playerID int64, recipe craftingrecord.Recipe, explicit []int64) (Result, error) {
	result := Result{Recipe: recipe}
	err := service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		inventory, err := service.furniture.ListInventory(txCtx, playerID)
		if err != nil {
			return err
		}
		selected, err := selectIngredients(inventory, recipe.Ingredients, explicit)
		if err != nil {
			return err
		}
		if recipe.Limited {
			remaining, consumed, consumeErr := service.store.ConsumeLimited(txCtx, recipe.ID)
			if consumeErr != nil {
				return consumeErr
			}
			if !consumed {
				return craftingrecord.ErrRecipeSoldOut
			}
			result.Exhausted = remaining == 0
		}
		for _, itemID := range selected {
			if err = service.furniture.DeleteInventoryItem(txCtx, itemID, playerID); err != nil {
				return craftingrecord.ErrItemUnavailable
			}
		}
		granted, err := service.granter.Grant(txCtx, furnitureservice.GrantParams{DefinitionID: recipe.RewardDefinitionID, OwnerPlayerID: playerID, Quantity: 1})
		if err != nil {
			return err
		}
		if len(granted) != 1 {
			return craftingrecord.ErrInvalid
		}
		result.Removed = selected
		result.Granted = granted[0]
		result.Definition, _, err = service.furniture.FindDefinitionByID(txCtx, recipe.RewardDefinitionID)
		if err != nil {
			return err
		}
		if recipe.Secret {
			result.Discovered, err = service.store.RememberRecipe(txCtx, playerID, recipe.ID)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return Result{}, err
	}
	service.afterCommit(playerID, recipe, result)
	return result, nil
}
func (service *Service) afterCommit(playerID int64, recipe craftingrecord.Recipe, result Result) {
	if service.events == nil {
		return
	}
	_ = service.events.Publish(context.Background(), bus.Event{Name: craftedevent.Name, Payload: craftedevent.Payload{PlayerID: playerID, RecipeID: recipe.ID, RewardDefinitionID: recipe.RewardDefinitionID}})
	if result.Discovered {
		_ = service.events.Publish(context.Background(), bus.Event{Name: discoveredevent.Name, Payload: discoveredevent.Payload{PlayerID: playerID, RecipeID: recipe.ID}})
	}
	if result.Exhausted {
		_ = service.events.Publish(context.Background(), bus.Event{Name: exhaustedevent.Name, Payload: exhaustedevent.Payload{RecipeID: recipe.ID}})
	}
}
func selectIngredients(inventory []furnituremodel.Item, ingredients []craftingrecord.Ingredient, explicit []int64) ([]int64, error) {
	available := make(map[int64][]int64, len(ingredients))
	for _, item := range inventory {
		if item.InInventory() && !item.MarketplaceReserved {
			available[item.DefinitionID] = append(available[item.DefinitionID], item.ID)
		}
	}
	for definitionID := range available {
		sort.Slice(available[definitionID], func(i, j int) bool { return available[definitionID][i] < available[definitionID][j] })
	}
	if len(explicit) > 0 {
		valid := make(map[int64]struct{}, len(explicit))
		for _, ids := range available {
			for _, id := range ids {
				valid[id] = struct{}{}
			}
		}
		for _, id := range explicit {
			if _, found := valid[id]; !found {
				return nil, craftingrecord.ErrItemUnavailable
			}
		}
		return append([]int64(nil), explicit...), nil
	}
	selected := make([]int64, 0)
	for _, ingredient := range ingredients {
		items := available[ingredient.DefinitionID]
		if len(items) < int(ingredient.Amount) {
			return nil, craftingrecord.ErrIngredients
		}
		selected = append(selected, items[:ingredient.Amount]...)
	}
	return selected, nil
}
