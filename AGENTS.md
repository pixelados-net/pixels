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
- Classify Messenger packets below `messenger/friend`, `messenger/session`, or
  `messenger/social`; do not restore a flat package for every Messenger action.
- Classify Navigator packets below `navigator/browse`, `navigator/create`,
  `navigator/favorite`, or `navigator/session`; keep one packet per leaf package.

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

### FEATURE: Dynamic Plugins

- Owns `internal/plugin`, `sdk/plugin`, `sdk/command`, `sdk/event`,
  `sdk/player`, `sdk/priority`, plugin route registration, and the room-chat
  plugin seams.
- Uses capability-first packages below `internal/plugin`: `loader` owns native
  discovery and dependency ordering; `host` composes scoped capabilities;
  `player`, `command`, `event`, `route`, and `permission` own their individual
  SDK implementations; `runtime` owns only shared callback isolation state.
  Do not collapse those capabilities back into one runtime package.
- Loads embedded manifests from `plugins/<name>/*.so`, validates SDK majors,
  resolves dependencies before registration, and isolates plugin panics and
  callback timeouts without exposing mutable realm state or infrastructure.
- Plugin routes live below `/plugins/<name>` behind `X-API-Key` and publish
  separate OpenAPI documents. Plugin permissions live below
  `plugin.<name>.*`. Commands use Brigodier and are consumed before normal room
  chat; typed plugin events remain separate from the internal post-commit bus.
- Go native plugins must be rebuilt with the exact host Go version, platform,
  SDK checkout, and dependency graph. They are unsupported on Windows.
- Test after changes:
  - `go test -race ./internal/plugin/... ./sdk/... ./networking/connection ./internal/realm/chat/send ./pkg/http/pluginroutes`
  - `go test ./internal/plugin/runtime ./networking/connection -run '^$' -bench . -benchmem`
  - Build the ignored `plan/demoplugin` fixture with `-buildmode=plugin`, run
    its native verifier, then follow its README with Nitro.
  - Verify plugin routes reject missing API keys, permission-gated commands do
    not leak into chat, `chat.send` cancellation prevents delivery, and a
    panicking plugin scope does not stop unrelated plugins or native handlers.

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
- Provides protected, OpenAPI-documented player administration routes for
  idempotent creation, exact case-insensitive username lookup, profile reads,
  optimistic partial updates, and soft deletion. Creation assigns the default
  permission group atomically; deletion closes an active player session.
- Test after changes:
  - `go test ./internal/realm/player/...`
  - `go test ./pkg/http/player/routes ./pkg/http/openapi`
  - Authenticate with a seeded SSO ticket and verify user info bootstrap.
  - Enter and leave a room and verify live player room presence updates.
  - Create, find, update, and soft-delete a player through `/api/admin/players`;
    verify a deleted player cannot be found or authenticate again.

### FEATURE: Club Entitlements

- Owns player club fields, runtime entitlement projection, and HC gates.
- Provides a top-level `HC` catalog category with Nitro-native `vip_buy` and
  `club_gifts` pages, duration offers, and monthly gift metadata.
- Club level and expiration are loaded once with the player, projected through
  Nitro permissions, and reused by catalog and room settings without hot-path
  database reads.
- First membership, uninterrupted streak, durable accrual boundary, total club
  time, and total VIP time are separate fields. Never reconstruct active time
  from scheduler ticks or use first membership as the current streak.
- HC controls only wall visibility and wall/floor thickness; chat remains a
  normal room setting. Global settings managers may perform administrative
  overrides.
- Test after changes:
  - `go test ./internal/realm/player/... ./internal/permission/broadcast/...`
  - Login as seeded `demo` and verify Nitro receives HC level `1`; login as
    `alice` and verify VIP level `2` plus historical VIP days.
  - Open the `HC` catalog category, purchase a duration, and verify the wallet,
    expiration, club status, and available monthly gifts refresh correctly.
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

### FEATURE: Room Bundles

- Owns `internal/realm/room/record/bundle`, bundle persistence below room and
  furniture databases, room-bundle catalog offers, and protected administration
  routes below `/api/admin/rooms` and `/api/admin/catalog`.
- Bundle templates remain hidden from Navigator searches and owner room counts.
  A purchase atomically clones room settings, optional custom geometry, and
  grouped furniture, charges the buyer, writes purchase provenance, and emits
  `room.bundle.purchased` only after commit.
- Furniture cloning uses one `INSERT ... SELECT`, preserves placement and
  interaction state, and clears LTD, marketplace, trade, and gift ownership
  state. Bots remain outside bundle cloning until a bot realm exists.
- Test after changes:
  - `go test -race ./internal/realm/room/record/bundle ./internal/realm/catalog/service ./internal/realm/furniture/repository`
  - `go test ./pkg/http/room/routes ./pkg/http/catalog/routes ./pkg/http/openapi`
  - Run `BenchmarkPreview` with `-benchmem` and verify preview queries remain
    grouped instead of materializing every template item in Go.
  - Apply and validate schema plus development seeds from a clean database.
  - Open Nitro's complete-room catalog page, buy every seeded bundle, enter the
    resulting rooms, and verify geometry, furniture state, room limits, wallet
    rollback, Navigator visibility, and localized success/failure feedback.

### FEATURE: Crafting, Recycler, and Credit Exchange

- Owns `internal/realm/crafting`, crafting/recycler/exchange packets, the furniture `allow_recycle` projection, and protected crafting routes.
- Provides placed altar validation, visible and secret recipes, atomic exact ingredient consumption, concurrency-safe limited stock, exact-batch Ecotron recycling with injected rarity draws, and atomic placed/inventory credit furniture redemption.
- Nitro Renderer already owns the altar, Ecotron, and credit-furniture double-click triggers; Nitro React owns the crafting, recycler, and exchange widgets. Do not treat these actions as inventory-only flows.
- Test after changes:
  - `go test -race ./internal/realm/crafting/... ./networking/inbound/crafting/... ./networking/outbound/crafting/... ./networking/inbound/recycler/... ./networking/outbound/recycler/...`
  - `go test ./internal/realm/crafting/... -run '^$' -bench . -benchmem`
  - Apply and validate schema plus development seeds from a clean database.
  - Follow `CRAFTING-QA.md`, including double-click triggers, concurrent limited crafts, recycler invalid-batch rollback, and placed credit furniture removal.

### FEATURE: Camera, Photos, and Object Storage

- Owns `internal/realm/camera`, camera packets below `networking`, the
  `USER_SETTINGS_CAMERA` player-setting adapter, `pkg/storage`, photo evidence
  in moderation, and protected routes below `/api/admin/camera`.
- Provides bounded client-rendered PNG uploads, permanent public S3-compatible
  URLs, one active photo per player, explicit capture lifecycle, independent
  atomic multi-copy purchase, idempotent publication, deterministic room
  thumbnails, external-image furniture, photo CFH, audited moderation, and
  bounded lock-free metrics.
- Camera effects remain client-side. The server accepts only the flattened PNG
  at checkout and never reconstructs a room scene or interprets visual filters.
  Storage startup must fail fast when its bucket or credentials are unusable.
- Photo furniture and room thumbnails require durable public URLs; never replace
  them with expiring presigned URLs. Keep warmed metric updates allocation-free,
  and never move storage or database I/O into the room simulation tick.
- Every photo upload writes the canonical `.png` and Nitro's deterministic
  `_small.png` wall-rendering companion from the same immutable byte slice.
  Receipt rollback and orphan cleanup must delete both keys; thumbnails remain
  single-object uploads.
- Purchase and publication reuse the same active capture. Each purchase grants
  and charges an independent furniture copy; publication is unique and charged
  once per capture. New uploads supersede the prior active photo without
  breaking durable furniture or publication references.
- One aggregate scheduler deletes only expired pending or superseded objects
  proven to have no furniture link and no active publication. Storage I/O runs
  outside database locks, uses bounded batches and durable retries, and never
  enters the room tick.
- Test after changes:
  - `go test ./networking/inbound/camera/... ./networking/outbound/camera/... ./networking/inbound/session/render/... ./networking/inbound/user/settings/camerafollow/...`
  - `go test -race ./internal/realm/camera/... ./internal/realm/player/settings/... ./internal/realm/moderation/... ./pkg/storage/...`
  - `go test ./internal/realm/camera/... -run '^$' -bench . -benchmem`
  - Apply and validate schema plus development seeds from a clean database.
  - Follow `plan/CAMERA-QA.md`: capture, checkout, purchase, placed-photo viewer,
    publication cooldown, thumbnail rights, photo CFH, camera-follow persistence,
    lifecycle states, safe cleanup, admin permissions, concurrency, rollback,
    and Nitro build.

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

### FEATURE: Avatar Expressions and Effects

- Owns `internal/realm/room/world/action`, focused room action commands and
  handlers, `internal/realm/player/effect`, effect persistence, effect packets,
  effect-granting furniture, catalog effect rewards, and player effect routes.
- Provides persistent dances, transient signs and gestures, free-floor posture,
  manual and inactivity-driven idle projection, durable effect stacks capped at
  99 charges, one visible effect slot, active expiration, and synthetic primary
  permission-group effects.
- Replacing a deliberate avatar action clears the prior projection, waits the
  configured short transition delay, and then starts the replacement. Movement
  cancellation remains immediate and does not allocate timers on the walk path.
- Effect activation starts one charge once; selecting the same active effect
  must never restart its duration. The single global expiry query consumes at
  most one charge per selected row and clears a selected expired effect.
- Effect-giver furniture chooses from its configured pool and effect tiles use
  the live player's gender. Furniture and catalog rewards grant, activate, and
  select immediately; catalog changes remain inside the purchase transaction.
- Administrative HTTP grants select by default so they work without Nitro's
  missing effects panel; callers may send `enable: false` for inventory-only
  delivery. Grant-and-enable is atomic and projects inventory, activation,
  selection, and current-room state only after commit.
- Pure-effect catalog offers use Nitro product type `e` and must not resolve
  furniture definition zero. Mixed offers grant their furniture and effect in
  the same transaction.
- Test after changes:
  - `go test -race ./internal/realm/player/effect ./internal/realm/room/world/action ./internal/realm/furniture/interactions/essential/... ./internal/realm/catalog/...`
  - `go test ./networking/inbound/room/entities/... ./networking/outbound/room/entities/... ./networking/inbound/user/effect/... ./networking/outbound/user/effect/... ./pkg/http/player/routes ./pkg/http/openapi`
  - `go test -run '^$' -bench . -benchmem ./internal/realm/player/effect ./internal/realm/room/world/action`
  - Apply Liquibase and development seeds; verify `player_effects`,
    `players.active_effect_id`, permission group `room_effect_id`, furniture
    effect fields, and catalog effect reward fields.
  - With two clients in one room, test dance `0` through `5`, wave, kiss, laugh,
    sleep, sit, signs `0` through `18`, floor sit/stand, manual AFK, automatic
    AFK, movement cancellation, and late-entry dance/idle projection.
  - Click the seeded effect giver while adjacent and from far away, walk male
    and female avatars over the seeded effect tile, and verify immediate room
    projection plus inventory charge changes.
  - Buy permanent, one-day, HC-only, and mixed effect offers; verify club gates,
    balances, no furniture for pure effects, automatic selection, switching,
    disabling with effect zero, expiry, reconnect, and charge cap.
  - Open `/docs`, grant and revoke an effect through `Admin Players`, and verify
    default immediate selection, explicit `enable: false`, online incremental
    packets, current-room projection, and offline persistence.

### FEATURE: Navigator Realm

- Owns `internal/realm/navigator`.
- Uses capability-first packages: `browse` owns categories, room cards, searches,
  room information, and live counters; `create` owns room creation checks and
  execution; `favorite` owns favorite lifecycle events; `session` owns viewer
  bootstrap and session events; `record` defines persistence data and contracts;
  `database` implements PostgreSQL; and `core` coordinates shared persistence
  behavior. Packet adapters live with their commands. Do not restore realm-wide
  `commands`, `handlers`, `events`, `model`, `repository`, or `service` trees.
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

### FEATURE: Messenger Realm

- Owns `internal/realm/messenger`, messenger packets under `networking`, and
  `pkg/http/messenger/routes`.
- Uses capability-first packages: `friend` owns friendship mutations and their
  events, `session` owns bootstrap/search/profile requests and ignore handling,
  `profile` owns presence and observed-profile refreshes, `social` owns invites,
  follow, and private messages, `runtime` owns delivery and async chat logging,
  `record` owns domain records/contracts, and `database` owns PostgreSQL.
- Do not restore realm-wide `events`, `repository`, `service`, `privacy`, or
  `presence` packages. Events belong below the capability that publishes them.
- Provides directional friendships, unilateral relationship markers, pending
  requests, Nitro-native friend-list bootstrap and deltas, Redis-cached prefix
  search, room invitations, follow and populated-room discovery, private chat,
  live online/room presence projection, messenger privacy, persistent ignored
  users, and public relationship summaries.
- Messenger delivery reuses authenticated session bindings and connections; do
  not add a separate viewer, audience, or presence registry.
- Nitro's `MESSENGER_REQUESTS` packet owns the pending request count. Do not add
  minimail packets or invent a separate pending-count packet. Send this packet
  only for Nitro's explicit request and mutation refreshes, not during messenger
  init, because Nitro requests it immediately after init.
- Ignored users remain friends unless friendship is explicitly removed. Room
  talk, shout, whisper, typing, and private messenger delivery must honor the
  recipient's directional live ignore projection without database hot-path reads.
- Public-profile observation is embedded in each live player. Relationship
  changes refresh only players currently observing that profile through
  `messenger.relation.changed`; do not add a profile-viewer registry.
- Test after changes:
  - `go test -race ./internal/realm/messenger/... ./networking/inbound/messenger/... ./networking/outbound/messenger/...`
  - `go test ./internal/realm/messenger/core -run '^$' -bench . -benchmem`
  - Open Nitro with two seeded friends and verify online/offline and room
    presence change without refreshing the friend list.
  - Send, accept, decline, and remove friend requests; verify both online and
    offline targets, no duplicate initial request, and normal/HC list limits.
  - Ignore and unignore both friends and non-friends; verify ignored talk,
    shout, whisper, typing, and private messages are hidden only from the player
    who ignored the sender.
  - Assign heart, smile, bobba, and none relationships from the friend list and
    avatar menu; keep that player's profile open in a second client and verify
    its relationship summary updates without reopening the profile.
  - Search the same prefix from two players and verify the shared Redis cache;
    invite, follow, find a populated room, and exchange private messages.
  - Change room-invite privacy in Nitro and all privacy flags through the
    protected `Admin Messenger` routes in `/docs`.

### FEATURE: Moderation and Global Sanctions

- Owns `internal/realm/sanction`, `internal/realm/moderation`, moderation packets
  under `networking`, and `pkg/http/moderation/routes`.
- Moderation is capability-first: `cfh` owns reporter adapters and sanction
  status, `staff` owns the modtool queue/read/action adapters, `guide` and
  `guardian` own their domain managers with nested handlers, `runtime` owns
  shared live-session delivery, `core` coordinates common issue persistence,
  `record` defines contracts, and `database` provides PostgreSQL implementations.
  Do not restore a realm-wide moderation `handlers` package.
- Sanctions keep the common mutation workflow in `core`; registered side effects
  live below `enforcement/projection`, `enforcement/session`, or
  `enforcement/warning`; login hydration and ban gating belong to `session`;
  persistence contracts and PostgreSQL implementations remain in `record` and
  `database` respectively. Do not restore the flat `sanction/effect` package.
- Provides one global punishment record and registrable effects for bans, mutes,
  warnings, trade locks, and kicks; timestamp-derived active state; offline
  warnings; escalation policy; CFH topics and frozen evidence; atomic moderator
  issue claims; modtool information and actions; guide sessions; and durable,
  anonymized guardian voting.
- Global mute and trade-lock checks use the live player sanction snapshot. Do not
  add database reads to room-chat or direct-trade hot paths, and do not restore a
  second writer for `players.allow_trade`.
- Test after changes:
  - `go test -race ./internal/realm/sanction/... ./internal/realm/moderation/...`
  - `go test ./networking/inbound/moderation/... ./networking/outbound/moderation/... ./pkg/http/moderation/routes ./pkg/http/openapi`
  - `go test ./internal/realm/player/live ./internal/realm/sanction/core -run '^$' -bench 'Benchmark(MuteCheck|TradeLockCheck|ActiveSanctionLookup)$' -benchmem`
  - In Nitro, report another player, inspect pending calls, pick/release/close the
    issue from a moderator account, and verify the reporter receives the result.
  - Apply and revoke overlapping mutes and trade locks; verify chat and trading
    update immediately, and verify active bans reject login.
  - Put guide and guardian accounts on duty, complete one guide session with
    feedback, and complete acceptable, actionable, and tied guardian votes.
  - Open `/docs` and verify the `Admin Moderation` group; exercise every route
    with and without `X-API-Key`.

### FEATURE: Social Groups and Forums

- Owns `internal/realm/group`, social-group packets below `networking`, group
  furniture links, group catalog products, and `pkg/http/group/routes`.
- Uses capability-first `identity`, `membership`, `badge`, `forum`, `room`,
  `runtime`, `record`, and `database` packages. Social groups are not permission
  groups; hotel-wide overrides remain registered permission nodes.
- Provides atomic HC/currency-backed creation, normalized badge compilation,
  immutable group/player/room snapshots, join/request/roles/favorite lifecycle,
  HQ furniture return, room/navigator/profile/WIRED projections, group furniture,
  forum entitlement/policies/threads/posts/unread/moderation/CFH, protected audited
  administration, bounded telemetry, and development rooms `130`–`135`.
- Forum UI remains a Nitro client task documented in `TODO.md`; backend and wire
  contracts are complete. Packet `2864` is safe compatibility-only, while 265
  and 2445 remain encode-tested and are not sent to the current Nitro client.
- Test after changes:
  - `go test -race ./internal/realm/group/... ./networking/inbound/group/... ./networking/outbound/group/...`
  - `go test ./internal/realm/catalog/... ./internal/realm/furniture/... ./internal/realm/room/... ./internal/realm/navigator/... ./pkg/http/group/... ./pkg/http/openapi`
  - `go test -run '^$' -bench . -benchmem ./internal/realm/group/...`
  - Keep warmed membership, role, favorite, furniture-link, player snapshot, and
    badge registry lookups at zero allocations.
  - Apply and validate migrations plus development seeds on an empty PostgreSQL
    database, then follow `GROUPS-QA.md` for rooms `130`–`135`, APIs, forum
    harness, rollback, concurrency, lifecycle, and authorization checks.

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

### FEATURE: Rollers and Room Heights

- Owns fixed-point room heights under `internal/realm/room/world`, roller
  interactions under `internal/realm/furniture/interactions/roller`, packet
  `3207`, room roller speed, and the roller development seed.
- Heights use quarter-unit fixed point. Convert only at database and wire
  boundaries; world physics must never mix scaled and unscaled values.
- Physical furniture volumes suppress covered walking planes and enforce avatar
  clearance. Invalidated paths re-anchor only to walkable sections and trapped
  units retain an escape path.
- One room-owned cycle advances indexed rollers. A unit or item moves at most
  once per cycle, walking units take precedence, occupied destinations block,
  furniture never rolls uphill, and walk hooks run after the configured delay.
- Roller speed `-1` disables movement, `0` advances every room tick, and positive
  values count 500ms room cycles. Nitro does not send this field in its settings
  packet, so live edits use `PATCH /api/admin/rooms/:id/roller`.
- Test after changes:
  - `go test -race ./internal/realm/furniture/interactions/roller ./internal/realm/room/world/... ./internal/realm/room/runtime/live ./networking/outbound/room/furniture/rolling`
  - `go test -run '^$' -bench BenchmarkRollerTick -benchmem ./internal/realm/furniture/interactions/roller`
  - Build a three-roller chain and verify one-tile-per-cycle movement, packet
    animation, durable final positions, stationary-unit movement, and destination
    blocking.
  - Stack furniture and verify fractional heights, blocked occupied volumes,
    avatar clearance, path invalidation recovery, and no movement freeze.
  - Change roller speed through the protected admin route and verify active rooms
    adopt the new cadence without reconnecting.

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
- Gift purchases use Nitro's four-list wrapping configuration, validate client
  wrapping indexes, confirm the buyer purchase, and refresh an online recipient's
  unseen inventory without making offline delivery fail.
- Unopened gifts retain the contained furniture definition but project the
  selected positive wrapper sprite and packed box/ribbon variant in inventory
  and rooms; negative furniture ids are reserved for contents revealed later.
- Provides protected catalog page/offer CRUD, cache publication, sanitize-list,
  OpenAPI documentation, localized Nitro external texts, and development seed
  data.
- Test after changes:
  - `go test ./internal/realm/catalog/... ./networking/... ./pkg/http/catalog/...`
  - Open Nitro catalog and verify the base and interaction seed pages.
  - Buy a regular offer and verify wallet deduction, the inventory novelty
    count, and immediate inventory refresh.
  - Buy the LTD sofa and verify remaining stock and sold-out behavior.
  - Gift a seeded furniture offer to an online and offline player; verify the
    buyer dialog closes and the online recipient receives an inventory novelty.
  - Call `/api/admin/catalog/refresh` and verify connected clients re-fetch.

### FEATURE: MISC Furniture Catalog

- Owns MISC furniture/catalog development seeds and
  `internal/realm/catalog/trophy`.
- Provides asset-safe wall decoration, plants, carpets, lighting, dividers,
  complete classic lines, shared-definition poster variants, trophies, and
  multi-product decoration packs.
- Supports definition-configured reversed gate states for assets whose closed
  Nitro frame is state `1`, without changing normal gate semantics.
- Trophy purchases consume client extra data only for trophy definitions. The
  server resolves the buyer, strips protocol separators, applies the global
  hotel filter, truncates Unicode text, and persists the immutable inscription.
- Test after changes:
  - `go test ./internal/realm/catalog/trophy ./internal/realm/catalog/...`
  - Run Liquibase validation and apply the development seed changelog.
  - Verify generated definitions have unique sprites, whitelisted interactions,
    and enabled offers; keep the sanitize-list baseline documented in
    `docs/MISC.md`.
  - In Nitro, exercise every new page, poster art, wall placement, carpet
    traversal, closed/open Bazaar curtain collision, classic sit/lay slot,
    trophy inscription, the top-level room-paint page, and both packs.

### FEATURE: Store Final and Subscriptions

- Owns catalog bundle, gift, voucher, and freshness behavior plus
  `internal/realm/subscription`, its packets, scheduler, seeds, and protected
  administration routes.
- Supports server-authoritative bulk discounts, multi-product bundles, wrapped
  furniture gifts, one-time vouchers, catalog novelty and expiration, HC/VIP
  purchase and extension, kickback paydays, monthly club gifts, targeted offers,
  and seasonal calendar rewards.
- Subscription membership is durable state. Player club fields and the embedded
  live player entitlement are derived projections updated only after successful
  membership persistence.
- The subscription scheduler reconciles timestamp-based active time, expires
  memberships, materializes every missed payday period, pays newly due rewards
  immediately to online players, and pays offline rewards on the next login.
  Player connection bootstrap sends active calendar state, available monthly
  gifts, and neutral Builders Club compatibility state.
- Catalog purchase history links every purchase row to the concrete furniture
  instances it granted. Payday period queries use committed purchase timestamps
  and exclude pages explicitly marked outside kickback accounting.
- Targeted offers must have future expiration, banner, icon, localized copy,
  enabled state, and remaining per-player capacity before projection. Real
  audience eligibility remains a documented TODO and must not be invented ad hoc.
- Builders Club is a compatibility stub only and must remain explicitly marked
  `UNIMPLEMENTED:`. Direct SMS billing remains `DEFERRED:` until a carrier
  provider exists. Neither path may grant limits, membership, furniture, or
  currency through its compatibility response.
- Test after changes:
  - `go test -race ./internal/realm/catalog/... ./internal/realm/subscription/...`
  - `go test ./networking/inbound/catalog/... ./networking/outbound/catalog/...`
  - `go test ./networking/inbound/subscription/... ./networking/outbound/subscription/...`
  - `go test -run '^$' -bench . -benchmem ./internal/realm/catalog/service ./internal/realm/subscription/core ./networking/codec`
  - Run Liquibase `validate` and apply both schema and development seed
    changelogs after changing store persistence.
  - Buy six or more units and verify the authoritative discount, then buy the
    seeded multi-product bundle and verify every product appears in inventory.
  - Send and open a wrapped gift, redeem `WELCOME2026` twice, and verify the
    second redemption cannot duplicate its reward.
  - Purchase and extend HC/VIP, inspect kickback, claim one monthly gift, and
    reconnect to verify membership and pending payday state survive.
  - Verify `demo` receives one due payday, `alice` projects VIP history, `bob`
    catches up two independent periods, and expired `carol` starts a fresh
    streak without retaining VIP.
  - Stop the process between lifecycle passes, restart it, and verify elapsed
    club time is reconciled once without duplicate payday rows.
  - Purchase or dismiss a targeted offer and open normal, future, duplicate,
    and staff calendar doors.
  - Create, patch, and disable store records through `/api/admin/catalog` and
    `/api/admin/subscriptions`; verify catalog mutations publish exactly one
    catalog refresh packet per connected client.

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
- Hand items and visible avatar effects live on the room unit, not `Player`, and
  are projected to every room occupant. Durable effect ownership remains in the
  player effect service.
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

### FEATURE: Marketplace and Direct Trade

- Owns `internal/realm/marketplace`, `internal/realm/trade`, Marketplace and
  Trade packets under `networking`, focused furniture/player trading fields,
  and `pkg/http/trading/routes`.
- Marketplace preserves seller ownership while an item is durably reserved in
  limbo, hides it from inventory, applies upward-rounded buyer commission, and
  settles purchase, delivery, cancellation, expiry, tokens, and redemption in
  shared PostgreSQL transactions.
- Nitro exposes Marketplace through the seeded `marketplace` and
  `marketplace_own_items` catalog layouts; keep both pages in the catalog tree.
- Before opening the Marketplace posting form, Nitro must request packet `848`.
  A zero token balance must offer packet `1866`, wait for packet `54`, and open
  the form only after the server confirms a positive balance. Do not silently
  buy tokens from the listing handler.
- Direct Trade keeps sessions and staged items in synchronized runtime indexes.
  Every offer mutation resets agreement; settlement revalidates every item and
  transfers or consumes both offers atomically. Room leave and disconnect cancel
  the trade and clear unit status.
- Trade protocol participant fields are durable player/web ids, not room-local
  unit ids. Unit ids remain internal to room targeting and status projection.
- After durable settlement, send Nitro `TRADE_COMPLETED` before furniture
  inventory invalidation packet `3151` to both participants. The invalidation
  clears staged client groups and reloads transferred ownership; packet `3928`
  is only the server ping and must never substitute for an inventory refresh.
- Trade rejection packet `217` must preserve Nitro's native reasons: `1` for a
  globally disabled hotel, `6` for room policy, `7` for the requester cooldown
  or busy state, and `8` for an unavailable target. Do not collapse room policy
  into the global-disabled message.
- Community Goals packets remain deferred until a real implementation or wire
  reference exists; do not invent those twelve packet shapes.
- Test after changes:
  - `go test -race ./internal/realm/marketplace/... ./internal/realm/trade/...`
  - `go test ./networking/inbound/marketplace/... ./networking/outbound/marketplace/... ./networking/inbound/trade/... ./networking/outbound/trade/...`
  - `go test -run '^$' -bench . -benchmem ./internal/realm/marketplace/core/tests ./internal/realm/trade/core ./internal/realm/trade/runtime`
  - Apply Liquibase and verify Marketplace listing/token/stat tables, trade
    audit logs, furniture flags/LTD reservation, and the player trade lock.
  - Click Sell with zero tokens, confirm the configured token purchase, and
    verify credits and token balance refresh before the posting form opens.
  - Initiate a trade as a normal member in the seeded Pixels Lobby and verify it
    opens; disable room trading and verify Nitro reports the room-specific error.
  - List/buy/cancel/expire/redeem regular and LTD items, then race two buyers
    against one listing and verify exactly one wins.
  - Trade regular and redeemable-credit furniture, mutate offers after accept,
    leave/disconnect mid-trade, and verify no item can be staged, placed, or
    listed in two workflows at once.
  - Open `/docs` and verify `Admin Trading`; lock/unlock one player, read their
    trade log, and force-close an open Marketplace listing.

### FEATURE: Room Bots

- Owns `internal/realm/bot`, `sdk/bot`, bot packets under `networking`, and
  `pkg/http/bot/routes`.
- Provides durable inventory and placed bots, owner/rights authorization,
  bounded room and inventory limits, Nitro type-four unit projection, native
  bot configuration skills, random walking, automatic filtered chat, following,
  and deferred position persistence.
- Bot movement advances from the existing room-owned 500ms loop. Never create a
  timer or goroutine per bot; slow behavior hooks use the fixed shared worker
  pool, and the ordinary cycle must remain free of database I/O and allocations.
- Built-in behaviors are `generic`, `bartender`, and `visitor_log`. External
  behavior implementations register through the controlled `sdk/bot` contract;
  unknown durable behavior names fall back safely to `generic`.
- Bartender keywords are whole-word, admin-managed mappings. Visitor history is
  bounded and starts at the room owner's durable last logout. Bot speech passes
  through global and room filters plus the room WIRED speech bridge; only a
  matching SAY stack consumes normal delivery.
- Complete-room bundle purchases clone template bots and ordered chat lines in
  the same transaction when `PIXELS_ROOM_BUNDLE_BOTS_ENABLED` is true.
- Test after changes:
  - `go test -race ./internal/realm/bot/... ./internal/realm/room/... ./networking/inbound/bot/... ./networking/outbound/bot/...`
  - `go test ./pkg/http/bot/routes ./pkg/http/openapi ./networking/inbound/inventory/bots ./networking/outbound/inventory/bots/...`
  - `go test ./internal/realm/bot/core/tests -run '^$' -bench BenchmarkBotCycleTick -benchmem`
  - Place, configure, walk, follow, pick up, and delete a seeded bot; verify the
    inventory and room entity update without reconnecting.
  - Say exact and substring bartender keywords at near and far distances; verify
    only an exact nearby word transfers the hand item.
  - Enter the visitor bot's room as its owner after other visits, accept the
    summary, and verify the bounded list appears only once per placement.
  - Buy a complete-room bundle containing a template bot and verify the cloned
    room owns a separate bot with the same placement and chat configuration.
  - Open `/docs` and verify the protected `Admin Bots` route group.

### FEATURE: Room Decoration and Remaining Room Entities

- Owns `internal/realm/room/decoration`, decorator commands below
  `internal/realm/furniture/commands/decor`, wall-furniture projections, room
  appearance persistence, and hand-item transfer commands below
  `internal/realm/room/world/commands/handitem`.
- Provides consumable floor colors/patterns, wallpaper, and window landscapes;
  generic wall-item placement, movement, pickup, and entry snapshots; editable
  filtered post-its with sticky-pole guest placement; mannequin outfit naming,
  saving, and same-gender application; three durable mood-light presets; and
  typed background-toner object data.
- Room bundles clone floor, wallpaper, landscape, wall items, decorator object
  data, and dimmer presets in their existing purchase transaction. Development
  seeds include catalog surface variants and a decorated bundle template.
- A room may contain only one dimmer. The persistence guard serializes placement
  by room so concurrent requests cannot bypass the limit.
- Player hand-item drop and adjacent transfer resolve room-local unit ids through
  the world's reverse unit index. Keep that lookup at zero allocations and do
  not scan all units or query persistence on this path.
- `UNIT_NUMBER` and `UNIT_INFO` remain focused outbound encoders; they do not own
  additional room state.
- Test after changes:
  - `go test -race ./internal/realm/room/... ./internal/realm/furniture/...`
  - `go test ./networking/inbound/furniture/... ./networking/outbound/room/...`
  - `go test -run '^$' -bench BenchmarkUnitByID -benchmem ./internal/realm/room/world/runtime/tests`
  - Buy and apply every seeded floor, wallpaper, and landscape variant; verify
    every current occupant updates immediately and re-entry restores the choice.
  - Buy, place, move, and pick up a mood light; verify a second mood light is
    rejected, all three presets save, and toggle state survives a restart.
  - Place a post-it as owner and through a sticky pole as a visitor; edit color
    and filtered text, then re-enter and verify the durable wall snapshot.
  - Save and rename male and female mannequins; apply them with matching and
    mismatched genders and verify only clothing parts change in the valid case.
  - Configure and toggle a background toner and verify other occupants and late
    entrants receive its integer-array object data.
  - Give a hand item to an adjacent player, reject a distant target, and drop it;
    verify every occupant sees both unit hand-item updates.
  - Buy the decorated complete-room bundle and verify its surfaces, post-it,
    mannequin, toner, mood light, and mood-light presets are independent copies.

### FEATURE: Room WIRED

- Owns `internal/realm/room/world/wired`, its room commands/handlers/database,
  WIRED packets below `networking`, and `pkg/http/room/routes/wired`.
- Provides an audited registry for all 76 canonical Arcturus behaviors plus the
  explicit `wf_cnd_valid_moves` extension: 17 triggers, 30 effects, 24
  conditions, selectors/blob, and durable highscore boards. Unknown SQL-only
  `wf_*` assets never become executable implicitly.
- Compiles immutable per-room stack generations with indexed event lookup,
  deterministic AND/OR/random/unseen semantics, breadth-first call stacks,
  bounded trace/effect/delay budgets, room-owned timers, compact execution activation,
  lifecycle invalidation, low-cardinality metrics, and a bounded trace ring.
- Box animation follows Arcturus execution order: toggle the matched trigger,
  successful extras, and only effects that actually apply (after their delay).
  Conditions and unselected effects do not animate. Animation uses compact
  `FURNITURE_STATE` (`2376`), never a full placement update (`3776`).
- Chat, completed movement, furniture state, bots, moderation, social groups,
  equipped badges, teams, games, rewards, and highscore projection use focused
  realm boundaries. Dispatch performs no PostgreSQL reads and an event without
  candidates must remain zero-allocation.
- Nitro supports five inbound editor packets and seven outbound WIRED packets;
  object-data packets `1453` and `2547` carry typed highscore state. Config saves
  use optimistic versions and same-room targets, and pickup cleans reverse links.
- Development seeds provide the complete functional catalog plus six configured
  rooms (`110` through `115`) covering every registered behavior, bots, social
  group/badge conditions, rewards, game pieces, failure paths, and all twelve
  highscore variants.
- Test after changes:
  - `go test ./networking/inbound/furniture/wired/... ./networking/outbound/furniture/wired/... ./networking/outbound/room/furniture/objectdata/...`
  - `go test -race ./internal/realm/room/world/wired/... ./internal/realm/room/database/wired/...`
  - `go test ./pkg/http/room/routes/wired ./pkg/http/openapi ./internal/realm/room/... ./internal/realm/furniture/... ./internal/realm/bot/... ./internal/realm/chat/...`
  - `go test ./internal/realm/room/world/wired/runtime/tests -run '^$' -bench . -benchmem`
  - Apply Liquibase with development context and enter rooms `110`–`115`; open,
    save, and re-open one trigger, condition, and effect before exercising the
    full checklist in `plan/rooms/WIRED.md` Part 12.
  - Close/reopen a room with pending timers and delays, pick up selected targets,
    run a call-stack cycle, and confirm old generations do not execute or leak.
- Keep packet logs open around periodic stacks: box animation must use `2376`;
    `3776` is reserved for real furniture placement/state mutations.
  - Use `/docs` to inspect `Admin WIRED`, manage custom settings/rewards,
    start/end/reset the QA game, toggle hidden boxes, and inspect traces/metrics.

### FEATURE: Progression — Achievements, Talents, Quests, and Quizzes

- Owns `internal/realm/progression`, progression packets below `networking`,
  progression administration below `pkg/http/progression`, and the small
  committed gameplay events consumed from other realms.
- Uses capability-first `achievement`, `engine`, `talent`, `quest`, `quiz`,
  `poll`, `promo`, `trigger`, `compat`, `record`, and `database` packages.
  Badge inventory remains in `internal/realm/player/achievement`; progression
  coordinates it through focused contracts and never creates a second badge
  owner.
- Provides 51 data-driven achievement groups with 452 cumulative levels,
  immutable trigger indexes, atomic multi-level rewards, slot-preserving badge
  replacement, durable score, derived talent tracks, optional TRADE and guide
  gates, campaigns, single-active quests, daily/seasonal offers, exact goal
  metadata, safety quizzes, room word polls, promotional badges, and protected
  audited administration.
- Gameplay publishers emit committed realm events. The progression subscriber
  performs a zero-allocation definition lookup and queues bounded write-behind
  deltas; it must not add PostgreSQL reads to movement, room entry, chat,
  furniture placement, or other gameplay hot paths. Hydrated in-memory progress
  forecasts threshold crossings so level-ups flush immediately, while ordinary
  deltas stay batched.
- Achievement level and reward mutations lock the player-definition row and
  execute badge, currency, score, and progress changes in one shared PostgreSQL
  transaction. Quest completion and promotional claims are replay-safe; admin
  mutations and their audit row commit atomically.
- Nitro owns the existing Achievements and word-quiz views. Quest, talent,
  safety-quiz, and promotional-claim views are client work documented in
  `TODO.md`; backend and wire behavior remain testable through packet harnesses
  and `/docs` without inventing success state.
- Resolution, game-achievement, and competition contracts without trustworthy
  gameplay semantics remain explicit neutral compatibility adapters. Never
  fabricate rankings, contests, progress, or rewards merely because a header
  exists.
- Test after changes:
  - `go test ./internal/realm/progression/... ./networking/inbound/progression/... ./networking/outbound/progression/...`
  - `go test -race ./internal/realm/progression/... ./internal/realm/player/achievement/...`
  - `go test ./internal/realm/progression/... -run '^$' -bench . -benchmem`
  - Keep trigger lookup, pending batch recording, and metrics recording at zero
    allocations; explain and benchmark any regression before merging it.
  - Apply and validate schema plus development seed twice; verify 51
    definitions, 452 levels, eight quests, six talent levels, five safety
    questions, and no duplicate rows.
  - Follow `plan/QUEST-ACHIEVEMENTS-QA.md` for exact Nitro buttons, personas,
    packet-harness flows, `/docs` operations, rewards, permissions, and
    compatibility checks.

### FEATURE: Room Games, Game Center, and Polls

- Owns server-authored Banzai, Freeze, Football, Tag, and game timers below
  `internal/realm/room/world/games`; shared teams, scores, highscores, and game
  triggers remain below `internal/realm/room/world/wired/game`.
- Room games advance only on the existing room owner loop. Never add a
  goroutine or timer per match, ball, or explosion; delayed work belongs to the
  room scheduler. Banzai mass captures use objectdata batch 1453.
- WIRED score is tracked separately and excluded from achievement score.
  Progress-achievement/start-quest/progress-quest effects require
  `room.wired.admin`; ordinary room rights are insufficient.
- Game Center is an external URL/parameter launcher backed by an immutable
  cache and audited administration below `/api/admin/games`. Empty URL means an
  honest queue failure; external arena protocol remains decode-only.
- DB polls use an immutable `roomID -> poll` cache for room entry, persist each
  answer once, and grant optional badges idempotently. Word quiz remains a
  separate bounded live room flow. Renderer infobus wire exists, but its React
  view is deferred in `TODO.md`.
- Test after changes:
  - `go test -race ./internal/realm/room/world/games/... ./internal/realm/room/world/wired/... ./internal/realm/gamecenter/... ./internal/realm/progression/poll/...`
  - `go test ./networking/inbound/gamecenter/... ./networking/outbound/gamecenter/... ./networking/inbound/progression/poll/... ./networking/outbound/progression/poll/...`
  - `go test -run '^$' -bench . -benchmem ./internal/realm/room/world/games/...`
  - Apply and validate schema plus development seeds from a clean database;
    follow `plan/GAMES-QA.md` in rooms 150–153 with two clients.

### FEATURE: Rooms Finals, Subscription Landing, and Protocol Compatibility

- Owns room promotions, spectator entry, final room-control actions,
  non-media furniture interactions, Subscription hotel-view packets, and the
  explicit Notifications/Other compatibility closure.
- Room Ads purchase is one PostgreSQL transaction across Catalog charging and
  promotion upsert. Catalog and room projections run only after commit; service
  rewards never create phantom furniture.
- Spectators count outside visible room capacity, receive room state, never own
  a world unit, and must not produce joined/left avatar ghosts.
- Rentables calculate expiry from timestamps rather than ticks. Lovelock uses a
  guarded two-player state machine. Mystery Box waits on the room scheduler,
  Trophy inscriptions are owner-only and filtered, and fireworks recharge on
  that same scheduler while publishing committed progression events.
- Builders Club purchase-and-place remains dormant when
  `PIXELS_BUILDERS_CLUB_FURNITURE_LIMIT=0`; a positive test policy must preserve
  the shared purchase transaction and authoritative placement checks.
- Retired room, interstitial, phone, FAQ, and campaign requests decode strictly
  and intentionally perform no mutation. Their code comments must retain the
  client-evidence reason for NOOP behavior.
- The protocol audit has exactly 20 packets left: Furniture Media for Trax,
  Jukebox/Sound Machine, and YouTube display. No other realm may be described as
  missing a packet package; see `plan/STATUS.md`.
- Test after changes:
  - `go test -race ./internal/realm/room/... ./internal/realm/furniture/... ./internal/realm/catalog/... ./internal/realm/subscription/...`
  - `go test ./networking/inbound/room/... ./networking/outbound/room/... ./networking/inbound/furniture/... ./networking/outbound/furniture/... ./networking/inbound/subscription/... ./networking/outbound/subscription/...`
  - `go test ./internal/realm/connection/compatibility ./internal/realm/moderation/staff ./networking/inbound/notification/... ./networking/inbound/other/... ./networking/outbound/notification/... ./networking/outbound/other/...`
  - `go test -run '^$' -bench . -benchmem ./internal/realm/room/promotion ./internal/realm/room/runtime/live ./internal/realm/furniture/interactions/... ./internal/realm/catalog/commands/builders ./internal/realm/subscription/...`
  - Apply and validate schema plus development seeds, then follow
    `plan/finals/FINAL-QA.md` in room `160` with two clients.

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
