# Architecture

Pixels is organized around four top-level layers: reusable infrastructure in `pkg/`, feature behavior in `internal/`, packet coding in `networking/`, and a controlled extension surface in `sdk/`. This page focuses on the middle one, since that's where almost all of the actual game logic lives, and walks through what each realm actually does, not just what its name implies. For the exact package layout rules and naming conventions, see [`AGENTS.md`](https://github.com/niflaot/pixels/blob/main/AGENTS.md) in the repository.

## The realm pattern

Everything player-facing is organized as a **realm**: a self-contained feature domain living under `internal/realm/<name>`, with its own persistence, its own commands and packet handlers, and its own tests. A realm doesn't reach into another realm's internals. When it needs something another realm owns, it depends on that realm's exported service interface, wired together through [`go.uber.org/fx`](https://github.com/uber-go/fx) at startup rather than through shared global state.

Most realms follow a recognizable internal shape, though larger ones split it further by capability instead of keeping one flat folder per concern:

- **`record/`** holds the domain model and the storage contract, meaning plain Go structs and the interface a database layer has to satisfy, with no PostgreSQL-specific code in sight.
- **`database/`** is the PostgreSQL implementation of that contract, plus the realm's own Liquibase migrations and development seed data.
- **Commands and handlers** decode an inbound packet, turn it into a domain command, and dispatch it. Handlers stay thin; the actual behavior lives in a service.
- **`events/`** are the things a realm publishes after something durable happens, such as a room being created or a trade completing, that other realms can react to without being directly coupled to the realm that caused them.
- **`config/`** or **`policy/`** holds the realm's own tunables, each with a safe default, loaded from environment variables.

A packet handler and the domain logic it triggers are always in the same realm as the data they touch. Following "what handles packet X" is a matter of finding the realm that owns the feature X belongs to, not searching the whole codebase.

## Realms

### `player`

Core account state that isn't specific to any one feature: identity and login, profile fields such as motto and current look, wardrobe (owned and redeemable clothing pieces), and avatar effects granted by the catalog (dances, gestures, one-off animations). It also owns the badge inventory itself: which badges a player owns, which ones are equipped, and in which slots, plus the daily respect ledger. The engine that actually decides when a badge is *earned* by leveling up an achievement lives in `progression`; `player` is where the resulting badge is owned and worn.

### `session`

Binds an authenticated connection to a player for as long as that connection lasts. Deliberately small: it's the join point every other realm uses to go from "a WebSocket sent this packet" to "this specific player sent this packet," not a place with feature logic of its own.

### `connection`

Registers every realm's packet handlers against the shared inbound handler registry at startup, and nothing more. Kept intentionally thin so that transport and security concerns never leak into realm code.

### `room`

The largest realm, and the one most others plug into. It owns room layout and the floor plan editor, live occupancy and entry, including a doorbell approval flow for locked rooms and dedicated closed-room entry rules, furniture placement and pickup, per-room rights and moderation (kick, mute, ban, and an append-only audit log), room chat delivery, room likes, and rollers with real height resolution: a column-and-section model per tile rather than one flat stacked height. It also owns Wired end to end, meaning a full registry of triggers, effects and conditions, including legacy header aliases, reward-limit tracking, and a bridge into bots and room units.

On top of that substrate, `room` implements four classic furniture games end to end, all server-authored: Battle Banzai, where claiming a tile runs a real flood-fill to lock the largest enclosed area a color owns; Freeze, with snowball throwing, area-of-effect explosions, power-ups and lives; Football, with real ball physics driven by an eight-direction bounce table and goal detection; and the tag variants (Ice Tag, Rollerskate, Bunnyrun). Team membership, scoring, and highscore boards for all four share one generic engine rather than four separate ones.

### `gamecenter`

Deliberately thin, and that's accurate rather than an oversight: this is the external game lobby, meaning the game list, join/leave queue, and launching a third-party game by URL. The real, playable game logic described above lives in `room`, not here. `gamecenter` exists because the Nitro client's own game center UI still expects a lobby protocol to talk to, even for games the client itself never renders.

### `furniture`

Furniture definitions and player inventory of furniture, plus the dispatch table behind every per-item interaction: gates, multi-state toggles, teleports, dice, post-it notes, mannequins, vending machines, and the rest of the catalog of "furniture that does something when you click it." Placement, pickup, and stacking height all live here too, feeding the column model `room` uses for pathing.

### `catalog`

The product catalog itself and the purchase flow that turns a catalog page offer into owned furniture, a currency grant, or both.

### `inventory`

The player's wallet: a multi-currency balance (credits, duckets, diamonds, and any other configured type), admin grant/deduct/set operations, and an optional audit ledger per currency type.

### `crafting`

Crafting altars with both openly visible and secret recipes resolved by exact-ingredient matching, a recycler that consumes an exact batch of items for a random prize drawn from a configured pool, and credit-furniture exchange, all committed atomically so a failed reward never leaves ingredients half-consumed.

### `camera`

The in-room camera end to end: the client renders and uploads the photo, `camera` stores it in S3-compatible object storage, and a purchase mints a real furniture item carrying the photo's data so it can be hung on a wall and reopened later. It also owns publishing a photo to a public gallery with a cooldown, rendering a room's thumbnail, and reporting a photo into moderation's call-for-help queue.

### `progression`

The trigger-based engine behind achievements, talent tracks, quests, and quizzes. Achievements are leveled definitions where crossing a threshold replaces the player's badge for that group with the next level's badge and pays a configured reward; talent tracks derive their own level from a set of underlying achievements and grant perks once enough of them are reached. Quests are organized into campaigns with ordered series, including daily and seasonal ones, each resolving through the same trigger keys the achievement engine uses. It also owns the safety quiz, the in-room word quiz, and one-off promotional badge claims.

### `pet`

Pet identity and species data, presence and autonomous behavior such as wandering, following, and staying put, a full training command set, riding, breeding, and monsterplants, plus the ongoing needs (food, drink, toys) that drive a pet's stats over time.

### `bot`

Room bots: their behavior loop, in-room chat, and the settings an owner configures for them.

### `chat`

Room chat delivery, global and per-room word filtering, flood control with configurable tiers, and the chat history log moderation reads from.

### `group`

Social groups: identity and badge composition, membership and roles, headquarters furniture and home-room decoration rights, and each group's own forum, including thread and post moderation and reporting a forum thread or message straight into moderation's intake.

### `messenger`

Friends, friend requests, and private messaging between players.

### `navigator`

Room search and browsing, organized into categories such as favorites, visit history, a player's own rooms, and hotel-wide listings, plus staff-picked rooms and room info.

### `marketplace`

The buy/sell furniture marketplace, including reserving an item out of a seller's inventory while it's listed.

### `trade`

Direct, session-based player-to-player trading: both sides build an offer, confirm, and the trade settles atomically or not at all.

### `subscription`

Club membership tiers (HC and VIP), scheduled paydays with streak bonuses and a percentage kickback on eligible catalog spending, monthly membership gifts, and store offers.

### `moderation`

Call-for-help intake from rooms, instant messages, forum threads, and reported photos, all funneled into one issue queue; staff tools to pick up, chat-review, and close issues; guide sessions; and guardian chat review.

### `sanction`

The centralized global punishment engine behind moderation: a single model for bans, mutes, warnings, trade locks, and kicks, issued by a moderator or the system itself, with an expiration and a reason attached to every one.

Alongside `internal/realm`, a few smaller packages provide shared machinery that realms build on rather than feature behavior of their own: `internal/auth` handles single sign-on ticket issuing and validation, `internal/permission` is the permission group and node registry every realm's admin routes check against, `internal/command` is the generic command dispatch pattern realms use for tick-scoped work, and `internal/tick` is the room scheduler primitive behind timed behavior such as rollers and Wired delays.

## Networking

`networking/codec` implements the wire format itself: frame and payload encoding shared by every packet. `networking/connection` is transport-agnostic session state, the handler registry, and security policy; it has no idea whether a given connection came in over a plain or encrypted channel by the time a packet handler sees it. `networking/inbound` and `networking/outbound` hold the actual packet definitions, one small package per packet, organized by realm and named after what the packet does rather than always matching its historical protocol name. A packet's Go package is independent of which realm handles it, since the mapping from header number to behavior lives in that realm's handler registration, not in the packet package itself.

## Reusable infrastructure

`pkg/` holds components with no game-specific behavior of their own: `postgres` and `redis` wrap their respective clients with the project's configuration and lifecycle conventions, `storage` wraps an S3-compatible object storage client, `http` hosts the Fiber-based admin API and OpenAPI documentation, `i18n` resolves localized player-facing text, `logger` and `build` provide structured logging and build metadata, and `bus` is the lightweight event bus realms publish domain events onto.

## Extension surface

`sdk/` is the controlled, stable surface for code that isn't part of the emulator core: `sdk/bot` and `sdk/plugin` expose the minimum a bot implementation or a plugin needs, without granting access to realm internals.
