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
- ROOM/NITRO IMPROVEMENT: Expose the existing per-room roller speed to hotel
  users. Add localized owner/rights commands `:speed <value>` and
  `:setspeed <value>` that persist and immediately project values from `-1`
  (disabled) through `20`, while retaining the protected admin route. Add a
  Nitro room-settings control for the same value through a dedicated compatible
  integration because Nitro's native room-settings packet has no roller-speed
  field. Do not treat `PIXELS_ROLLER_HOOK_DELAY` as roller cadence and do not add
  a global speed override unless its precedence over per-room values is defined.
- MODERATION/NITRO IMPROVEMENT: Finish the Guardian chat-review integration
  before treating it as manually testable. Nitro React currently has no views or
  message-event listeners for review offers, acceptance, anonymized transcripts,
  voting status, votes, results, or detach behavior even though nitro-renderer
  exposes the corresponding composers and parsers. Reconcile packet `1922` field
  order between `GuideSessionOnDutyUpdateMessageComposer` and Pixels' guide-duty
  decoder, and correct packets `735`, `143`, `1829`, and `3276` to match Nitro's
  timeout, status-array, winning-vote, own-vote, and final-status wire shapes.
  Add exact wire tests on both sides and a complete UI flow. Acceptance must work
  with `PIXELS_GUARDIAN_COUNT=1` using Bob as reporter, Alice as the reported
  player, and Carol as the sole on-duty guardian; then repeat with the default
  count of three using Demo, Alice, and Carol as reviewers, Bob as reporter, and
  a fifth non-reviewer player as the reported target. Verify accept, reject,
  timeout, disconnect, anonymization, strict-majority sanction, mixed-result
  staff escalation, Redis ignored-offer exclusion, and durable ticket/vote audit.

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
