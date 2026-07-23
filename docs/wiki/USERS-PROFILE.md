# Identity, Profile, and Player State

Second page of the Users section. [[USERS-MODEL]] explained the durable/live split; this page walks the durable state families that make up a player, each with its own subpackage, repository, and packets under `internal/realm/player`.

## Identity

The account core: ID, canonical username, creation date. Usernames resolve case-insensitively through a canonical form, which is what the admin lookup route (`GET /api/admin/players/by-username/{username}`) and every social feature use, so `Demo`, `demo` and `DEMO` are one person. Identity is deliberately minimal; everything expressive lives in the profile.

## Profile

The social surface: motto, current figure string and gender, achievement score, and the player-to-player **respect** mechanic. Respect is quota-based, a daily allowance of respects you can give, tracked alongside how many you've received, and giving one is idempotent per interaction, feeding both the profile counter and the progression realm's respect achievements. Motto changes pass the same text filtering as chat, and a failed motto save is localized feedback to the client, never a disconnect.

## Figure validation

`player/figure` parses and validates avatar figure strings against the hotel's figure data catalog (the file behind `PIXELS_FIGURE_DATA_PATH`). A look change is validated part by part server-side: set types, part IDs, color indexes, and gender compatibility all have to check out against the catalog before the look is accepted and broadcast. Clients don't get to invent clothing they don't own; parts gated behind club membership or clothing redemption (see [[INVENTORY-COLLECTIONS]]) are checked against the player's actual entitlements.

The parser is also careful about *partial* acceptance: an invalid part rejects the change wholesale rather than silently stripping pieces, so what you see in the client is always exactly what the server stored.

## Wardrobe

Saved outfits: numbered slots holding a figure string and gender each, so players can switch looks without rebuilding them. Saving validates through the same figure pipeline above. The wardrobe also participates in the redeemable-clothing flow: once a clothing furniture item is redeemed, its parts become permanently valid for this player's future looks.

## Avatar effects

`player/effect` owns durable effect inventories: which avatar effects a player has, with quantities or durations, and which one is currently activated. Effects arrive from catalog purchases and furniture rewards, and activation projects to the room so everyone sees the effect start. The room engine itself also applies *transient* effects (team colors in games, riding a horse); those are runtime state on the unit, not entries in this durable inventory. The split matters when you wonder why a game effect doesn't show up in the effects tab.

## Client settings

`player/settings` persists client preferences: the three volume sliders, old-chat rendering, camera-follow blocking, and the home room. Two implementation details worth knowing:

- Writes go through a coalescing `Writer`, so a slider being dragged becomes one row update, not fifty. The live snapshot updates immediately; the durable write lands on the coalescing interval or at disconnect.
- Home room selection validates visibility server-side. You can't set an invisible room you have no rights to as home, and the check reuses the room realm's own rights service rather than duplicating the rule.

## Where the packets live

Each family registers its own handlers following the [[HANDLERS-AND-COMMANDS]] conventions: settings mutations, look and motto changes, wardrobe slots, effect activation. They all resolve the actor through the binding registry, write durable-first, then update the live snapshot: the exact shape shown in [[USERS-MODEL]]. If you're looking for any specific user packet, start at `internal/realm/player/<family>/handler.go` and you'll find the registration list at the bottom of the file.
