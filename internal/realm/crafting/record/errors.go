package record

import "errors"

var (
	// ErrDisabled reports globally disabled crafting behavior.
	ErrDisabled = errors.New("crafting is disabled")
	// ErrAltarNotFound reports a missing or inactive altar.
	ErrAltarNotFound = errors.New("crafting altar not found")
	// ErrRecipeNotFound reports a missing or invisible recipe.
	ErrRecipeNotFound = errors.New("crafting recipe not found")
	// ErrRecipeSoldOut reports exhausted limited stock.
	ErrRecipeSoldOut = errors.New("crafting recipe sold out")
	// ErrIngredients reports an invalid or insufficient ingredient bag.
	ErrIngredients = errors.New("invalid crafting ingredients")
	// ErrItemUnavailable reports an invalid, foreign, placed, or reserved item.
	ErrItemUnavailable = errors.New("crafting item unavailable")
	// ErrRecyclerClosed reports disabled recycler behavior.
	ErrRecyclerClosed = errors.New("recycler is closed")
	// ErrRecyclerBatch reports a malformed recycler batch.
	ErrRecyclerBatch = errors.New("invalid recycler batch")
	// ErrRecyclerPrize reports an invalid recycler prize configuration.
	ErrRecyclerPrize = errors.New("recycler prize unavailable")
	// ErrExchangeValue reports furniture without redeemable credits.
	ErrExchangeValue = errors.New("furniture has no exchange value")
	// ErrConflict reports an optimistic administration conflict.
	ErrConflict = errors.New("crafting state changed concurrently")
	// ErrInvalid reports malformed crafting administration input.
	ErrInvalid = errors.New("invalid crafting input")
)
