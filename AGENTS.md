# AGENTS.md

This repository contains Pixels, a fast and idiomatic Go emulator for the pixel protocol. Agents working here should keep the codebase simple, reusable, and easy to extend without hiding behavior behind premature abstractions.

## Project Layout

- `pkg/` contains reusable global components such as storage, Redis, WebSocket helpers, logging, and process utilities.
- `internal/` contains pixel-protocol realm features that are private to this emulator.
- `networking/` contains pixel-protocol packet coding, decoding, framing, and transport logic.
- `sdk/` contains controlled reusable implementations for plugin creation and extension points.
- `cmd/` contains executable entry points.

## Networking Layout

- `networking/codec/` contains two-way wire encoding and decoding helpers.
- `networking/connection/` contains transport-agnostic connection sessions, registries, handler routing, state, security policy, and disconnect reasons.
- `networking/crypto/` contains cryptographic contracts and implementations such as Diffie-Hellman.
- `networking/inbound/` contains client-to-server packet definitions.
- `networking/outbound/` contains server-to-client packet definitions.
- Connection sessions may hold transport and security callbacks, but packet handlers must stay transport-agnostic and security-agnostic.
- Connection sessions must unwrap security before dispatching packets, so handlers never know whether a packet came from encrypted or plain traffic.
- Realm packages own their handlers and commands; connection handler registries only register handlers and route decoded plain packets to them.
- Inbound packet packages decode only; expose `Decode(packet codec.Packet) (Payload, error)` and do not expose packet constructors.
- Outbound packet packages encode only; expose `Encode(...) (codec.Packet, error)` and do not expose packet decoders or public payload structs.
- Outbound required protocol fields must be function parameters, while optional protocol fields must use packet-local option functions such as `WithReason(value)`.
- Inbound decoders must validate the packet header before decoding the payload.
- Packet definitions must not create `c2s` or `s2c` package directories.
- Keep exactly one packet per packet package, with `packet.go` and `packet_test.go`.
- Prefer direct functional package names such as `networking/outbound/session/hotel/availability`.
- Avoid one-word ladder paths for packet names, such as `closed/and/opens/at`.
- When packets share a functional concept, classify them by inbound or outbound and use a concise package name.

## Package Rules

- Prefer small, single-name packages with nested paths over long package names.
- Use `networking/session/ping/packet.go` and `networking/session/ping/packet_test.go` instead of names like `networking/session/pingpacket.go`.
- Keep exactly one realm event per event package, similar to packet packages.
- Place realm events under concise paths such as `internal/realm/player/events/disconnected/event.go`.
- Event packages expose `Name` and, when needed, `Payload`; do not group several event names or payloads in one realm-level `event.go`.
- Prefer deriving shared event behavior through small helpers at call sites over creating broad event registry packages.
- Group packet handlers by stable realm capability, such as `handlers/entry`, `handlers/moderation`, `handlers/rights`, or `handlers/settings`; do not create a package for a handler that only decodes one packet and dispatches one command.
- Keep each grouped handler in its own focused file with action-qualified constructors and registration functions, such as `NewBan` and `RegisterBan`.
- Group small related commands by stable capability when their individual packages would contain only one stateless command file; keep action-qualified command and handler types in separate files.
- Organize large realms by stable capability before technical role. Room code belongs under `access`, `control`, `record`, `runtime`, `world`, or `database`; place that capability's commands, handlers, and events inside the same subtree instead of restoring realm-wide `commands`, `handlers`, or `events` packages.
- Domain capabilities define the persistence contracts they consume. Concrete PostgreSQL repositories belong under the realm's `database` subtree and must not be imported by domain services.
- Keep substantial command workflows in dedicated packages when they own multiple files, private coordination, or focused tests.
- Keep each package focused on one responsibility.
- Keep each file focused on one responsibility.
- Keep every Go source file at or below 250 lines.
- Markdown planning and documentation files are exempt from the 250-line limit, but should still stay structured and easy to scan.
- Keep each package to a maximum of six file pairs, where `hello.go` plus `hello_test.go` counts as one pair.
- If a package needs more tests after six file pairs, create a `tests/` folder inside that package.

## Go Style

- Write Go the Go way: clear data flow, small functions, explicit errors, goroutines where concurrency is natural, and channels only where they clarify ownership.
- Document every package, function, method, struct, interface, type, const, var, field, private field, and test helper in Go doc style.
- Do not add comments inside function bodies.
- Avoid unnecessary interfaces. Introduce an interface only when it decouples a real boundary, supports multiple implementations, or enables focused tests.
- Let Go infer generic type arguments when the compiler can infer them clearly.
- Prefer `&Type{}` over `new(Type)` for pointer allocation so initialization stays explicit and consistent.
- Prefer composition over inheritance-like hierarchies.
- Do not create god functions that register, configure, validate, and execute unrelated responsibilities; split orchestration by behavior before one function becomes a dumping ground.
- Keep public APIs conservative and stable.
- Keep private APIs readable enough that new contributors can follow them quickly.

## Configuration

- Keep configuration holders close to the component they configure.
- Give each configuration holder its own `config.go`, such as `redis/config.go`, `storage/config.go`, or `logger/config.go`.
- Keep protocol host and port in `pkg/config/app/config.go` unless a concrete transport boundary needs its own settings.
- Do not create long registry-style configuration structs with every setting in the project.
- Compose small configuration structs into application-level configuration only where dependency injection or startup needs a single value.
- Every configuration field must have a default value.
- Keep `.env.example` documented whenever configuration variables are added, removed, or renamed.
- Load local `.env` files for development, but keep environment variables as the source of truth.

## HTTP Routes

- Document every HTTP route in `pkg/http/openapi`.
- Group administrative routes by capability, such as `/api/admin/currencies` or
  `/api/admin/notifications`, instead of accumulating unrelated behavior below
  `/api/admin/players`; pass the target player id through the request body or a
  query parameter as appropriate.
- Include request headers, request bodies, responses, possible error codes, and response bodies.
- Keep `/status`, `/ws`, and development-only documentation routes public.
- Protect private routes with the configured API key header.
- Return meaningful HTTP errors instead of generic status failures, and make logs include enough context to explain why a request failed.

## End-User Communication

- Localize every hotel-facing message before sending it to users, including bubbles, alerts, cautions, room entry errors with text, and future notification packets.
- Use `pkg/i18n` for server-side end-user text; packets must serialize already-resolved text and must not own translation logic.
- Keep technical logs, internal error values, and protocol disconnect diagnostics outside i18n unless they are shown to users.
- Prefer stable namespaced translation keys such as `session.bubble.furniture.no_rights` over raw text or short unscoped keys.

## Testing

- All code must maintain more than 80% test coverage.
- Add tests with every behavioral change.
- Keep tests focused on behavior, not implementation details.
- Prefer table-driven tests when cases share the same setup.
- CI must compile and test the full module before changes are considered ready.

## Implemented Feature Index

Use this section as a searchable map of major implemented behavior and the
minimum manual checks expected when touching it.

### FEATURE: Startup, Config, Logging

- Owns `cmd/`, `pkg/config`, `pkg/logger`, `pkg/build`, and Fx wiring.
- Loads `.env`, environment defaults, app host/port/access key, logger level,
  logger format, build version, and commit hash.
- Test after changes:
  - `go test ./pkg/config/... ./pkg/logger/... ./pkg/build/...`
  - `go run cmd/main.go` logs environment, host, port, and version.

### FEATURE: HTTP API, OpenAPI, Admin Auth

- Owns `pkg/http` and `pkg/http/openapi`.
- Provides `/status`, `/ws`, development-only `/docs`, SSO ticket creation,
  connection admin routes, localized player notifications, room admin routes,
  and navigator admin routes.
- Private routes require `X-API-Key`; `/status`, `/ws`, and `/docs` stay public.
- Test after changes:
  - `go test ./pkg/http/...`
  - Open `/docs` in development and verify route groups are visible.
  - Call private routes with and without `X-API-Key`.
  - Send `POST /api/admin/notifications/send` to an online player and
    verify the localized bubble or alert packet arrives.

### FEATURE: Redis SSO

- Owns `internal/auth/sso` and Redis-backed ticket storage.
- Creates one-time SSO tickets with TTL, optional IP binding, and player id.
- Consuming a ticket deletes it, so the same ticket cannot authenticate twice.
- Test after changes:
  - `go test ./internal/auth/sso/...`
  - Create a ticket through `POST /api/sso/tickets`.
  - Login once with that ticket and verify a second login fails.

### FEATURE: Postgres, Liquibase, Seeds

- Owns `pkg/postgres` and realm `database/` folders.
- Provides Postgres pool config, per-realm migrations, and per-environment seeds.
- Test after changes:
  - Run migrations against the `.env` database.
  - Verify seeded demo players, room layouts, rooms, categories, navigator data.
  - `go test ./pkg/postgres/... ./internal/realm/...`

### FEATURE: Event Bus and Commands

- Owns `pkg/bus` and `internal/command`.
- Provides prioritized local events and typed command dispatch with debug logs.
- Realm packet handlers decode packets and dispatch commands; realm commands own
  behavior.
- Test after changes:
  - `go test ./pkg/bus ./internal/command`
  - Run with debug logs and verify `event published` and `command dispatched`.

### FEATURE: Permission Groups and Player Overrides

- Owns `internal/permission`, permission database changelogs, permission protocol
  packets, and `pkg/http/permission/routes`.
- Provides typed registered nodes, wildcard grants, inheritable weighted groups,
  direct player allow/deny overrides, local plus Redis caching, live projection,
  default `member` assignment, and seeded `admin` and `moderator` access.
- Permission resolution order is direct player override, highest-weight matching
  group, then specificity and nearest child within that group's inheritance chain.
- Catalog page access uses optional `required_node`; player-originated currency
  deductions honor `currency.economy.infinite` without bypassing admin mutations.
- Test after changes:
  - `go test ./internal/permission/... ./pkg/http/permission/...`
  - `go test ./networking/outbound/session/permissions ./networking/outbound/session/perks`
  - `go test ./internal/permission/... -run '^$' -bench . -benchmem`
  - Keep warmed local cache hits and normal permission resolution at zero
    allocations; explain and benchmark any regression before merging it.
  - Open `/docs` and verify the `Admin Permissions` route group.
  - Assign and revoke a group/direct node while a player is online and verify
    `USER_PERMISSIONS` and `USER_PERKS` are projected immediately.
  - Create a player and verify membership in the seeded `member` group.
  - Verify demo resolves `admin`, Alice resolves `moderator`, and Bob and Carol
    resolve `member` as their highest-weight groups.

### FEATURE: Pixel Codec and Packets

- Owns `networking/codec`, `networking/inbound`, and `networking/outbound`.
- Encodes and decodes declarative packet definitions for handshake, security,
  session, currencies, navigator, and room packets.
- Test after changes:
  - `go test ./networking/...`
  - Run packet benchmarks after changing hot path packets.
  - Verify inbound decoders reject unexpected headers.

### FEATURE: Connection Sessions and WebSocket Transport

- Owns `networking/connection` and `pkg/http/websocket`.
- Provides transport-agnostic sessions, state machine, security policy, handler
  registries, disconnect reasons, heartbeat, ordered WebSocket writes, and
  debug packet logging.
- Test after changes:
  - `go test ./networking/connection ./pkg/http/websocket/...`
  - Connect Nitro to `/ws`, authenticate, and verify packet receive/send logs.
  - Verify admin connection count/list/disconnect routes.

### FEATURE: Player Realm

- Owns `internal/realm/player`.
- Provides persistent player/profile models, repositories, services, live player
  registry, session peer, embedded navigator viewer, and current room presence.
- Test after changes:
  - `go test ./internal/realm/player/...`
  - Authenticate with a seeded SSO ticket and verify user info bootstrap.
  - Enter and leave a room and verify live player room presence updates.

### FEATURE: Club Entitlements

- Owns player club fields, runtime entitlement projection, and HC gates.
- Club level and expiration are loaded once with the player, projected through
  Nitro permissions, and reused by catalog and room settings without hot-path
  database reads.
- HC controls only wall visibility and wall/floor thickness; chat remains a
  normal room setting. Global settings managers may perform administrative
  overrides.
- Test after changes:
  - `go test ./internal/realm/player/... ./internal/permission/broadcast/...`
  - Login as seeded `demo` and verify Nitro receives club level `2`.
  - Login as `bob` or `carol` and verify changing HC room fields is rejected.
  - Verify club-only catalog pages and offers follow active expiration.

### FEATURE: Catalog Core

- Owns `internal/realm/catalog` plus catalog-backed furniture grants.
- Provides catalog page/offer/LTD persistence, immutable cache generations,
  rank and club visibility gates, atomic purchases, and orphan-definition
  sanitation. Protocol packets and HTTP administration begin in K4 and K5.
- Catalog purchases share one PostgreSQL transaction across catalog, currency,
  and furniture repositories; currency events run only after commit.
- Test after changes:
  - `go test -race ./internal/realm/catalog/...`
  - `go test ./internal/realm/furniture/... ./internal/realm/inventory/...`
  - Run Liquibase `validate` after changing catalog migrations.

### FEATURE: Inventory Currencies

- Owns `internal/realm/inventory/currency`, currency packets under
  `networking/*/user/currency`, and `pkg/http/clientconfig`.
- Provides a catalog-driven wallet, composable player `CurrencyHolder`, atomic
  PostgreSQL mutations, optional per-type audit ledger, wallet authentication
  bootstrap, packet `273` handling, real-time player-only balance projection,
  protected admin mutation routes, and public Nitro config/text resources.
- Durable balances stay in the currency service; `Player` composes the holder
  capability without caching or owning currency rules.
- Admin currency alerts are opt-in (`alert` defaults to `false`), must be
  localized through `pkg/i18n`, and must not make an already-committed balance
  mutation fail when the target player is offline.
- Test after changes:
  - `go test ./internal/realm/inventory/... ./networking/... ./pkg/http/clientconfig ./pkg/http/currency/routes`
  - Apply Liquibase and verify `player_currencies` and
    `currency_ledger_entries` exist.
  - Authenticate in Nitro and verify packets `3475` and `2018` are sent.
  - Open `/client/ui-config.json` and
    `/client/texts/es/ExternalTexts.json`.
  - Grant and deduct currency through the admin API with `alert` omitted,
    `false`, and `true`; verify only the explicit `true` case sends a localized
    generic alert.

### FEATURE: Navigator Realm

- Owns `internal/realm/navigator`.
- Provides navigator persistence, embedded viewer state, init/search/create/info
  handlers, room forwarding, favorites data, saved searches, preferences, lifted
  rooms, category preferences, and debounced live category counts. Room creation
  validates selectable categories and enabled server layouts, enforces the
  current 100-room ownership limit, and converts expected validation failures
  into localized client alerts without disconnecting the session.
- Development seeds provide every Nitro room-creator layout from `model_a`
  through `model_9`, using the real Arcturus heightmaps and doors. All layouts
  remain non-HC until subscription policy is implemented. The public
  `/client/ui-config.json` projects enabled layouts back to Nitro using model
  suffixes, tile sizes, and zero club level without requiring client changes.
- Test after changes:
  - `go test ./internal/realm/navigator/...`
  - In Nitro, open navigator and verify metadata tabs, flat categories, settings,
    saved searches, favorites, lifted rooms, and collapsed categories.
  - Search hotel/myworld/official views and verify room cards show live counts.
  - Create a room and verify `navigator.room_created`.
  - Create rooms with small, elevated, and multi-level models; verify their
    heightmaps and doors render, and invalid models produce a localized alert.
  - Reach the room ownership limit and verify Nitro receives its native limit
    response instead of creating another room.
  - Request room info and verify missing rooms return `navigator.nosuchflat`.

### FEATURE: Room Realm

- Owns `internal/realm/room`.
- Uses six top-level capabilities: `access` for admission and doorbells,
  `control` for rights, moderation, settings, audit, and filtering, `record` for
  persistent room metadata, `runtime` for live occupancy and projections,
  `world` for layouts, tiles, furniture surfaces, units, and movement, and
  `database` for PostgreSQL implementations plus migrations and seeds.
- Commands, handlers, and events live under their owning capability. Do not add
  realm-wide `room/commands`, `room/handlers`, or `room/events` trees.
- Provides room layouts, categories, tags, persistent room metadata, runtime room
  registry, occupancy events, entry commands, model/heightmap packets, and tag
  packets.
- Initial room entry sends the model name once; subsequent room-model requests
  send only door and heightmap geometry so Nitro cannot enter a request loop.
- Path cancellation caused by furniture or fixture changes must broadcast a
  neutral final unit status without `mv`; silent cancellation leaves clients
  animating movement indefinitely.
- Test after changes:
  - `go test ./internal/realm/room/...`
  - Click a room from navigator, enter it, and verify empty room model renders.
  - Enter the same room repeatedly and verify packet `2300` does not loop.
  - Change furniture during a walk and verify the unit stops on its current tile.
  - Fill a runtime room to capacity and verify `room.entry_error`.
  - Verify `room.occupancy_changed`, `room.entered`, and `room.left` events.

### FEATURE: Room Chat

- Owns `internal/realm/chat`, `networking/inbound/chat`,
  `networking/outbound/chat`, and `pkg/http/chat/routes`.
- Provides Nitro talk, shout, whisper, typing, native mute/flood feedback,
  distance-aware audiences, room mute-all bypasses, configurable Redis flood
  control, global and room censorship, validated bubble styles, and bounded
  partitioned history written asynchronously with PostgreSQL COPY batches.
- Authorized whisper observers receive a localized `To {recipient}: {message}`
  packet; ordinary senders and recipients continue receiving the original text.
- Global and room filters compile immutable Aho-Corasick snapshots when their
  dictionaries change. Matching ignores separators and detects embedded text,
  so entries such as `chancleta` also censor `chan cleta` and `chancletacion`.
- Prefix, bold, and mention wire fields are intentionally absent because Nitro's
  chat packet shape does not contain them; do not add server-only packet fields.
- Arcturus renders rank prefixes by sending temporary `UNIT_CHANGE_NAME` packet
  `2182` before chat and restoring the username immediately afterward. Pixels
  does not implement this projection yet; preserve packet ordering and sanitize
  prefix HTML if that behavior is added explicitly.
- Test after changes:
  - `go test ./internal/realm/chat/... ./networking/inbound/chat/... ./networking/outbound/chat/... ./pkg/http/chat/routes`
  - `go test -run '^$' -bench . -benchmem ./internal/realm/chat/send ./internal/realm/chat/history ./pkg/textfilter`
  - In Nitro, verify talk radius, room-wide shout, private whisper, typing,
    mute/mute-all, flood feedback, censorship, and bubble selection.
  - Exercise `/api/admin/chat/filters`, `/api/admin/chat/bubbles`, and room/player
    chat history routes with and without `X-API-Key`.

### FEATURE: Room Settings, Filters, and Mute-All

- Owns `internal/realm/room/control/settings`, `internal/realm/room/control/wordfilter`, room
  settings commands/handlers, and the corresponding networking packets.
- Provides owner/rights/staff settings authorization, protocol-native settings
  request/save, optimistic updates, bcrypt password replacement, atomic tag
  replacement, live metadata refresh, room word filters, and ephemeral mute-all.
- Settings changes broadcast Nitro room-info, visualization, and chat refresh
  packets and publish `room.settings_updated`; filter and mute-all changes publish
  their own realm events.
- Test after changes:
  - `go test -race ./internal/realm/room/control/settings ./internal/realm/room/control/wordfilter ./internal/realm/room/control/commands/settings/...`
  - Run the room settings and word-filter packet tests under `networking/...`.
  - Run `BenchmarkContainsCached`, `BenchmarkCanManageOwner`, and
    `BenchmarkUpdateValidation` with `-benchmem`; hot lookups should allocate zero.
  - In Nitro, open room settings as owner and rights holder, save every tab, add
    and remove a filter word, toggle mute-all, and verify all current occupants
    receive the updated room information without reconnecting.

### FEATURE: Room Floor Plan Editor

- Owns `internal/realm/room/control/floorplan`, its commands and handlers,
  `room_custom_layouts`, and floor plan packets under `networking`.
- Resolves custom geometry by room id before falling back to the room's fixed
  model. Nitro packet `875` consumes its confirmed seven wire fields, while
  authoritative door height is derived from the parsed map. Packet `3559`
  reuses the existing entry-tile handler and also projects room thickness,
  while packet `1687` returns furniture-occupied tiles. Packet `1664` contains
  only door x, door y, and direction; door height must not be serialized there.
- Validates 64 by 64 geometry, charset, row shape, usable tiles, door, rotation,
  thickness, wall height, blocking furniture, and a distributed save cooldown.
- Successful active-room saves replace the server world in place and send a
  same-room `ROOM_FORWARD` to every occupant. Nitro cannot replace an existing
  room renderer from a new model packet alone; the forward rebuilds its local
  room session without disconnecting the WebSocket.
- Nitro does not send an auto-pickup field. The internal save contract supports
  bounded transactional auto-pickup for controlled server callers, while the
  Nitro packet handler always uses the conservative disabled value.
- Test after changes:
  - `go test ./internal/realm/room/control/floorplan ./internal/realm/room/control/commands/floorplan ./internal/realm/room/world/layout ./internal/realm/room/database/layout`
  - `go test ./networking/inbound/room/floorplan/... ./networking/outbound/room/floorplan/...`
  - Run `BenchmarkValidateSave` and `BenchmarkBlockedItems` with `-benchmem`.
  - Open the editor as an owner, rights holder, ordinary visitor, and staff
    member; verify only authorized users can request or save floor plans.
  - Save an empty room and verify every occupant stays connected, receives one
    same-room `ROOM_FORWARD`, and automatically respawns at the new door.
  - Place furniture over a changed tile and verify save is rejected with a
    localized floor-plan bubble until the furniture is removed.

### FEATURE: Closed Room Entry

- Owns `internal/realm/room/access/entry`, `internal/realm/room/access/doorbell`, room
  doorbell commands, and room doorbell packets.
- Supports bcrypt room passwords, Redis-backed attempt windows and lockouts,
  doorbell approval queues, timeout cleanup on the existing room tick, localized
  human durations, global `room.enter.any`, `room.enter.full`, and
  `room.doorbell.answer.any` nodes, and one-time admin entry bypasses.
- Persistent room-right holders and global doorbell responders may answer
  waiting requests. Owners and room-right holders bypass password, doorbell,
  and invisible modes through the `entry.RightsChecker` boundary.
- Test after changes:
  - `go test ./internal/realm/room/access/entry ./internal/realm/room/access/doorbell/...`
  - `go test ./internal/realm/room/access/commands/doorbell/... ./networking/...`
  - Enter password rooms with correct and incorrect passwords; verify lockout
    closes the password prompt and reports the configured localized duration.
  - Ring a doorbell, accept and reject it as owner and moderator, then verify the
    queue remains while an authorized responder is present and expires otherwise.
  - Call `POST /api/admin/rooms/players/:playerId/teleport` with `bypass` both ways.

### FEATURE: Room and Navigator Admin Routes

- Owns `pkg/http/room/routes` and related OpenAPI models.
- Provides protected room list/detail/occupancy/close/forward routes plus
  navigator categories and lifted room routes.
- Test after changes:
  - `go test ./pkg/http/...`
  - `GET /api/admin/rooms`, `/api/admin/rooms/:id`,
    `/api/admin/rooms/:id/occupancy`.
  - `POST /api/admin/rooms/:id/close` closes active runtime rooms.
  - `POST /api/admin/rooms/:id/forward` sends `room.forward` to active occupants.
  - `GET /api/admin/navigator/categories` and `/api/admin/navigator/lifted`.

### FEATURE: Catalog Realm

- Owns `internal/realm/catalog`, catalog packets under `networking`, and
  `pkg/http/catalog/routes`.
- Provides cached page trees, furniture offer projection, credits/points and
  numbered LTD purchases, embedded player catalog viewers, and command-driven
  packet handlers.
- Successful furniture purchases send Nitro's unseen-item marker before the
  purchase confirmation and then invalidate the inventory list.
- Provides protected catalog page/offer CRUD, cache publication, sanitize-list,
  OpenAPI documentation, localized Nitro external texts, and development seed
  data.
- Test after changes:
  - `go test ./internal/realm/catalog/... ./networking/... ./pkg/http/catalog/...`
  - Open Nitro catalog and verify the base and interaction seed pages.
  - Buy a regular offer and verify wallet deduction, the inventory novelty
    count, and immediate inventory refresh.
  - Buy the LTD sofa and verify remaining stock and sold-out behavior.
  - Call `/api/admin/catalog/refresh` and verify connected clients re-fetch.

### FEATURE: Furniture Inventory Synchronization

- Owns furniture inventory packets under `networking/outbound/inventory` and
  inventory projections under `internal/realm/furniture/commands/inventory`.
- Encodes floor and wall furniture, including wallpaper, floor paint, and
  landscape inventory categories.
- Purchases and pickups mark returned items unseen and send an inventory
  refresh; room pickup broadcasts remain limited to current room occupants.
- Active room furniture management is resolved through
  `room.CanManageFurniture`; owners and the room's embedded rights projection may
  place, move, and pick up furniture without handler-local permission policies.
- Global staff furniture management uses the explicit
  `room.furniture.any.manage` node. It must not create persistent room-right rows;
  picked-up furniture always returns to its actual owner's inventory.
- Test after changes:
  - `go test ./networking/outbound/inventory/... ./internal/realm/furniture/...`
  - Buy an item and verify Nitro shows the inventory novelty count.
  - Pick up a placed item with inventory open and verify it appears without a
    client reload.
  - Enter another player's room and verify place, move, and pickup return a
    localized no-rights bubble without changing inventory or room state.
  - Login as seeded `demo` and verify global furniture management works in a
    room owned by another player without adding `demo` to that room's rights.

### FEATURE: Paired Furniture Teleports

- Owns `internal/realm/furniture/interactions/teleport`, teleport pair
  persistence, furniture use packet `99`, and room movement walk-on events.
- Provides owner-validated durable pairs, click pads, zero-delay walk-on tiles,
  room-owner-cycle animation phases, authoritative controlled movement, and
  same-room or cross-room travel. Cross-room travel uses Nitro `ROOM_FORWARD`
  plus a one-time destination consumed before destination entity bootstrap.
  Destination furniture updates wait until after bootstrap so Nitro can render
  the open-door state before the controlled walk-out step.
- Catalog teleport offers grant an even number of instances and persist their
  pairs in the same transaction as charging and item creation. Pairing failure
  rolls back the purchase; runtime use never guesses a partner by proximity.
- Active transition lookup and interaction-tile detection stay in memory;
  animation ticks never query PostgreSQL and never create timers or goroutines.
- A normal teleport waits for the source and destination `mv` statuses to settle
  before forwarding or closing; do not send a neutral final status in the same
  owner tick as the visible movement packet.
- Cross-room teleports project source state `2` for one visual phase before
  closing and forwarding. Destination exits use interaction-controlled movement
  so normal furniture walkability cannot silently cancel the exit step.
- Both endpoints are reserved atomically for the duration of a transition, so
  simultaneous users cannot race shared furniture states or overlap transfers.
- An unreachable approach is a soft rejection: release the room transit and
  both endpoint reservations before returning, and never publish a started
  event for that rejected attempt.
- `PIXELS_FURNITURE_TELEPORT_BYPASS_LOCKED` may bypass password, doorbell, and
  invisible destination gates. Room bans always remain authoritative.
- Test after changes:
  - `go test -race ./internal/realm/furniture/interactions/teleport/... ./internal/realm/room/...`
  - `go test -run '^$' -bench . -benchmem ./internal/realm/furniture/interactions/teleport/...`
  - Click seeded items `1001`/`1002` in room `1` and verify opening, transfer,
    controlled walk-out, and synchronized close state for every occupant.
  - Use seeded item `1003` to reach item `1004` in room `2`; verify Nitro loads
    the destination room, opens the paired item, and visibly walks the player
    out before closing it.
  - Walk over seeded items `1005`/`1006` and verify use without clicking.
  - Buy a Telephone Box offer, place both granted items, and verify either side
    activates the other without manual database pairing.
  - Repeat a click during an active transition, remove a target, disconnect in
    transit, and test locked plus banned destination rooms.
  - Block the source approach tile with another player, verify the attempt is
    rejected, clear the tile, and verify the same player can retry immediately.
  - Activate both endpoints concurrently with different users and verify one
    transition completes before the pair accepts the next user.

### FEATURE: Toggle and Gate Furniture

- Owns `internal/realm/furniture/commands/interact`,
  `internal/realm/furniture/interactions/toggle`,
  `internal/realm/furniture/interactions/gate`, and the compact furniture state
  packet under `networking/outbound/room/furniture/state`.
- Packet `99` has one generic adapter. It delegates paired teleports before
  resolving build-right-protected `default`, `toggle`, and `gate` behavior; do
  not register competing handlers for the same packet.
- Generic toggles cycle `interaction_modes_count` defensively. Invalid durable
  state is logged and treated as state zero without aborting the interaction.
- Gate state `1` is open and walkable; every other value is closed. Closing or
  opening rebuilds the resolver fixture atomically with the runtime snapshot,
  while PostgreSQL uses compare-and-swap to prevent concurrent click drift.
- A gate cannot change while any unit occupies any tile in its rotated
  footprint. The occupancy check must remain allocation-free.
- Furniture state changes broadcast Nitro `FURNITURE_STATE` packet `2376` and
  publish `furniture.used`. Accepted movement publishes `furniture.walked_on`
  and `furniture.walked_off`; moving a generic interaction out from beneath a
  unit publishes a synthetic walk-off, except for rollers.
- Test after changes:
  - `go test -race ./internal/realm/furniture/... ./internal/realm/room/... ./networking/outbound/room/furniture/state`
  - `go test -run '^$' -bench . -benchmem ./internal/realm/furniture/interactions/toggle ./internal/realm/furniture/interactions/gate`
  - Click a normal multi-state item repeatedly and verify every occupant sees
    the complete state cycle without a full furniture-position update.
  - Buy and place the seeded Aquamarine Gate, open it, walk through it, and
    close it; verify a unit on any footprint tile prevents closing.
  - Start a path through an open gate and close it before arrival; verify the
    unit receives a neutral stop and does not cross the closed fixture.
  - Move a togglable item from beneath a unit and verify one
    `furniture.walked_off` event; moving a roller must not duplicate that event.

### FEATURE: Essential Furniture Interactions

- Owns `internal/realm/furniture/interactions/essential`, the room-owned task
  queue under `internal/realm/room/runtime/live/task`, and outbound unit hand
  item packet `1474`.
- Provides delayed dice/color-wheel/random-state resolution, pressure and color
  plates, controlled one-way gates and switches, physical multiheight updates,
  vending/hand-item delivery, and environmental cannon kicks.
- Nitro dice uses inbound packets `1990`/`1533` and outbound packet `3431`;
  color-wheel clicks use inbound packet `2144`. Keep these dedicated routes in
  addition to generic furniture use packet `99`.
- Placing or moving furniture must preserve the persisted owner and extra data
  in the runtime item. Losing the initial extra data breaks guarded delayed
  state resolution and leaves random furniture in its rolling animation.
- Delayed animation state is runtime-only. Final random, switch, and multiheight
  state uses guarded persistence; room close cancels every pending callback.
- Furniture workflows reserve unit control and must release it after success,
  failure, pickup, room leave, or target disappearance. Never create a timer or
  goroutine per furniture use.
- Hand items live on the room unit, not `Player`, and are projected to every room
  occupant. Avatar effects remain deferred until the complete effect holder and
  protocol behavior are implemented.
- Test after changes:
  - `go test -race ./internal/realm/furniture/interactions/essential ./internal/realm/room/runtime/live/...`
  - `go test -run '^$' -bench . -benchmem ./internal/realm/room/runtime/live/task ./internal/realm/furniture/interactions/essential`
  - Roll and close a Holodice; verify headers `1990`, `1533`, and `3431`, that
    `-1` prevents concurrent rolls, and that a final value appears.
  - Walk on/off the seeded Pressure Plate and verify every occupant sees `1/0`.
  - Use the Floor Switch from a distance and verify one controlled approach and
    exactly one state transition.
  - Cycle the Pink Pura Block with and without furniture stacked above it; verify
    heightmap and standing-unit Z updates remain synchronized.
  - Use the Mini-bar and Hand Item Tester and verify carried items are visible to
    all occupants.
  - Click a Cannon while Nitro is still walking to it; verify the click survives
    arrival, the shot follows `rotation + 6`, and owner/`room.unkickable` users
    remain while vulnerable users return to desktop with a localized notice.

### FEATURE: Room Rights, Moderation, and Audit

- Owns `internal/realm/room/control/rights`, `internal/realm/room/control/moderation`,
  `internal/realm/room/control/audit`, their room packets, and room audit HTTP routes.
- Persists build rights, current mutes and bans, append-only rights history, and
  append-only moderation history. Kick has no current-state row but is audited.
- Rights are projected into each active `live.Room`; do not add a separate
  rights registry. Furniture checks must remain an in-memory `O(1)` lookup.
- Audit subscribers run inside the mutation transaction. Runtime projections,
  packet broadcasts, kicks, and bans run only after a successful commit.
- Expected permission and validation denials are localized soft gameplay
  errors. Database, codec, and unexpected errors remain explicit failures.
- Nitro has no confirmed room mute-list inbound packet. Read active mutes through
  the moderation service or protected HTTP route; do not invent packet headers.
- Test after changes:
  - `go test ./internal/realm/room/control/rights/... ./internal/realm/room/control/moderation/... ./internal/realm/room/control/audit/...`
  - `go test ./networking/inbound/room/... ./networking/outbound/room/... ./pkg/http/room/routes`
  - Run room rights and moderation benchmarks with `-benchmem`.
  - Grant and revoke rights in Nitro; verify furniture controls and the rights
    level update immediately for every occupant.
  - Kick, mute, unmute, ban, and unban owner, rights-holder, moderator, protected,
    offline, and ordinary-player cases.
  - Verify bans block entry, rights bypass closed-room gates, and all successful
    mutations appear in the protected audit endpoints.

### FEATURE: Room Upvotes

- Owns `internal/realm/room/control/votes`, room vote persistence, vote packets,
  and `pkg/http/room/routes/votes`.
- Provides permanent idempotent room upvotes, owner exclusion, atomic score
  increments, per-occupant Nitro eligibility projection, entry bootstrap, and
  protected status, list, and cast administration routes.
- Test after changes:
  - `go test ./internal/realm/room/control/votes ./internal/realm/room/database/votes ./networking/inbound/room/like ./networking/outbound/room/score ./pkg/http/room/routes/votes`
  - `go test -run '^$' -bench . -benchmem ./internal/realm/room/control/votes ./networking/inbound/room/like ./networking/outbound/room/score`
  - Enter a room as its owner, a previous voter, and a new visitor; verify only
    the new visitor sees the active like control.
  - Vote twice and verify the score increments once while every occupant sees
    the same score with their own eligibility.
  - Open `/docs` and verify the `Admin Room Votes` route group.

## SDK Rules

- Treat `sdk/` as a controlled extension surface.
- Ask before adding new SDK APIs, exported types, extension hooks, or compatibility promises.
- Keep SDK additions decoupled from realm internals.
- Prefer explicit capability objects and small contracts over broad plugin access.

## Change Discipline

- Keep changes scoped to the requested behavior.
- Do not refactor unrelated code while implementing a feature.
- Split responsibilities before files or packages become hard to scan.
- Preserve the legacy tree unless the task explicitly targets it.
