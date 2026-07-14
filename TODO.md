# TODO

Upon already coded things.

- currency/types.json must not be preserved, find another way to handle it via env
- Fix permissions of HC in room settings.
- NITRO IMPROVEMENT: Build the missing "My Effects" inventory panel. Pixels
  already sends the full and incremental effect inventory and handles Nitro's
  activation, selection, and disable composers, but this Nitro React checkout
  has no component that consumes those events or sends those composers. The UI
  must render cached effect thumbnails, consume the Pixels external-text source,
  show permanent state, charge count, duration, and remaining active time,
  activate an owned charge with composer `2959`, select exactly one visible
  effect with composer `1752`, and disable it with effect id zero. Once this
  panel is usable, restore Arcturus-compatible catalog behavior: purchases add
  effects without activating or selecting them; effect-giver and effect-tile
  furniture may continue granting and selecting immediately. Make the admin API
  grant-only by default while retaining explicit `enable: true` support.
- Complete room favorites. Persistence, initialization, packet definitions, and
  `navigator.favorite_changed` already exist, but commands, handlers, room
  validation, the favorite limit, packet `2524` projection, and handler
  registration are still missing for inbound add (`3817`) and remove (`309`).
  The vendored Nitro React client also has no room-favorite control, so backend
  completion alone will not expose this action in the current UI. Do not modify
  Nitro as part of the server implementation; treat client support as a separate
  compatibility decision.
- NITRO IMPROVEMENT: Hide or disable the inventory's "Place in room" action
  when the current player cannot manage furniture in the active room. Nitro
  currently shows the action whenever any room session exists and lets Pixels
  reject packet `1258` with the localized no-rights bubble. Keep the server-side
  `room.furniture.any.manage` and active room-rights check authoritative; any
  client-side gate must remain consistent with owner, local rights, and global
  furniture-management permissions rather than treating it as security.

## Store Boundaries

- TODO: Add real targeted-offer eligibility policies only after campaign
  requirements are defined. Current offers are global and use only enabled,
  expiration, dismissal, and per-player purchase-limit gates; future policies
  may compose club tier, permission group, account age, or purchase history.
- TODO: Add sell with multiple currencies.

- UNIMPLEMENTED: Builders Club has no gameplay effect. It is a discontinued
  subscription tier without a real Arcturus implementation; Pixels only sends
  neutral compatibility packets so Nitro never waits indefinitely.
- DEFERRED: Direct SMS Club billing has no effect because Pixels has no
  carrier billing provider. The protocol request receives an explicit
  unavailable response.
- DEFERRED: Room bundles have no effect because complete room cloning,
  bundled room layouts, and bot ownership do not exist yet.
- UNIMPLEMENTED: Marketplace and direct player trade have no effect in the
  store realm because they belong to the separate `economy-trading` protocol
  surface and require escrow, commissions, and offer lifecycle rules.
- UNIMPLEMENTED: Crafting and recycling have no effect because they belong to
  the separate `crafting-recycling` realm and need recipes and material state.
- UNIMPLEMENTED: Group catalog products have no effect because Pixels has no
  group realm capable of owning or granting them.
- UNIMPLEMENTED: Pet, bot, and badge catalog products have no effect because
  those recipient realms do not exist; accepting the products would charge a
  player without being able to grant the result.
- UNIMPLEMENTED: The `VipHC` achievement has no effect because Pixels has no
  achievement realm to receive accumulated membership progress.
