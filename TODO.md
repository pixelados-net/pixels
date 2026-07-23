# TODO

Two buckets. **Normal** is work that lives entirely in this repository: server logic, policy, wire fixes, config. **Nitro** is work that is blocked on, or lives inside, the `nitro-react`/`nitro-renderer` client: Pixels already sends or accepts the wire contract, but the client has no view, control, or correct behavior to use it. Nitro items should not be read as missing server work; re-implementing server behavior to work around a missing client view is the wrong fix.

## Normal

### Last protocol coverage gap

The reproducible cross-check against all 922 entries of
`legacy/pixel-protocol/docs/protocol/packet-catalog.md` gives **902/922**.
The only 20 packets without a package are Furniture Media: Trax,
Jukebox/Sound Machine, and YouTube display.

- c2s: `336`, `753`, `1325`, `1435`, `2069`, `2304`, `3005`, `3050`,
  `3189`, `3498`.
- s2c: `34`, `105`, `469`, `1112`, `1140`, `1381`, `1411`, `1554`,
  `1748`, `2602`.

No package debt remains outside that cluster. Everything else below is
product, server, or client work. None of it should be recounted as a
missing server packet.

### Server and policy

- Fix permissions of HC in room settings.
- Implement `gld_gate` as a real conditional group gate before enabling its
  catalog offer. Arcturus opens it automatically only for an avatar that
  belongs to the linked group (or has an explicit hotel override), then
  closes it after the avatar walks off. Pixels' current path resolver uses
  room-wide static fixtures, so seeding it as ordinary custom furniture would
  either block everyone or let everyone through. Add per-mover conditional
  traversal, open/close projection, reconnect state, override permission,
  WIRED behavior, and zero-allocation warmed membership checks; do not
  approximate it with the generic owner-only double-click gate.
- Reconcile the Guardian chat-review wire shape before treating the feature
  as manually testable. Reconcile packet `1922` field order between
  `GuideSessionOnDutyUpdateMessageComposer` and Pixels' guide-duty decoder,
  and correct packets `735`, `143`, `1829`, and `3276` to match Nitro's
  timeout, status-array, winning-vote, own-vote, and final-status wire
  shapes. Add exact wire tests on both sides.
- RETIRED PROTOCOL: MiniMail no longer exists as a Pixels feature. Keep
  outbound packets `1911` and `2803` only as documented deprecated wire
  stubs with golden tests. They must have no runtime call sites, database,
  unread-count query, cache, API, events, login projection, scheduler,
  seeds, or replacement UI.
- CMS OWNERSHIP: Email and verification belong to the CMS, not the
  emulator. Pixels keeps the historical email/welcome-email inbound packets
  as bounded, authenticated no-ops and the outbound wire packages only as
  compatibility documentation. Keep them free of database tables, provider
  calls, callbacks, APIs, login projection, and fake success responses.
  Revisit only if a future CMS integration defines an explicit server
  contract.
- REDESIGN USER EXPERIENCE: Replace the retired NUX/welcome-gift journey
  with a newly designed onboarding experience before adding product state
  or rewards. Until that design exists, inbound `1299`, `1822`, and `2638`
  are documented no-op handlers and outbound `3575`, `3639`, `2293`, and
  `2707` are isolated compatibility packet definitions with no call sites.
  Do not create onboarding migrations, seeds, APIs, reward grants, progress
  state, or legacy UI ahead of that design.

### Store and catalog policy

- Add real targeted-offer eligibility policies only after campaign
  requirements are defined. Current offers are global and use only
  enabled, expiration, dismissal, and per-player purchase-limit gates;
  future policies may compose club tier, permission group, account age, or
  purchase history.
- Add sell with multiple currencies.
- POLICY DEFERRED: Builders Club purchase-and-place is implemented
  atomically, but remains dormant while
  `PIXELS_BUILDERS_CLUB_FURNITURE_LIMIT=0`. Enabling it is a
  subscription-policy decision, not missing wire or backend behavior.
- DEFERRED: Direct SMS Club billing has no effect because Pixels has no
  carrier billing provider. The protocol request receives an explicit
  unavailable response.
- STORE INTEGRATION: Badge catalog products now have a safe durable badge
  recipient, but their catalog product shape and purchase presentation
  still need a focused audit before enabling an offer. Reuse player badge
  inventory; do not model a second badge owner or treat a raw product
  string as authority.
- PROTOCOL COMPATIBILITY: Achievements are real and no longer use an empty
  snapshot. Desktop promo articles, furniture aliases, and Community Goals
  stay explicit empty compatibility snapshots for unreachable client
  surfaces. Trax song-info compatibility remains neutral until the final
  Media domain is implemented; do not fabricate song rows. `VipHC` is
  progressed by the durable HC payday event.

## Nitro

Pixels already implements the durable backend and exact wire contracts for
everything below. In every case, the missing piece is a view, control, or
correct client behavior in `nitro-react`/`nitro-renderer`. Do not infer
completion, rewards, or state locally in the client, and do not build
server-side workarounds for a client gap.

### Progression

- Build the missing quest list, campaign, daily/seasonal offer, active
  tracker, cancellation, completion, and next-quest journey. Pixels owns
  the complete durable quest backend and exact wire contracts, but this
  Nitro React checkout has no quest view or link-event consumer. The UI
  must use the existing request/accept/cancel/tracker packets and must
  never infer completion, rewards, or campaign availability locally.
- Build the missing citizenship/helpers talent-track panel. Pixels derives
  and persists consecutive levels, grants items, badges, and perks exactly
  once, and projects the native talent packets. The UI should explain
  requirements and rewards and refresh from server state; `TRADE` and
  guide eligibility remain authoritative server gates.
- Build the missing safety-quiz panel using the existing question/result
  packets. The server owns question order, exact answer count, pass
  persistence, and the one-time `SafetyQuizGraduate` trigger. Do not
  expose correct answers in a new HTTP player endpoint or grant the result
  client-side.
- Add a visible promotional-badge claim entry point only after its product
  placement is decided. Pixels already supports availability windows,
  global caps, idempotent claims, status lookup, badge projection, and
  audited administrative claims; the client currently has no normal
  control that sends these composers.
- Build the missing "My Effects" inventory panel. Pixels already sends the
  full and incremental effect inventory and handles Nitro's activation,
  selection, and disable composers, but this Nitro React checkout has no
  component that consumes those events or sends those composers. The UI
  must render cached effect thumbnails, consume the Pixels external-text
  source, show permanent state, charge count, duration, and remaining
  active time, activate an owned charge with composer `2959`, select
  exactly one visible effect with composer `1752`, and disable it with
  effect id zero. Once this panel is usable, restore Arcturus-compatible
  catalog behavior: purchases add effects without activating or selecting
  them; effect-giver and effect-tile furniture may continue granting and
  selecting immediately. Make the admin API grant-only by default while
  retaining explicit `enable: true` support.
- Add a visible room-favorite control. Pixels now implements favorite
  persistence, initialization, add/remove (`3817`/`309`), room and
  visibility validation, the limit, incremental packet `2524`, complete
  list refresh, and live viewer state. The vendored Nitro React client has
  no reliable control that emits the action, so this is now UI-only
  compatibility work.

### Camera

- Design the public photo-gallery journey before allowing players to
  browse or purchase another player's publication. Pixels currently
  publishes durable, administrable gallery records and deliberately
  ignores the optional `photoId` accepted by the modern purchase composer,
  matching the real Arcturus own-capture purchase flow. A future UI needs
  explicit author rights, pricing, ownership-transfer, moderation,
  takedown, and duplicate-purchase rules; do not turn the administrative
  gallery endpoint into an implicit player marketplace.
- Design a player-facing way to reopen the current checkout after
  reconnecting. Pixels retains one active capture for the configured
  pending-retention window and safely removes it afterward, but the
  audited Nitro client has no event or view that asks the server for a
  previous capture. Do not invent a wire contract until the recovery UI is
  designed.
- Define and implement the visual meaning of `blockCameraFollow`. Pixels
  persists and echoes the native user setting, but neither the audited
  Nitro renderer nor Arcturus excludes another avatar from a
  client-rendered screenshot. Real enforcement needs a compatible
  room-state projection and renderer behavior for other users; do not
  expose private settings through an invented packet or treat the stored
  flag as security.
- Keep `UPDATE_ROOM_THUMBNAIL` (`2468`)/`THUMBNAIL_UPDATE_RESULT` (`1927`),
  `PHOTO_COMPETITION` (`3959`)/`COMPETITION_STATUS` (`133`), and
  `CAMERA_SNAPSHOT` (`463`) isolated as tested compatibility wire until a
  real client call site and payload semantics are captured. The active
  Nitro flow uses PNG render packets `3226` and `1982`; do not build a
  server-side image compositor or photo contest from packet names alone.

### Groups

- Make the favorite-group star in the user's profile a larger, accessible
  click target with a tooltip and stop its click from bubbling into
  group-card selection. Keep it visually and spatially distinct from the
  full-width Leave/Abandon button and from the non-interactive gray
  member-rank star rendered below the member count. Give that rank icon
  its own explanatory tooltip so it is not mistaken for the favorite
  action. Pixels already persists the single favorite, projects it to the
  room, and invalidates the open profile after a successful change; this
  task is only Nitro interaction clarity.
- Build the missing end-user group forum views. Pixels now implements
  forum entitlement, settings, policies, active/most-viewed/my group
  lists, paginated threads and posts, unread markers, pin/lock/hide,
  moderation, live updates, rate limits, and forum CFH packets. Nitro
  Renderer already exposes the composers/parsers, but this Nitro React
  checkout only has the catalog forum-product layout and no forum panel.
  Define the product journey before exposing it as complete: explain that
  purchasing the terminal enables a forum for the selected owned group,
  decide whether entry happens from the terminal, group profile, toolbar,
  or a dedicated forum browser, and make that entry point visible. Add the
  missing catalog page/item descriptions and translations so Nitro never
  shows raw `catalog.page.group_forums` or `catalog.item.group_forum_terminal`
  keys or an empty explanation panel. Decide what activating the placed
  terminal must do and show a clear already-enabled, non-owner, no-group,
  and purchase-success state. Until this journey is defined, do not treat
  the catalog purchase or rooms 130-135 as sufficient manual forum QA. The
  eventual UI must consume the existing renderer events, keep the opened
  thread/message cursor coherent, render localized policy failures, and
  expose report controls without treating client state as authorization.
- Add an owner-transfer administration flow only if a product decision
  makes it player-facing. Pixels intentionally exposes durable ownership
  transfer through the protected Admin Groups API because the audited
  Nitro client has no composer or view for it. Do not invent a packet or
  silently let an owner leave; transfer or group deactivation must
  preserve exactly one owner.
- Make the member roster expose only actions the current social role can
  perform. `GroupMembersView` currently shows the **Pending** filter to
  ordinary members and uses one `admin` flag both for accepting/removing
  members and for promoting/demoting them. Pixels correctly allows
  owner/admin to resolve requests and remove lower-ranked members, while
  role changes remain owner-only. Hide the pending filter from
  non-managers and split roster-management from role-management
  capability in the React view (or add a compatible owner signal) so a
  group admin is not offered a rank button that the server must reject.
  Harden the zero-result state as well: combining a retained search query
  with another filter can produce `totalPages === 0`, but the current view
  renders page `1 / 0` and leaves the next-page control enabled. Keep the
  member card mounted, show an explicit empty result, and disable both
  pagination controls without reloading or closing the hotel connection.
  Server authorization remains authoritative.
- Make group-profile navigation actions bring their destination to the
  front or close the group card. The current React actions emit
  `catalog/open/guild_custom_furni` and `navigator/search/groups`, but the
  catalog or Navigator can remain visually behind the group information
  window and look like the click did nothing. Preserve the server-owned
  canonical catalog page name and Navigator search semantics; this is
  window focus/stacking and user feedback, not authorization.
- Audit group badge preview/render consistency across creator, manager,
  group information, room information, profiles, Navigator, and linked
  furniture. Pixels persists the normalized parts and compiled badge code,
  while Nitro React selects the assets and scale. Keep intentional pixel
  art crisp, use the same layer order and colors after reconnect, and
  surface a clear preview if an asset combination clips or scales
  differently. Do not "fix" a rendering defect by rewriting the
  authoritative badge code.
- Support live group-badge replacement in the room renderer and React
  group views. `FurnitureGuildCustomizedVisualization` currently fills
  `_badgeAssetNameNormalScale` and `_badgeAssetNameSmallScale` only while
  the first value is empty, so an `OBJECTS_DATA_UPDATE` reaches the room
  object but keeps rendering the previous badge asset. Nitro React also
  registers `GROUP_DETAILS_CHANGED` without consuming it in a view.
  Invalidate both cached badge asset names when the group badge changes,
  refresh the room group widget and open group cards, then re-enable
  Pixels' incremental badge projection. Until then, badge edits become
  visible when the room objects are rebuilt on re-entry; color-only
  furniture projection remains live.

### Moderation

- Finish the Guardian chat-review integration's client side once the wire
  reconciliation above lands. Nitro React currently has no views or
  message-event listeners for review offers, acceptance, anonymized
  transcripts, voting status, votes, results, or detach behavior even
  though nitro-renderer exposes the corresponding composers and parsers.
  Add a complete UI flow and verify it against both guardian-pool sizes:
  `PIXELS_GUARDIAN_COUNT=1` with Bob as reporter, Alice as the reported
  player, and Carol as the sole on-duty guardian; then the default count
  of three with Demo, Alice, and Carol as reviewers, Bob as reporter, and
  a fifth non-reviewer player as the reported target. Verify accept,
  reject, timeout, disconnect, anonymization, strict-majority sanction,
  mixed-result staff escalation, Redis ignored-offer exclusion, and
  durable ticket/vote audit.

### Pets

- Build the missing breeding confirmation and result UI. Pixels implements
  the durable nest session and all Nitro breeding packets (`1746`, `634`,
  `1625`, `2621`, `2527`, and `1553`), and nitro-renderer emits their room
  events, but this Nitro React checkout has no complete dialog that lets
  both owners inspect the parents, confirm the offspring name, cancel, and
  see the rarity/result. Keep the server timeout and idempotency
  authoritative; the client must never infer completion merely because it
  closed a dialog.
- Wire the Monster Plant `Revive` action in
  `AvatarInfoWidgetOwnPetView`. The audited Nitro React branch renders the
  button for a dead owned plant, but its `case 'revive'` is empty and
  sends no composer, so the server cannot react. Until that UI is
  implemented, revival uses the native placed-product flow: activate a
  `MONSTERPLANT_REVIVAL` potion, select an eligible dead plant, confirm,
  and send `USE_PET_PRODUCT` (`1328`) with furniture item id plus pet id.
  Also show localized feedback when a placed potion has no eligible plant
  instead of silently producing no target bubbles; authorization and
  lifecycle validation remain server-side.
- Render the Monster Plant `Harvest` action when the server projects
  `canHarvest=true`. `AvatarInfoWidgetOwnPetView` already has a working
  `case 'harvest'` that calls `roomSession.harvestPet`, but no menu or
  infostand control invokes it, so manual QA currently has to inject
  `HARVEST_PET` (`1521`) with the plant id. Keep the server's maturity,
  ownership, compare-and-swap, reward, and idempotency checks
  authoritative.
- Render a dead Monster Plant as `rip` in its info and context-menu
  preview. `AvatarInfoUtilities.getPetInfo` currently maps every plant at
  or above `adultLevel` to `std` without checking `petData.dead`. Resolve
  `dead` to `rip` before the adult and growth-stage branches; this is a
  preview-only UI correction and must not mutate the authoritative room
  unit.

### Room

- Expose the existing per-room roller speed to hotel users. Add localized
  owner/rights commands `:speed <value>` and `:setspeed <value>` that
  persist and immediately project values from `-1` (disabled) through
  `20`, while retaining the protected admin route. Add a Nitro
  room-settings control for the same value through a dedicated compatible
  integration because Nitro's native room-settings packet has no
  roller-speed field. Do not treat `PIXELS_ROLLER_HOOK_DELAY` as roller
  cadence and do not add a global speed override unless its precedence
  over per-room values is defined.
- Add compact lives/ammo/power-up HUDs for Freeze and team/score overlays
  for Battle Banzai. The server-authoritative games are fully playable
  through furniture state, avatar effects, and rolling; these overlays
  must only render state and never become game authority.
- Nitro Renderer registers room-poll packets `5200`/`5201` and composer
  `6200`, but Nitro React has no view subscribing to those events. Add an
  accessible choice/result view before exposing infobus polls as a
  visible hotel feature; DB polls continue through the existing poll
  widget.
- GAME CENTER: the external URL launcher is complete. Add weekly
  leaderboard views only if an external game begins reporting real
  scores; do not display fabricated rows for the empty default table.

### Crafting

- Crafting, secret recipes, Ecotron, and credit exchange are connected in
  the active `../nitro-react`, but the secret and recycler selectors
  expose raw furniture instance IDs. Render furniture thumbnails, names,
  selected counts, and accessible state instead of `#<itemId>` buttons;
  keep selection bounded without duplicating inventory models.
- Add Nitro external texts for `crafting.title.secret`, `crafting.btn.check`,
  and `crafting.btn.secret`, and replace the inherited
  `Study Table Crafting Table` title with a clear localized Crafting
  title. Reset secret match/exact state whenever selection changes and
  after a craft, and remove consumed IDs from selection on live inventory
  updates.
