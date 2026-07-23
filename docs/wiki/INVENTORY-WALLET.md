# The Currency Wallet

Second page of the Inventory section. Money is its own small realm, `internal/realm/inventory/currency`, and it's the one every other paid feature in the codebase routes through. This page covers how currency types are modeled, how balances mutate atomically, and how the client stays in sync.

## Currency types are data

There's no hardcoded enum of "credits, duckets, diamonds" scattered through the codebase. Currency types are catalog rows:

```go
// Definition describes one configured currency type.
type Definition struct {
	Type   int32  // protocol id: -1 credits, 0 duckets, 5 diamonds…
	Key    string // stable name used in config and i18n
	Ledger bool   // whether mutations require audit ledger entries
	Color  string
}
```

The development catalog enables three: credits (`-1`), duckets (`0`), and diamonds (`5`). Those exact protocol ids matter because the client renders currency icons by number, not by name. Adding a seasonal currency is a catalog entry and client text, not a schema change or a redeploy.

## One mutation surface

Every balance change in the entire codebase (catalog purchases, crafting's credit-furniture exchange, camera photo purchases, room ad promotions, achievement rewards, admin corrections) goes through the same narrow service contract:

```go
// Granter applies signed player currency mutations.
type Granter interface {
	Grant(ctx context.Context, params GrantParams) (int64, error)
}

type GrantParams struct {
	PlayerID     int64
	CurrencyType int32
	Amount       int64  // signed: positive credits, negative debits
	Reason       string
	ActorKind    string // "player", "admin", or "system"
	ActorID      *int64
}
```

A charge is just a `Grant` with a negative amount and a reason. There's no separate "debit" method to keep in sync with "credit", one code path, one place to get atomicity right. Currency types marked `Ledger` additionally write an append-only audit entry per mutation, which is what lets an operator answer "where did this player's balance actually come from" for the currencies that need that accountability (typically the premium ones), without paying the storage cost for every type.

## Why nothing is cached

The live player carries a deliberately thin `Holder`. It knows the player's ID and reads balances through the service on demand. There is no cached balance sitting on the live player that could drift from the database:

```go
// Wallet reads the current durable wallet through a narrow reader contract.
func (holder *Holder) Wallet(ctx context.Context, currencies currencyservice.Reader) ([]currencymodel.Balance, error) {
	return currencies.Wallet(ctx, holder.playerID)
}
```

This matters because balances change from many realms concurrently, and a cached copy would need invalidation logic duplicated everywhere money moves. Reading through the service instead means there's exactly one source of truth, always.

## Atomic charge-and-deliver

The pattern that makes "pay for it and get it, or neither" possible is the scoped transaction described in [[INFRASTRUCTURE]]: a purchase opens one transaction, and both the currency debit and the item grant join it through their respective repositories' `WithinTransaction`. Crafting's item exchange is a compact real example:

```go
err := service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
	// … validate ownership, delete the redeemed item …
	result.Credits = int64(definition.RedeemableCredits)
	_, err = service.currencies.Grant(txCtx, currencyservice.GrantParams{
		PlayerID: playerID, CurrencyType: creditsType, Amount: result.Credits, Reason: "item_exchange",
	})
	return err
})
```

If the grant fails, the item deletion rolls back too. There is no state where a player loses an item and doesn't get paid, or gets paid twice for one item.

## Keeping the client in sync

Balances are sent at login (part of the authentication bootstrap in [[AUTH-SSO]]) and on the client's explicit wallet request packet. A single balance change pushes a lightweight activity-point notification so the purse animates immediately instead of waiting for a full wallet refresh.

## Administration

`/api/admin/currencies/{wallet,grant,deduct,set,types}` exposes the same service to staff, with the standard actor-and-reason audit trail. Grants can optionally trigger a localized in-client alert (`alert: true`). The response distinguishes `alertRequested` from `alertSent` so a grant to an offline player still reports success without falsely claiming the player was notified.
