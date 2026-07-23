# Badges, Pets, Bots, Effects, and Clothing

Third page of the Inventory section: everything a player owns besides furniture and money. Each of these has its own owning realm. This page is the map, with pointers to where each is covered in depth.

## Badges

Badge inventory lives in the player realm and is covered in full in [[USERS-BADGES-PERMISSIONS]]: owned codes, five equip slots, idempotent grants, and in-place upgrades when an achievement badge levels up. The short version for this map: a badge is either owned-and-unequipped or owned-and-in-a-slot, never anything more complicated, and the client's badge tab is fed by the same list/current/equip packet trio regardless of whether the badge came from an achievement, a promotion, or an admin grant.

## Pets

Pets have their own inventory service inside the pet realm, following the same nullable-placement model as furniture: a pet is either sitting in inventory or placed in a room, and picking it up or placing it is one field changing on one row, not a move between tables. Pet lists are fragmented the same way furniture is (see [[INVENTORY-FURNITURE]]), and placement enforces both a per-player cap and a per-room cap so a room can't be flooded with pets. Pet-specific state, stats, breed, training progress, lives alongside the inventory row but is owned by the pet realm's own services, not by the generic inventory machinery.

## Bots

Bots follow the identical shape in the bot realm: an inventory list, a place command, a pick-up command. A bot's behavior configuration (what it says, how it moves) is edited once placed and travels with the bot between inventory and room.

## Avatar effects

Durable per-player effect inventories live in the player realm (`player/effect`, detailed in [[USERS-PROFILE]]): which effects a player owns, with quantities or durations, and which one is active. This is distinct from the *transient* effects the room engine applies at runtime, such as a game's team color or riding a horse, which are live unit state, not inventory rows. If you're wondering why an effect a player is visibly using doesn't show up in their effects tab, that's usually the distinction: it's a runtime effect, not an owned one.

## Clothing

Two related things share the word "clothing," and it's worth keeping them apart:

- The **wardrobe** is saved outfits: numbered slots holding a full figure string each, part of the player realm, covered in [[USERS-PROFILE]].
- **Redeemable clothing** is a furniture item that, when redeemed, permanently unlocks wearable parts for that player. Redemption consumes the furniture item and records the unlocked parts against the player; the figure validator described in [[USERS-PROFILE]] then accepts those parts in future look changes. Once redeemed, the unlock has nothing left to do with the furniture item that granted it. You can't "un-redeem" by re-acquiring the item.

## The three invariants

Every collection on this page and the two before it share the same three rules, and they're worth stating once explicitly because they explain *why* the code looks the way it does everywhere you find it:

1. **Durable state is the only truth.** Every live projection, the client's inventory window, the badge equip snapshot, the currency holder, can be rebuilt from the database at any moment. Reconnecting always heals a stale view; nothing is ever permanently lost by a dropped connection.
2. **Mutations are incremental.** Committed changes push adds, removes, and refresh markers. Full relists happen only when the client explicitly asks, never as a side effect of some unrelated change.
3. **Cross-collection operations are one transaction.** Trading furniture for credits, buying a pet, redeeming clothing, anything touching two of these systems commits or fails as a unit, through the scoped-transaction mechanism in [[INFRASTRUCTURE]]. There is no "it debited but didn't deliver" state anywhere in the codebase, by construction rather than by careful bookkeeping.
