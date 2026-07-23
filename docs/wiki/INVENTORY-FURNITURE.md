# The Furniture Inventory

First page of the Inventory section. "Inventory" in Pixels is several parallel systems: furniture, money, badges, pets, bots, effects, clothing. This page covers the biggest: how owned furniture is stored, listed, and kept in sync with the client. [[INVENTORY-WALLET]] covers money; [[INVENTORY-COLLECTIONS]] covers everything else.

## One table, no drift

Owned furniture lives in a single durable table of items, where "in inventory" simply means the item has no room placement: `RoomID` is null. There is no separate inventory table that could drift out of sync with placements: picking an item up and placing it down are updates to the same row.

Each item row carries:

- the **definition** reference, what kind of furniture this is (see [[FURNITURE-MODEL]]),
- **owner**, optional **room and coordinates**, rotation or wall position,
- **`ExtraData`**, the client-visible state string: a dice face, a post-it's color and text, a photo's JSON,
- the **limited edition serial** when the item is an LTD,
- **gift wrapping** state (wrapper sprite, ribbon, sender, message) for unopened presents,
- a **`MarketplaceReserved`** flag that pulls an item into marketplace limbo while listed, so it can't be simultaneously traded, placed, or sold twice,
- server-only **`Metadata`** for structured data the client never sees.

That last distinction, `ExtraData` is wire-visible state and `Metadata` is server-only, repeats across the codebase and is worth internalizing early.

## Fragmented listing

The client reads inventory through a fragmented list protocol: the server slices the full item list into fixed-size fragments and the client requests them page by page. A player hoarding five thousand items produces a sequence of bounded packets instead of one megapacket that stalls the connection. Fragment size is a config value, and the same pattern is reused by the pet inventory.

## Incremental updates

Mutations never resend the list. Committed changes push the minimal delta: remove packets for consumed items, an add packet for a granted one, then a refresh marker, using the shared projection helper described in [[PROJECTIONS]]:

```go
// Inventory sends committed removals, one addition, and a refresh marker.
func Inventory(ctx context.Context, connection netconn.Context, removed []int64,
	granted furnituremodel.Item, definition furnituremodel.Definition) error
```

Everything that grants or consumes furniture routes through this: crafting results, recycler prizes, catalog purchases, camera photo purchases, trade settlements, marketplace deliveries. The client's inventory window updates in place, and a full relist only happens when the client explicitly asks (opening the window fresh, or the away-from-room inventory request that aliases the same read).

## Unseen tracking

Newly acquired items are flagged **unseen** per category, which is what renders the "new" indicator on the client's inventory tabs. Dedicated reset packets clear the state per item or per category once viewed, and the same flow serves edge cases like group furniture being returned to its owner when removed from a group room: the returned items arrive flagged unseen so the owner notices them.

## Where the code lives

The furniture inventory has no realm of its own; it's a capability of the furniture realm. The read/list/unseen handlers sit in `internal/realm/furniture/handlers/inventory`, grants live in the furniture service (`Grant`, `GrantGift`), and the deltas are pushed by whichever realm committed the change, through the shared projection helpers. If you're implementing a feature that awards an item, you want `furnitureservice.Grant` inside your transaction and the projection helper after commit. That's the pattern every existing reward flow already follows.
