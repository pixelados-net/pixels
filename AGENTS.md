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
  default `member` assignment, and seeded `admin` access.
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
  rooms, category preferences, and debounced live category counts.
- Test after changes:
  - `go test ./internal/realm/navigator/...`
  - In Nitro, open navigator and verify metadata tabs, flat categories, settings,
    saved searches, favorites, lifted rooms, and collapsed categories.
  - Search hotel/myworld/official views and verify room cards show live counts.
  - Create a room and verify `navigator.room_created`.
  - Request room info and verify missing rooms return `navigator.nosuchflat`.

### FEATURE: Room Realm

- Owns `internal/realm/room`.
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
- Active room furniture management is owner-only through
  `room.CanManageFurniture`; place, move, and pickup commands must check it
  before persistence or world mutation. Extend that capability when room rights
  are introduced instead of adding handler-local permission policies.
- Test after changes:
  - `go test ./networking/outbound/inventory/... ./internal/realm/furniture/...`
  - Buy an item and verify Nitro shows the inventory novelty count.
  - Pick up a placed item with inventory open and verify it appears without a
    client reload.
  - Enter another player's room and verify place, move, and pickup return a
    localized no-rights bubble without changing inventory or room state.

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
